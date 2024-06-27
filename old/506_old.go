package messages

import "github.com/onflow/flow-go/model/flow"
import (
	"github.com/onflow/flow-go/crypto"
	"github.com/onflow/flow-go/model/flow"
)

// DKGMessage is the type of message exchanged between DKG nodes.
type DKGMessage struct {
	Orig          int
	Orig          uint64
	Data          []byte
	DKGInstanceID string
}

// NewDKGMessage creates a new DKGMessage.
func NewDKGMessage(orig int, data []byte, dkgInstanceID string) DKGMessage {
	return DKGMessage{
		Orig:          orig,
		Orig:          uint64(orig),
		Data:          data,
		DKGInstanceID: dkgInstanceID,
	}
}

// DKGMessageIn is a wrapper around a DKGMessage containing the network ID of
// the sender.
type DKGMessageIn struct {
// PrivDKGMessageIn is a wrapper around a DKGMessage containing the network ID
// of the sender.
type PrivDKGMessageIn struct {
	DKGMessage
	OriginID flow.Identifier
}

// DKGMessageOut is a wrapper around a DKGMessage containing the network ID of
// PrivDKGMessageOut is a wrapper around a DKGMessage containing the network ID of
// the destination.
type DKGMessageOut struct {
type PrivDKGMessageOut struct {
	DKGMessage
	DestID flow.Identifier
}

// BcastDKGMessage is a wrapper around a DKGMessage intended for broadcasting.
// It contains a signature of the DKGMessage signed with the staking key of the
// sender.
type BcastDKGMessage struct {
	DKGMessage
	Signature crypto.Signature
}
