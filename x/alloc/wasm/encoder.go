package wasm

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sgwasm "github.com/public-awesome/stargaze/v14/internal/wasm"
	"github.com/public-awesome/stargaze/v14/x/alloc/types"
)

var _ sgwasm.Encoder = Encoder

type AllocMsg struct {
	FundFairburnPool *FundFairburnPool `json:"fund_fairburn_pool,omitempty"`
}

type FundFairburnPool struct {
	Amount []wasmvmtypes.Coin `json:"amount"`
}

func (fcp FundFairburnPool) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	amount, err := wasmkeeper.ConvertWasmCoinsToSdkCoins(fcp.Amount)
	if err != nil {
		return nil, err
	}
	msg := types.NewMsgFundFairburnPool(contract, amount)
	return []sdk.Msg{msg}, nil
}

func Encoder(contract sdk.AccAddress, data json.RawMessage, _ string) ([]sdk.Msg, error) {
	msg := &AllocMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	if msg.FundFairburnPool != nil {
		return msg.FundFairburnPool.Encode(contract)
	}
	return nil, fmt.Errorf("wasm: invalid custom alloc message")
}
