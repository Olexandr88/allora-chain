package keeper

import (
	"context"
	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	"fmt"
	alloraMath "github.com/allora-network/allora-chain/math"
	"github.com/allora-network/allora-chain/x/emissions/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RESERVED_BLOCK = 0

// Boolean true if topic is active, else false
func (k *Keeper) GetNextPossibleChurningBlockByTopicId(ctx context.Context, topicId TopicId) (BlockHeight, bool, error) {
	currentBlock := sdk.UnwrapSDKContext(ctx).BlockHeight()
	block, err := k.topicToNextPossibleChurningBlock.Get(ctx, topicId)
	if err != nil {
		if errors.IsOf(err, collections.ErrNotFound) {
			return RESERVED_BLOCK, false, nil
		}
		return RESERVED_BLOCK, false, err
	}
	return block, block >= currentBlock, nil
}

// It is assumed the size of the outputted array has been bounded as it was constructed
// => can be safely handled in memory.
func (k *Keeper) GetActiveTopicIdsAtBlock(ctx context.Context, block BlockHeight) (types.TopicIds, error) {
	idsOfActiveTopics, err := k.blockToActiveTopics.Get(ctx, block)
	if err != nil {
		if errors.IsOf(err, collections.ErrNotFound) {
			topicIds := []TopicId{}
			return types.TopicIds{TopicIds: topicIds}, nil
		}
		return types.TopicIds{}, err
	}
	return idsOfActiveTopics, nil
}

// Boolean is true if the block is not found (true if no prior value), else false
func (k *Keeper) GetLowestActiveTopicWeightAtBlock(ctx context.Context, block BlockHeight) (types.TopicIdWeightPair, bool, error) {
	weight, err := k.blockToLowestActiveTopicWeight.Get(ctx, block)
	if err != nil {
		if errors.IsOf(err, collections.ErrNotFound) {
			return types.TopicIdWeightPair{}, true, nil
		}
		return types.TopicIdWeightPair{}, false, err
	}
	return weight, false, nil
}

// Removes data for a block if it exists in the maps:
// - blockToActiveTopics
// - blockToLowestActiveTopicWeight
// No op if the block does not exist in the maps
func (k *Keeper) PruneTopicActivationDataAtBlock(ctx context.Context, block BlockHeight) error {
	err := k.blockToActiveTopics.Remove(ctx, block)
	if err != nil {
		return err
	}

	err = k.blockToLowestActiveTopicWeight.Remove(ctx, block)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keeper) ResetLowestActiveTopicWeightAtBlock(ctx context.Context, block BlockHeight) error {
	activeTopicIds, err := k.GetActiveTopicIdsAtBlock(ctx, block)
	if err != nil {
		return err
	}

	if len(activeTopicIds.TopicIds) == 0 {
		return k.PruneTopicActivationDataAtBlock(ctx, block)
	}

	firstIter := true
	lowestWeight := alloraMath.NewDecFromInt64(0)
	idOfLowestWeightTopic := uint64(0)
	for _, topicId := range activeTopicIds.TopicIds {
		weight, err := k.GetTopicWeightFromTopicId(ctx, topicId)
		if err != nil {
			continue
		}
		if weight.Lt(lowestWeight) || firstIter {
			lowestWeight = weight
			idOfLowestWeightTopic = topicId
			firstIter = false
		}
	}

	data := types.TopicIdWeightPair{Weight: lowestWeight, TopicId: idOfLowestWeightTopic}
	return k.blockToLowestActiveTopicWeight.Set(ctx, block, data)
}

// Set a topic to inactive if the topic exists and is active, else does nothing
func (k *Keeper) inactivateTopicWithoutMinWeightReset(ctx context.Context, topicId TopicId) (BlockHeight, error) {
	topicExists, err := k.topics.Has(ctx, topicId)
	if err != nil {
		return RESERVED_BLOCK, err
	}
	if !topicExists {
		return RESERVED_BLOCK, nil
	}

	// Check if this topic is activated or not
	block, topicIsActive, err := k.GetNextPossibleChurningBlockByTopicId(ctx, topicId)
	if err != nil {
		return RESERVED_BLOCK, err
	}
	if !topicIsActive {
		return block, nil
	}

	topicIdsActiveAtBlock, err := k.GetActiveTopicIdsAtBlock(ctx, block)
	if err != nil {
		return RESERVED_BLOCK, err
	}
	// Remove the topic from the active topics at the block
	// If the topic is not found in the active topics at the block, no op
	newActiveTopicIds := []TopicId{}
	for i, id := range topicIdsActiveAtBlock.TopicIds {
		if id == topicId {
			newActiveTopicIds = append(topicIdsActiveAtBlock.TopicIds[:i], topicIdsActiveAtBlock.TopicIds[i+1:]...)
			break
		}
	}
	err = k.blockToActiveTopics.Set(ctx, block, types.TopicIds{TopicIds: newActiveTopicIds})
	if err != nil {
		return RESERVED_BLOCK, err
	}

	err = k.topicToNextPossibleChurningBlock.Remove(ctx, topicId)
	if err != nil {
		return RESERVED_BLOCK, err
	}

	// Set inactive for this topic
	err = k.activeTopics.Remove(ctx, topicId)
	if err != nil {
		return RESERVED_BLOCK, err
	}

	return block, nil
}

func (k *Keeper) addTopicToActiveSetRespectingLimitsWithoutMinWeightReset(
	ctx context.Context,
	topicId TopicId,
	block BlockHeight,
) (bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}

	topicIdsActiveAtBlock, err := k.GetActiveTopicIdsAtBlock(ctx, block)
	if err != nil {
		return false, err
	}
	existingActiveTopics := topicIdsActiveAtBlock.TopicIds

	// If the topic is already active at the block, no op
	for _, id := range existingActiveTopics {
		if id == topicId {
			return false, nil
		}
	}

	// If the number of active topics at the block is at the limit, remove the topic with the lowest weight
	if int32(len(existingActiveTopics)) >= int32(params.MaxActiveTopicsPerBlock) {
		// Remove the topic with the lowest weight
		lowestWeight, _, err := k.GetLowestActiveTopicWeightAtBlock(ctx, block)
		if err != nil {
			return false, err
		}

		weight, err := k.GetTopicWeightFromTopicId(ctx, topicId)
		if err != nil {
			return false, err
		}

		if weight.Lt(lowestWeight.Weight) {
			sdkCtx.Logger().Warn(fmt.Sprintf("Topic%d cannot be activated due to less than lowest weight at block %d", topicId, block))
			return false, nil
		}
		_, err = k.inactivateTopicWithoutMinWeightReset(ctx, lowestWeight.TopicId)
		if err != nil {
			return false, err
		}

		// Remove the lowest weight topic from the active topics at the block
		for i, id := range existingActiveTopics {
			if id == lowestWeight.TopicId {
				existingActiveTopics = append(existingActiveTopics[:i], existingActiveTopics[i+1:]...)
				break
			}
		}
	}

	existingActiveTopics = append(existingActiveTopics, topicId)
	// Add newly active topic to the active topics at the block
	newActiveTopicIds := types.TopicIds{TopicIds: existingActiveTopics}
	err = k.blockToActiveTopics.Set(ctx, block, newActiveTopicIds)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Set a topic to active if the topic exists, else does nothing
func (k *Keeper) ActivateTopic(ctx context.Context, topicId TopicId) error {
	topicExists, err := k.topics.Has(ctx, topicId)
	if err != nil {
		return err
	}
	if !topicExists {
		return nil
	}

	// Check topic activation with next possible churning block
	_, topicIsActive, err := k.GetNextPossibleChurningBlockByTopicId(ctx, topicId)
	if err != nil {
		return err
	}
	if topicIsActive {
		return nil
	}

	topic, err := k.GetTopic(ctx, topicId)
	if err != nil {
		return err
	}
	currentBlock := sdk.UnwrapSDKContext(ctx).BlockHeight()
	epochEndBlock := currentBlock + topic.EpochLength

	// Add this topic with epochend block
	// If the topic of the epochend block exceeds the limit, remove topic with lowest weight
	isAdded, err := k.addTopicToActiveSetRespectingLimitsWithoutMinWeightReset(ctx, topicId, epochEndBlock)
	if err != nil {
		return err
	}
	if !isAdded {
		return nil
	}

	err = k.topicToNextPossibleChurningBlock.Set(ctx, topicId, epochEndBlock)
	if err != nil {
		return err
	}

	// Update lowest topic weight of the block
	err = k.ResetLowestActiveTopicWeightAtBlock(ctx, epochEndBlock)
	if err != nil {
		return err
	}

	// Set active for this topic
	err = k.activeTopics.Set(ctx, topicId)
	if err != nil {
		return err
	}

	return nil
}

// Inactivate the topic
func (k *Keeper) InactivateTopic(ctx context.Context, topicId TopicId) error {
	_, err := k.inactivateTopicWithoutMinWeightReset(ctx, topicId)
	return err
}

// If the topic weight is not less than lowest weight keep it as activated
func (k *Keeper) UpdateActiveTopic(ctx context.Context, topicId TopicId) error {

	topic, err := k.GetTopic(ctx, topicId)
	if err != nil {
		return err
	}
	currentBlock := sdk.UnwrapSDKContext(ctx).BlockHeight()
	epochEndBlock := currentBlock + topic.EpochLength

	isAdded, err := k.addTopicToActiveSetRespectingLimitsWithoutMinWeightReset(ctx, topicId, epochEndBlock)
	if err != nil {
		return err
	}
	if !isAdded {
		return nil
	}

	err = k.topicToNextPossibleChurningBlock.Set(ctx, topicId, epochEndBlock)
	if err != nil {
		return err
	}

	err = k.ResetLowestActiveTopicWeightAtBlock(ctx, epochEndBlock)
	if err != nil {
		return err
	}
	return nil
}
