package wasm

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const (
	DistributionRoute = "distribution"
)

var _ Encoder = CustomDistributionEncoder

type FundCommunityPool struct {
	Amount []wasmvmtypes.Coin `json:"amount"`
}

func (fcp FundCommunityPool) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	amount, err := wasmkeeper.ConvertWasmCoinsToSdkCoins(fcp.Amount)
	if err != nil {
		return nil, err
	}
	msg := distributiontypes.NewMsgFundCommunityPool(amount, contract.String())
	return []sdk.Msg{msg}, nil
}

type DistributionMsg struct {
	FundCommunityPool *FundCommunityPool `json:"fund_community_pool,omitempty"`
}

func CustomDistributionEncoder(contract sdk.AccAddress, data json.RawMessage, _ string) ([]sdk.Msg, error) {
	msg := &DistributionMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	if msg.FundCommunityPool != nil {
		return msg.FundCommunityPool.Encode(contract)
	}
	return nil, fmt.Errorf("wasm: invalid custom distribution message")
}
