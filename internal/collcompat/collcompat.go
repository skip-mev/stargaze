package collcompat

import (
	collcodec "cosmossdk.io/collections/codec"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gogo/protobuf/proto"
)

type protoMessage[T any] interface {
	*T
	proto.Message
}

// ProtoValue inits a collections.ValueCodec for a generic gogo protobuf message.
func ProtoValue[T any, PT protoMessage[T]](cdc codec.BinaryCodec) collcodec.ValueCodec[T] {
	return &collValue[T, PT]{cdc.(codec.Codec), proto.MessageName(PT(new(T)))}
}

type collValue[T any, PT protoMessage[T]] struct {
	cdc         codec.Codec
	messageName string
}

func (c collValue[T, PT]) Encode(value T) ([]byte, error) {
	return c.cdc.Marshal(PT(&value))
}

func (c collValue[T, PT]) Decode(b []byte) (value T, err error) {
	err = c.cdc.Unmarshal(b, PT(&value))
	return value, err
}

func (c collValue[T, PT]) EncodeJSON(value T) ([]byte, error) {
	return c.cdc.MarshalJSON(PT(&value))
}

func (c collValue[T, PT]) DecodeJSON(b []byte) (value T, err error) {
	err = c.cdc.UnmarshalJSON(b, PT(&value))
	return
}

func (c collValue[T, PT]) Stringify(value T) string {
	return PT(&value).String()
}

func (c collValue[T, PT]) ValueType() string {
	return "github.com/cosmos/gogoproto/" + c.messageName
}
