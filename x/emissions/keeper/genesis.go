package keeper

import (
	"context"

	cosmosMath "cosmossdk.io/math"
	state "github.com/upshot-tech/protocol-state-machine-module"
)

// InitGenesis initializes the module state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *state.GenesisState) error {
	if err := k.SetParams(ctx, data.Params); err != nil {
		return err
	}

	// ensure the module account exists
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, state.ModuleName)
	k.authKeeper.SetModuleAccount(ctx, moduleAccount)
	if err := k.SetLastRewardsUpdate(ctx, 0); err != nil {
		return err
	}
	if err := k.SetTotalStake(ctx, cosmosMath.NewUint(0)); err != nil {
		return err
	}
	// reserve topic ID 0 for future use
	if _, err := k.IncrementTopicId(ctx); err != nil {
		return err
	}

	return nil
}

// ExportGenesis exports the module state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) (*state.GenesisState, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &state.GenesisState{
		Params: params,
	}, nil
}
