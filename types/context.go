package types

import (
	"context"
	"github.com/golang/protobuf/proto"

	wrsp "github.com/tepleton/wrsp/types"
)

/*

A note on Context security:

The intent of Context is for it to be an immutable object that can be cloned
and updated cheaply with WithValue() and passed forward to the next decorator
or handler. For example,

```golang
func Decorator(ctx Context, ms MultiStore, tx Tx, next Handler) Result {

	// Clone and update context with new kv pair.
	ctx2 := ctx.WithValueSDK(key, value)

	// Call the next decorator/handler.
	res := next(ctx2, ms, tx)

	// At this point, while `ctx` and `ctx2`'s shallow values haven't changed,
	// it's possible that slices or addressable struct fields have been
	// modified by the call to `next(...)`.
	//
	// This is generally undesirable because it prevents a decorator from
	// rolling back all side effects--which is the intent of immutable
	// `Context`s and store cache-wraps.
}
```

While well-written decorators wouldn't mutate any mutable context values, a malicious or buggy plugin can create unwanted side-effects, so it is highly advised for users of Context to only set immutable values.  To help enforce this contract, we require values to be certain primitive types, or a Cloner.

*/

type Cloner interface {
	Clone() interface{} // deep copy
}

type Context struct {
	context.Context
	// Don't add any other fields here,
	// it's probably not what you want to do.
}

func NewContext(header wrsp.Header, isCheckTx bool, txBytes []byte) Context {
	c := Context{context.Background()}
	c = c.setBlockHeader(header)
	c = c.setBlockHeight(header.Height)
	c = c.setChainID(header.ChainID)
	c = c.setIsCheckTx(isCheckTx)
	c = c.setTxBytes(txBytes)
	return c
}

func (c Context) Value(key interface{}) interface{} {
	value := c.Context.Value(key)
	if cloner, ok := value.(Cloner); ok {
		return cloner.Clone()
	}
	if message, ok := value.(proto.Message); ok {
		return proto.Clone(message)
	}
	return value
}

func (c Context) WithValue(key interface{}, value Cloner) Context {
	return c.withValue(key, value)
}

func (c Context) WithValueProto(key interface{}, value proto.Message) Context {
	return c.withValue(key, value)
}

func (c Context) WithValueString(key interface{}, value string) Context {
	return c.withValue(key, value)
}

func (c Context) WithValueInt32(key interface{}, value int32) Context {
	return c.withValue(key, value)
}

func (c Context) WithValueUint32(key interface{}, value uint32) Context {
	return c.withValue(key, value)
}

func (c Context) WithValueUint64(key interface{}, value uint64) Context {
	return c.withValue(key, value)
}

func (c Context) WithValueUnsafe(key interface{}, value interface{}) Context {
	return c.withValue(key, value)
}

func (c Context) withValue(key interface{}, value interface{}) Context {
	return Context{context.WithValue(c.Context, key, value)}
}

//----------------------------------------
// Our extensions

type contextKey int // local to the context module

const (
	contextKeyBlockHeader contextKey = iota
	contextKeyBlockHeight
	contextKeyChainID
	contextKeyIsCheckTx
	contextKeyTxBytes
)

func (c Context) BlockHeader() wrsp.Header {
	return c.Value(contextKeyBlockHeader).(wrsp.Header)
}

func (c Context) BlockHeight() int64 {
	return c.Value(contextKeyBlockHeight).(int64)
}

func (c Context) ChainID() string {
	return c.Value(contextKeyChainID).(string)
}

func (c Context) IsCheckTx() bool {
	return c.Value(contextKeyIsCheckTx).(bool)
}

func (c Context) TxBytes() []byte {
	return c.Value(contextKeyTxBytes).([]byte)
}

// Unexposed to prevent overriding.
func (c Context) setBlockHeader(header wrsp.Header) Context {
	var _ proto.Message = header // for cloning.
	return c.withValue(contextKeyBlockHeader, header)
}

// Unexposed to prevent overriding.
func (c Context) setBlockHeight(height int64) Context {
	return c.withValue(contextKeyBlockHeight, height)
}

// Unexposed to prevent overriding.
func (c Context) setChainID(chainID string) Context {
	return c.withValue(contextKeyChainID, chainID)
}

// Unexposed to prevent overriding.
func (c Context) setIsCheckTx(isCheckTx bool) Context {
	return c.withValue(contextKeyIsCheckTx, isCheckTx)
}

// Unexposed to prevent overriding.
func (c Context) setTxBytes(txBytes []byte) Context {
	return c.withValue(contextKeyTxBytes, txBytes)
}
