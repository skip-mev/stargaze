package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/public-awesome/stargaze/v14/testutil/keeper"
	"github.com/public-awesome/stargaze/v14/testutil/sample"
	"github.com/public-awesome/stargaze/v14/x/globalfee/keeper"
	"github.com/public-awesome/stargaze/v14/x/globalfee/types"
	"github.com/stretchr/testify/require"
)

func TestSetCodeAuthorization(t *testing.T) {
	testCases := []struct {
		testCase    string
		prepare     func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetCodeAuthorization
		expectError bool
	}{
		{
			"invalid sender address",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetCodeAuthorization {
				msg := types.MsgSetCodeAuthorization{
					Sender: "👻",
					CodeAuthorization: &types.CodeAuthorization{
						CodeID:  2,
						Methods: []string{"2"},
					},
				}
				return &msg
			},
			true,
		},
		{
			"sender not privileged",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetCodeAuthorization {
				sender := sample.AccAddress()
				msg := types.MsgSetCodeAuthorization{
					Sender: sender.String(),
					CodeAuthorization: &types.CodeAuthorization{
						CodeID:  2,
						Methods: []string{"2"},
					},
				}
				return &msg
			},
			true,
		},
		{
			"invalid methods",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetCodeAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg := types.MsgSetCodeAuthorization{
					Sender: sender.String(),
					CodeAuthorization: &types.CodeAuthorization{
						CodeID: 2,
					},
				}
				return &msg
			},
			true,
		},
		{
			"valid",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetCodeAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)
				msg := types.MsgSetCodeAuthorization{
					Sender: sender.String(),
					CodeAuthorization: &types.CodeAuthorization{
						CodeID:  2,
						Methods: []string{"2"},
					},
				}
				return &msg
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			k, c := keepertest.GlobalFeeKeeper(t)
			msgSrvr, ctx := keeper.NewMsgServerImpl(k), c

			msg := tc.prepare(c, k)

			_, err := msgSrvr.SetCodeAuthorization(ctx, msg)

			if tc.expectError {
				require.Error(t, err, tc)
			} else {
				require.NoError(t, err, tc)
				found := k.HasCodeAuthorization(ctx, msg.CodeAuthorization.CodeID)
				require.True(t, found)
			}
		})
	}
}

func TestRemoveCodeAuthorization(t *testing.T) {
	testCases := []struct {
		testCase    string
		prepare     func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgRemoveCodeAuthorization
		expectError bool
	}{
		{
			"invalid sender address",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgRemoveCodeAuthorization {
				msg := types.MsgRemoveCodeAuthorization{
					Sender: "👻",
					CodeID: 2,
				}
				return &msg
			},
			true,
		},
		{
			"sender not privileged",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgRemoveCodeAuthorization {
				sender := sample.AccAddress()
				msg := types.MsgRemoveCodeAuthorization{
					Sender: sender.String(),
					CodeID: 2,
				}
				return &msg
			},
			true,
		},
		{
			"valid",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgRemoveCodeAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg := types.MsgRemoveCodeAuthorization{
					Sender: sender.String(),
					CodeID: 2,
				}
				return &msg
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			k, c := keepertest.GlobalFeeKeeper(t)
			msgSrvr, ctx := keeper.NewMsgServerImpl(k), c
			err := k.SetCodeAuthorization(c, types.CodeAuthorization{
				CodeID:  2,
				Methods: []string{"mint"},
			})
			require.NoError(t, err)

			msg := tc.prepare(c, k)

			_, err = msgSrvr.RemoveCodeAuthorization(ctx, msg)

			if tc.expectError {
				require.Error(t, err, tc)
				found := k.HasCodeAuthorization(ctx, msg.GetCodeID())
				require.True(t, found)
			} else {
				require.NoError(t, err, tc)
				found := k.HasCodeAuthorization(ctx, msg.GetCodeID())
				require.False(t, found)
			}
		})
	}
}

