package types

import (
	cosmosMath "cosmossdk.io/math"
	alloraMath "github.com/allora-network/allora-chain/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() *GenesisState {
	return &GenesisState{
		Params:                                         DefaultParams(),
		NextTopicId:                                    0,
		Topics:                                         []*TopicIdAndTopic{},
		ActiveTopics:                                   []uint64{},
		RewardableTopics:                               []uint64{},
		TopicWorkers:                                   []*TopicAndActorId{},
		TopicReputers:                                  []*TopicAndActorId{},
		TopicRewardNonce:                               []*TopicIdAndBlockHeight{},
		InfererScoresByBlock:                           []*TopicIdBlockHeightScores{},
		ForecasterScoresByBlock:                        []*TopicIdBlockHeightScores{},
		ReputerScoresByBlock:                           []*TopicIdBlockHeightScores{},
		LatestInfererScoresByWorker:                    []*TopicIdActorIdScore{},
		LatestForecasterScoresByWorker:                 []*TopicIdActorIdScore{},
		LatestReputerScoresByReputer:                   []*TopicIdActorIdScore{},
		ReputerListeningCoefficient:                    []*TopicIdActorIdListeningCoefficient{},
		PreviousReputerRewardFraction:                  []*TopicIdActorIdDec{},
		PreviousInferenceRewardFraction:                []*TopicIdActorIdDec{},
		PreviousForecastRewardFraction:                 []*TopicIdActorIdDec{},
		PreviousForecasterScoreRatio:                   []*TopicIdAndDec{},
		TotalStake:                                     cosmosMath.ZeroInt(),
		TopicStake:                                     []*TopicIdAndInt{},
		StakeReputerAuthority:                          []*TopicIdActorIdInt{},
		StakeSumFromDelegator:                          []*TopicIdActorIdInt{},
		DelegatedStakes:                                []*TopicIdDelegatorReputerDelegatorInfo{},
		StakeFromDelegatorsUponReputer:                 []*TopicIdActorIdInt{},
		DelegateRewardPerShare:                         []*TopicIdActorIdDec{},
		StakeRemovalsByBlock:                           []*BlockHeightTopicIdReputerStakeRemovalInfo{},
		StakeRemovalsByActor:                           []*ActorIdTopicIdBlockHeight{},
		DelegateStakeRemovalsByBlock:                   []*BlockHeightTopicIdDelegatorReputerDelegateStakeRemovalInfo{},
		DelegateStakeRemovalsByActor:                   []*DelegatorReputerTopicIdBlockHeight{},
		Inferences:                                     []*TopicIdActorIdInference{},
		Forecasts:                                      []*TopicIdActorIdForecast{},
		Workers:                                        []*LibP2PKeyAndOffchainNode{},
		Reputers:                                       []*LibP2PKeyAndOffchainNode{},
		TopicFeeRevenue:                                []*TopicIdAndInt{},
		PreviousTopicWeight:                            []*TopicIdAndDec{},
		AllInferences:                                  []*TopicIdBlockHeightInferences{},
		AllForecasts:                                   []*TopicIdBlockHeightForecasts{},
		AllLossBundles:                                 []*TopicIdBlockHeightReputerValueBundles{},
		NetworkLossBundles:                             []*TopicIdBlockHeightValueBundles{},
		PreviousPercentageRewardToStakedReputers:       alloraMath.ZeroDec(),
		UnfulfilledWorkerNonces:                        []*TopicIdAndNonces{},
		UnfulfilledReputerNonces:                       []*TopicIdAndReputerRequestNonces{},
		LatestInfererNetworkRegrets:                    []*TopicIdActorIdTimeStampedValue{},
		LatestForecasterNetworkRegrets:                 []*TopicIdActorIdTimeStampedValue{},
		LatestOneInForecasterNetworkRegrets:            []*TopicIdActorIdActorIdTimeStampedValue{},
		LatestNaiveInfererNetworkRegrets:               []*TopicIdActorIdTimeStampedValue{},
		LatestOneOutInfererInfererNetworkRegrets:       []*TopicIdActorIdActorIdTimeStampedValue{},
		LatestOneOutInfererForecasterNetworkRegrets:    []*TopicIdActorIdActorIdTimeStampedValue{},
		LatestOneOutForecasterInfererNetworkRegrets:    []*TopicIdActorIdActorIdTimeStampedValue{},
		LatestOneOutForecasterForecasterNetworkRegrets: []*TopicIdActorIdActorIdTimeStampedValue{},
		CoreTeamAddresses:                              DefaultCoreTeamAddresses(),
		TopicLastWorkerCommit:                          []*TopicIdTimestampedActorNonce{},
		TopicLastReputerCommit:                         []*TopicIdTimestampedActorNonce{},
	}
}

func DefaultCoreTeamAddresses() []string {
	return []string{
		"allo16270t36amc3y6wk2wqupg6gvg26x6dc2nr5xwl",
		"allo1xm0jg40dcvccqvzqwv5skxlpc7t6eku69kfz6y",
		"allo1g4y6ra95z2zewupm7p45z4ny00rs7m24rj5hn8",
		"allo10w0jcq50ufsuy9332dkz6zf4gu00xm9zhfyn3s",
		"allo1lvymnmzndmam00uvxq8hr63jq8jfrups4ymlg2",
		"allo1d7vr2dxahkcz0snk28pets9uqvyxjdlysst3z3",
		"allo19gtttc7qg50n3hjn0qxdasdudf260cx7vevk8j",
		"allo1jc2mme2fj458kg08v2z92m8f9vsqwfzt0ju9ys",
		"allo1uff55lgqpjkw2mlsx2q0p8q8z7k7p00w9s4s0f",
		"allo136eeqhawxx66sjgsfeqk9gewq0e0msyu5tjmj3",
	}
}

// Validate performs basic genesis state validation returning an error upon any
func (gs *GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Ensure that the core team addresses are valid
	for _, addr := range gs.CoreTeamAddresses {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return err
		}
	}

	return nil
}
