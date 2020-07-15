package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/public-awesome/stakebird/x/curating/types"
)

func (k Keeper) GetRewardPool(ctx sdk.Context) (rewardPool authexported.ModuleAccountI) {
	return k.accountKeeper.GetModuleAccount(ctx, types.RewardPoolName)
}

func (k Keeper) InflateRewardPool(ctx sdk.Context) error {
	blockInflationAddr := k.accountKeeper.GetModuleAccount(ctx, auth.FeeCollectorName).GetAddress()
	blockInflation := k.bankKeeper.GetBalance(ctx, blockInflationAddr, types.DefaultStakeDenom)
	rewardPoolAllocation := k.GetParams(ctx).RewardPoolAllocation

	blockInflationDec := sdk.NewDecFromInt(blockInflation.Amount)
	rewardAmount := blockInflationDec.Mul(rewardPoolAllocation)
	rewardCoin := sdk.NewCoin(types.DefaultStakeDenom, rewardAmount.TruncateInt())

	return k.bankKeeper.SendCoinsFromModuleToModule(
		ctx, auth.FeeCollectorName, types.RewardPoolName, sdk.NewCoins(rewardCoin))
}