func TestSetContractAuthorization(t *testing.T) {
	testCases := []struct {
		testCase    string
		prepare     func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetContractAuthorization
		expectError bool
	}{
		{
			"invalid sender address",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetContractAuthorization {
				msg := types.MsgSetContractAuthorization{
					Sender: "👻",
					ContractAuthorization: &types.ContractAuthorization{
						ContractAddress: sample.AccAddress().String(),
						Methods:         []string{"2"},
					},
				}
				return &msg
			},
			true,
		},
		{
			"sender not privileged",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetContractAuthorization {
				sender := sample.AccAddress()
				msg := types.MsgSetContractAuthorization{
					Sender: sender.String(),
					ContractAuthorization: &types.ContractAuthorization{
						ContractAddress: sample.AccAddress().String(),
						Methods:         []string{"2"},
					},
				}
				return &msg
			},
			true,
		},
		{
			"invalid contract address",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetContractAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg := types.MsgSetContractAuthorization{
					Sender: sender.String(),
					ContractAuthorization: &types.ContractAuthorization{
						ContractAddress: "👻",
						Methods:         []string{"2"},
					},
				}
				return &msg
			},
			true,
		},
		{
			"contract doesn't exist",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetContractAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg := types.MsgSetContractAuthorization{
					Sender: sender.String(),
					ContractAuthorization: &types.ContractAuthorization{
						ContractAddress: sample.AccAddress().String(),
						Methods:         []string{"2"},
					},
				}
				return &msg
			},
			true,
		},
		{
			"invalid methods",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetContractAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg := types.MsgSetContractAuthorization{
					Sender: sender.String(),
					ContractAuthorization: &types.ContractAuthorization{
						ContractAddress: sample.AccAddress().String(),
					},
				}
				return &msg
			},
			true,
		},
		{
			"valid",
			func(ctx sdk.Context, keeper keeper.Keeper) *types.MsgSetContractAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg := types.MsgSetContractAuthorization{
					Sender: sender.String(),
					ContractAuthorization: &types.ContractAuthorization{
						ContractAddress: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du",
						Methods:         []string{"2"},
					},
				}
				return &msg
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			k, c := keepertest.GlobalFeeKeeper(t)
			msgSrvr, ctx := keeper.NewMsgServerImpl(k), c

			msg := tc.prepare(c, k)

			_, err := msgSrvr.SetContractAuthorization(ctx, msg)

			if tc.expectError {
				require.Error(t, err, tc)
			} else {
				require.NoError(t, err, tc)
				found := k.HasContractAuthorization(c, sdk.MustAccAddressFromBech32(msg.GetContractAuthorization().GetContractAddress()))
				require.True(t, found)
			}
		})
	}
}

func TestRemoveContractAuthorization(t *testing.T) {
	testCases := []struct {
		testCase    string
		prepare     func(ctx sdk.Context, keeper keeper.Keeper, contractAddress string) *types.MsgRemoveContractAuthorization
		expectError bool
	}{
		{
			"invalid sender address",
			func(ctx sdk.Context, keeper keeper.Keeper, contractAddress string) *types.MsgRemoveContractAuthorization {
				msg := types.MsgRemoveContractAuthorization{
					Sender:          "👻",
					ContractAddress: contractAddress,
				}
				return &msg
			},
			true,
		},
		{
			"sender not privileged",
			func(ctx sdk.Context, keeper keeper.Keeper, contractAddress string) *types.MsgRemoveContractAuthorization {
				sender := sample.AccAddress()
				msg := types.MsgRemoveContractAuthorization{
					Sender:          sender.String(),
					ContractAddress: contractAddress,
				}
				return &msg
			},
			true,
		},
		{
			"invalid contract address",
			func(ctx sdk.Context, keeper keeper.Keeper, contractAddress string) *types.MsgRemoveContractAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg := types.MsgRemoveContractAuthorization{
					Sender:          sender.String(),
					ContractAddress: "👻",
				}
				return &msg
			},
			true,
		},
		{
			"valid",
			func(ctx sdk.Context, keeper keeper.Keeper, contractAddress string) *types.MsgRemoveContractAuthorization {
				sender := sample.AccAddress()
				params := types.NewParams([]string{sender.String()})
				err := keeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg := types.MsgRemoveContractAuthorization{
					Sender:          sender.String(),
					ContractAddress: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du",
				}
				return &msg
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			k, c := keepertest.GlobalFeeKeeper(t)
			msgSrvr, ctx := keeper.NewMsgServerImpl(k), c
			contractAddr := sdk.MustAccAddressFromBech32("cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du")
			err := k.SetContractAuthorization(c, types.ContractAuthorization{
				ContractAddress: contractAddr.String(),
				Methods:         []string{"mint"},
			})
			require.NoError(t, err)

			msg := tc.prepare(c, k, contractAddr.String())

			_, err = msgSrvr.RemoveContractAuthorization(ctx, msg)

			if tc.expectError {
				require.Error(t, err, tc)
				found := k.HasContractAuthorization(c, contractAddr)
				require.True(t, found)
			} else {
				require.NoError(t, err, tc)
				found := k.HasContractAuthorization(c, contractAddr)
				require.False(t, found)
			}
		})
	}
}
