package events

import (
	apigen "github.com/go-libp2p-chat/go-node/gen/api"
	"github.com/go-libp2p-chat/go-node/message"
	"github.com/libp2p/go-libp2p-core/peer"
)

// Event represents a node event.
type Event interface {
	// MarshalToProtobuf maps events into API protobuf events.
	MarshalToProtobuf() *apigen.Event
}

// NewMessage occurs when a new message is received from a subscribed room (pubsub topic).
type NewMessage struct {
	Message  message.Message
	RoomName string
}

func (e *NewMessage) MarshalToProtobuf() *apigen.Event {
	return &apigen.Event{
		Type: apigen.Event_NEW_CHAT_MESSAGE,
		NewChatMessage: &apigen.EvtNewChatMessage{
			ChatMessage: &apigen.ChatMessage{
				SenderId:  e.Message.SenderID.Pretty(),
				Timestamp: e.Message.Timestamp.Unix(),
				Value:     e.Message.Value,
			},
			RoomName: e.RoomName,
		},
	}
}

/*

// NewMessage occurs when a new message is received from a subscribed room (pubsub topic).
type NewRateLimitMessage struct {
	Message  message.Message
	RoomName string
}

func (e *NewRateLimitMessage) MarshalToProtobuf() *apigen.Event {
	return &apigen.Event{
		Type: apigen.Event_NEW_RATE_LIMIT_MESSAGE,
		NewRateLimitMessage: &apigen.EvtNewRateLimitMessage{
			RateLimitMessage: &apigen.RateLimitMessage{
				SenderId:  e.Message.SenderID.Pretty(),
				Timestamp: e.Message.Timestamp.Unix(),
				Value:     e.Message.Value,
			},
			RoomName: e.RoomName,
		},
	}
}

// NewMessage occurs when a new message is received from a subscribed room (pubsub topic).
type NewBlockMessage struct {
	Message  message.Message
	RoomName string
}

func (e *NewBlockMessage) MarshalToProtobuf() *apigen.Event {
	return &apigen.Event{
		Type: apigen.Event_NEW_BLOCK_MESSAGE,
		NewBlockMessage: &apigen.EvtNewBlockMessage{
			BlockMessage: &apigen.BlockMessage{
				SenderId:  e.Message.SenderID.Pretty(),
				Timestamp: e.Message.Timestamp.Unix(),
				Value:     e.Message.Value,
			},
			RoomName: e.RoomName,
		},
	}
}

// NewMessage occurs when a new message is received from a subscribed room (pubsub topic).
type NewBanMessage struct {
	Message  message.Message
	RoomName string
}

func (e *NewBanMessage) MarshalToProtobuf() *apigen.Event {
	return &apigen.Event{
		Type: apigen.Event_NEW_BAN_MESSAGE,
		NewBanMessage: &apigen.EvtNewBanMessage{
			BanMessage: &apigen.BanMessage{
				SenderId:  e.Message.SenderID.Pretty(),
				Timestamp: e.Message.Timestamp.Unix(),
				Value:     e.Message.Value,
			},
			RoomName: e.RoomName,
		},
	}
}

// NewMessage occurs when a new message is received from a subscribed room (pubsub topic).
type NewModerateMessage struct {
	Message  message.Message
	RoomName string
}

func (e *NewModerateMessage) MarshalToProtobuf() *apigen.Event {
	return &apigen.Event{
		Type: apigen.Event_NEW_MODERATOR_MESSAGE,
		NewModeratorMessage: &apigen.EvtNewModeratorMessage{
			ModeratorMessage: &apigen.ModeratorMessage{
				SenderId:  e.Message.SenderID.Pretty(),
				Timestamp: e.Message.Timestamp.Unix(),
				Value:     e.Message.Value,
			},
			RoomName: e.RoomName,
		},
	}
}
*/

// PeerJoined occurs when a peer joins a room that we are subscribed to.
type PeerJoined struct {
	PeerID   peer.ID
	RoomName string
}

func (e *PeerJoined) MarshalToProtobuf() *apigen.Event {
	return &apigen.Event{
		Type: apigen.Event_PEER_JOINED,
		PeerJoined: &apigen.EvtPeerJoined{
			RoomName: e.RoomName,
			PeerId:   e.PeerID.Pretty(),
		},
	}
}

// PeerLeft occurs when a peer lefts a room that we are subscribed to.
type PeerLeft struct {
	PeerID   peer.ID
	RoomName string
}

func (e *PeerLeft) MarshalToProtobuf() *apigen.Event {
	return &apigen.Event{
		Type: apigen.Event_PEER_LEFT,
		PeerLeft: &apigen.EvtPeerLeft{
			RoomName: e.RoomName,
			PeerId:   e.PeerID.Pretty(),
		},
	}
}
