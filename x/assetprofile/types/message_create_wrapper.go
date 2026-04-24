package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateWrapper = "create_wrapper"

var _ sdk.Msg = &MsgCreateWrapper{}

// MsgCreateWrapper defines a message to create and register an ERC20 wrapper contract
type MsgCreateWrapper struct {
	Authority     string `protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	BaseDenom     string `protobuf:"bytes,2,opt,name=base_denom,json=baseDenom,proto3" json:"base_denom,omitempty"`
	DisplayName   string `protobuf:"bytes,3,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	DisplaySymbol string `protobuf:"bytes,4,opt,name=display_symbol,json=displaySymbol,proto3" json:"display_symbol,omitempty"`
	Decimals      uint64 `protobuf:"varint,5,opt,name=decimals,proto3" json:"decimals,omitempty"`
}

func NewMsgCreateWrapper(authority, baseDenom, displayName, displaySymbol string, decimals uint64) *MsgCreateWrapper {
	return &MsgCreateWrapper{
		Authority:     authority,
		BaseDenom:     baseDenom,
		DisplayName:   displayName,
		DisplaySymbol: displaySymbol,
		Decimals:      decimals,
	}
}

func (msg *MsgCreateWrapper) Route() string {
	return RouterKey
}

func (msg *MsgCreateWrapper) Type() string {
	return TypeMsgCreateWrapper
}

func (msg *MsgCreateWrapper) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

func (msg *MsgCreateWrapper) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateWrapper) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if msg.BaseDenom == "" {
		return errorsmod.Wrap(ErrInvalidBaseDenom, "base denom cannot be empty")
	}

	if msg.DisplayName == "" {
		return errorsmod.Wrap(ErrInvalidRequest, "display name cannot be empty")
	}

	if msg.DisplaySymbol == "" {
		return errorsmod.Wrap(ErrInvalidRequest, "display symbol cannot be empty")
	}

	if msg.Decimals < 6 || msg.Decimals > 18 {
		return ErrDecimalsInvalid
	}

	return nil
}

// MsgCreateWrapperResponse defines the response structure for executing a MsgCreateWrapper message.
type MsgCreateWrapperResponse struct {
	WrapperAddress string `protobuf:"bytes,1,opt,name=wrapper_address,json=wrapperAddress,proto3" json:"wrapper_address,omitempty"`
}
