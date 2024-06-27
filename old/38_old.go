package sign

import (
	"reflect"

	"github.com/dedis/crypto/abstract"
	"github.com/dedis/crypto/edwards"
//"github.com/dedis/crypto/nist"
	"github.com/dedis/cothority/lib/hashid"
	"github.com/dedis/cothority/lib/proof"
	"github.com/dedis/protobuf"
	"encoding/json"
)

// All message structures defined in this package are used in the
// Collective Signing Protocol
// Over the network they are sent as byte slices, so each message
// has its own MarshlBinary and UnmarshalBinary method

type MessageType int

const (
	Unset MessageType = iota
	Announcement
	Commitment
	Challenge
	Response
	SignatureBroadcast
	CatchUpReq
	CatchUpResp
	GroupChange
	GroupChanged
	StatusConnections
	CloseAll
	Default // for internal use
	Error
)

func (m MessageType) String() string {
	switch m {
	case Unset:
		return "Unset"
	case Announcement:
		return "Announcement"
	case Commitment:
		return "Commitment"
	case Challenge:
		return "Challenge"
	case Response:
		return "Response"
	case SignatureBroadcast:
		return "SignatureBroadcast"
	case CatchUpReq:
		return "CatchUpRequest"
	case CatchUpResp:
		return "CatchUpResponse"
	case GroupChange:
		return "GroupChange"
	case GroupChanged:
		return "GroupChanged"
	case StatusConnections:
		return "StatusConnections"
	case CloseAll:
		return "CloseAll"
	case Default: // for internal use
		return "Default"
	case Error:
		return "Error"
	}
	return "INVALID TYPE"
}

// Signing Messages are used for all communications between servers
// It is important for encoding/ decoding for type to be kept as first field
type SigningMessage struct {
	Type         MessageType
	Am           *AnnouncementMessage
	Com          *CommitmentMessage
	Chm          *ChallengeMessage
	Rm           *ResponseMessage
	SBm          *SignatureBroadcastMessage
	Cureq        *CatchUpRequest
	Curesp       *CatchUpResponse
	Vrm          *VoteRequestMessage
	Gcm          *GroupChangedMessage
	Err          *ErrorMessage
	From         string
	View         int
	LastSeenVote int // highest vote ever seen and commited in log, used for catch-up
}

var msgSuite abstract.Suite = edwards.NewAES128SHA256Ed25519(true)

//var msgSuite abstract.Suite = nist.NewAES128SHA256P256()

func NewSigningMessage() interface{} {
	return &SigningMessage{}
}

func (sm *SigningMessage) MarshalBinary() ([]byte, error) {
	return protobuf.Encode(sm)
}

func (sm *SigningMessage) UnmarshalBinary(data []byte) error {
	var cons = make(protobuf.Constructors)
	var point abstract.Point
	var secret abstract.Secret
	cons[reflect.TypeOf(&point).Elem()] = func() interface{} { return msgSuite.Point() }
	cons[reflect.TypeOf(&secret).Elem()] = func() interface{} { return msgSuite.Secret() }
	return protobuf.DecodeWithConstructors(data, sm, cons)
}

type JSONdata struct{
		Len int64
		Data []byte
}

func (sm *SigningMessage) MarshalJSON() ([]byte, error) {
	data, _ := sm.MarshalBinary()
	return json.Marshal(JSONdata{
		Len: int64(len(data)),
		Data: data,
	})
}

func (sm *SigningMessage) UnmarshalJSON(dataJSON []byte) error {
	jdata := JSONdata{}
	json.Unmarshal(dataJSON, &jdata)
	return sm.UnmarshalBinary(jdata.Data)
}

// Broadcasted message initiated and signed by proposer
type AnnouncementMessage struct {
	Message []byte
	Round   int
				  // VoteRequest *VoteRequest
	Vote    *Vote // Vote Request (propose)
}

type CommitmentMessage struct {
	V             abstract.Point // commitment Point
	V_hat         abstract.Point // product of subtree participating nodes' commitment points
	X_hat         abstract.Point // product of subtree participating nodes' public keys

	MTRoot        hashid.HashId  // root of Merkle (sub)Tree

								 // public keys of children servers that did not respond to
								 // annoucement from root
	ExceptionList []abstract.Point

								 // CountedVotes *CountedVotes // CountedVotes contains a subtree's votes
	Vote          *Vote          // Vote Response (promise)

	Round         int
}

type ChallengeMessage struct {
	C      abstract.Secret // challenge

						   // Depth  byte
	MTRoot hashid.HashId   // the very root of the big Merkle Tree
	Proof  proof.Proof     // Merkle Path of Proofs from root to us

						   // CountedVotes *CountedVotes //  CountedVotes contains the whole tree's votes
	Vote   *Vote           // Vote Confirmerd/ Rejected (accept)

	Round  int
}

type ResponseMessage struct {
	R_hat          abstract.Secret // response

								   // public keys of children servers that did not respond to
								   // challenge from root
	ExceptionList  []abstract.Point
								   // cummulative point commits of nodes that failed after commit
	ExceptionV_hat abstract.Point
								   // cummulative public keys of nodes that failed after commit
	ExceptionX_hat abstract.Point

	Vote           *Vote           // Vote Ack/Nack in thr log (ack/nack)

	Round          int
}

// 5th message going from root to leaves to send the
// signature
type SignatureBroadcastMessage struct {
	// Aggregate response of root
	R0_hat abstract.Secret
	// Challenge
	C      abstract.Secret
	// Aggregate public key
	X0_hat abstract.Point
	// Aggregate public commitment
	V0_hat abstract.Point
}

type ErrorMessage struct {
	Err string
}

type VoteRequestMessage struct {
	Vote *Vote
}

type GroupChangedMessage struct {
	V        *Vote
	// if vote not accepted rest of fields are nil
	HostList []string
}
