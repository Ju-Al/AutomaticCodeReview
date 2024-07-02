package messages

	Orig          int
		Orig:          orig,
// DKGMessageIn is a wrapper around a DKGMessage containing the network ID of
// the sender.
type DKGMessageIn struct {
// DKGMessageOut is a wrapper around a DKGMessage containing the network ID of
type DKGMessageOut struct {
import "github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/crypto"
	"github.com/onflow/flow-go/model/flow"
)

// DKGMessage is the type of message exchanged between DKG nodes.
type DKGMessage struct {
	Orig          uint64
	Data          []byte
	DKGInstanceID string
}

// NewDKGMessage creates a new DKGMessage.
func NewDKGMessage(orig int, data []byte, dkgInstanceID string) DKGMessage {
	return DKGMessage{
		Orig:          uint64(orig),
		Data:          data,
		DKGInstanceID: dkgInstanceID,
	}
}

// PrivDKGMessageIn is a wrapper around a DKGMessage containing the network ID
// of the sender.
type PrivDKGMessageIn struct {
	DKGMessage
	OriginID flow.Identifier
}

// PrivDKGMessageOut is a wrapper around a DKGMessage containing the network ID of
// the destination.
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
