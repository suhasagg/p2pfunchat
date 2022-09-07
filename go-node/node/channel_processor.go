package node

import (
	"context"
	"encoding/json"
	"fmt"
	block "github.com/go-libp2p-chat/go-node/blocklist"
	"github.com/go-libp2p-chat/go-node/events"
	"github.com/go-libp2p-chat/go-node/message"
	"github.com/go-libp2p-chat/go-node/storage/boltdbstorage"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	ratelimit "github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/memorystore"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	roomParticipantsTTLPermanent = math.MaxInt64
	roomParticipantsTTL          = time.Second * 300
//If user block ttl has expired, user will be unblocked
	roomBlockTTL                 = time.Minute * 1
//if user ban ttl has expired, user will be unbanned
	roomBanTTL                   = time.Hour * 168
)

type participantsEntry struct {
	ID       peer.ID `json:"id"`
	Nickname string  `json:"nickname"`

	ttl     time.Duration
	addedAt time.Time
}

//user moderation data stored with peer id, it contains timestamps when user was blocked/banned
type moderationData struct {
	nickname  string
	counter   int
	blocked   bool
	banned    bool
	blockedAt time.Time
	bannedAt  time.Time
}

// RoomMessageType enumerates the possible types of pubsub room messages.
type RoomMessageType string

const (
	// RoomMessageTypeChatMessage is published when a new chat message is sent from the node.
	RoomMessageTypeChatMessage RoomMessageType = "room.message"

	// RoomMessageTypeAdvertise is published to indicate a node is still connected to a room.
	RoomMessageTypeAdvertise RoomMessageType = "room.advertise"

	// RoomMessageTypeModerateMessage is published by moderator (a bot trained on user patterns) for content moderation.
	RoomMessageTypeModerateMessage RoomMessageType = "room.moderate"

	// RoomMessageTypeBlockMessage is published by moderator (a bot trained on user patterns) for content moderation.
	RoomMessageTypeBlockMessage RoomMessageType = "room.block"

	// RoomMessageTypeBanMessage is published by moderator (a bot trained on user patterns) for content moderation.
	RoomMessageTypeBanMessage RoomMessageType = "room.ban"

	// RoomMessageTypeBanMessage is published by moderator (a bot trained on user patterns) for content moderation.
	RoomMessageTypeRateLimitMessage RoomMessageType = "room.ratelimit"
)

// RoomMessageOut holds data to be published in a topic.
type RoomMessageOut struct {
	Type    RoomMessageType `json:"type"`
	Payload interface{}     `json:"payload,omitempty"`
}

// RoomMessageIn holds data to be received from a topic.
//
// The Payload field is lazily unmarshalled because it depends on the type of message published.
type RoomMessageIn struct {
	Type    RoomMessageType `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Channel holds room event and pubsub data.
type Channel struct {
	name         string
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	// Map is counter of moderation messages for user in a chat room
	usermoderationCounter sync.Map
	counter               int
	//moderationdata        map[peer.ID]*moderationData
	moderationdata sync.Map
	//user to limit cooling period feature of chat room via token bucket algorithm 
	ratelimitstore ratelimit.Store
	lock           sync.RWMutex
	participants   map[peer.ID]*participantsEntry
}

func setupDB() (*boltdbstorage.Storage, error) {
	logger, err := zap.NewDevelopment()
	db, err := bolt.Open("4.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}
	storage, err := boltdbstorage.New(db, []byte("topic"))
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	logger.Debug("BoltDB Setup Done")
	return storage, nil
}

func newChannel(name string, topic *pubsub.Topic, subscription *pubsub.Subscription) *Channel {
	var moderationcounter int
	return &Channel{
		name:  name,
		topic: topic,
		//moderationdata: make(map[peer.ID]*moderationData),
		subscription: subscription,
		counter:      moderationcounter,
		participants: make(map[peer.ID]*participantsEntry),
	}
}

func (r *Channel) addParticipant(peerID peer.ID, nickname string, ttl time.Duration) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	_, exists := r.participants[peerID]
	r.participants[peerID] = &participantsEntry{
		ID:       peerID,
		Nickname: nickname,

		ttl:     ttl,
		addedAt: time.Now(),
	}
	return exists
}

func (r *Channel) removeParticipant(peerID peer.ID) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, exists := r.participants[peerID]; !exists {
		return false
	}

	delete(r.participants, peerID)
	return true
}

func (r *Channel) getParticipants() []participantsEntry {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var participants []participantsEntry
	for _, p := range r.participants {
		participants = append(participants, *p)
	}

	return participants
}

func (r *Channel) refreshParticipants(onRemove func(peer.ID)) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for peerID, participant := range r.participants {
		if time.Now().Sub(participant.addedAt) <= participant.ttl {
			continue
		}

		// participant ttl expired
		delete(r.participants, peerID)
		onRemove(peerID)
	}
}

//After a time period, this method is called periodically to unblock/unban user  
func (r *Channel) unblockunbanParticipants() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.moderationdata.Range(func(k, v interface{}) bool {
		if v.(moderationData).blocked {
			if time.Now().Sub(v.(moderationData).blockedAt) <= roomBlockTTL {
				return true
			} else {
				r.moderationdata.Store(k.(peer.ID), moderationData{blocked: false})
				return true
			}
		}
		if v.(moderationData).banned {
			if time.Now().Sub(v.(moderationData).bannedAt) <= roomBanTTL {
				return true
			} else {
				r.moderationdata.Store(k.(peer.ID), moderationData{banned: false})
				return true
			}
		}
		return true
	})
}

func (r *Channel) setNickname(peerID peer.ID, nickname string) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	if p, found := r.participants[peerID]; found {
		p.Nickname = nickname
	}

	return false
}

func (r *Channel) getNickname(peerID peer.ID) (string, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if p, found := r.participants[peerID]; found {
		return p.Nickname, true
	}
	return "", false
}

// ChannelProcessor manages rooms through pubsub subscription and implements room operations.
type ChannelProcessor struct {
	logger         *zap.Logger
	ps             *pubsub.PubSub
	node           Node
	kadDHT         *dht.IpfsDHT
	bannedIPlist   *block.Blocklist
	boltdatabase   *boltdbstorage.Storage
	rooms          map[string]*Channel
	eventPublisher events.Publisher

	lock sync.RWMutex
}

// NewChannelProcessor creates a new room manager.
func NewChannelProcessor(logger *zap.Logger, node Node, kadDHT *dht.IpfsDHT, ps *pubsub.PubSub) (*ChannelProcessor, events.Subscriber) {

	logger, err := zap.NewDevelopment()
	evtPub, evtSub := events.NewSubscription()

	p := filepath.Join("testdata", "blocklist.cidr")
	f, err := os.Open(p)
	if err != nil {

	}
	iplist := block.New()
	_, err1 := iplist.Reload(f)
	if err1 != nil {
		logger.Debug("Error opening database")
	}

	var db *boltdbstorage.Storage
	go func() {
		db, err = setupDB()
		if err != nil {
			logger.Debug("Error opening database")
		}
	}()
	time.Sleep(time.Second * 2)

	mngr := &ChannelProcessor{
		logger:         logger,
		ps:             ps,
		node:           node,
		kadDHT:         kadDHT,
		bannedIPlist:   iplist,
		boltdatabase:   db,
		rooms:          make(map[string]*Channel),
		eventPublisher: evtPub,
	}
	go mngr.advertise()
	go mngr.refreshRoomsParticipants()
	//Longer running go routines active throughout the lifecycle of application, heartbeat/healing go routines concurrency pattern can be added to optimise these go routines
	go mngr.unblockunbanRoomsParticipants()
	return mngr, evtSub
}

// JoinAndSubscribe joins and subscribes to a room.
func (r *ChannelProcessor) JoinAndSubscribe(roomName string, nickname string) (bool, error) {
	if r.HasJoined(roomName) {
		return false, nil
	}

	logger := r.logger.With(zap.String("topic", roomName))

	cleanup := func(topic *pubsub.Topic, subscription *pubsub.Subscription) {
		if topic != nil {
			_ = topic.Close()
		}
		if subscription != nil {
			subscription.Cancel()
		}
	}

	logger.Debug("joining room topic")
	topicName := r.TopicName(roomName)
	topic, err := r.ps.Join(topicName)
	if err != nil {
		logger.Debug("failed joining room topic")
		return false, err
	}

	logger.Debug("subscribing to room topic")
	subscription, err := topic.Subscribe()
	if err != nil {
		logger.Debug("failed subscribing to room topic")

		cleanup(topic, subscription)
		return false, err
	}

	room := newChannel(roomName, topic, subscription)

	room.addParticipant(r.node.ID(), nickname, roomParticipantsTTLPermanent)
 
	//Cooling period feature implementation via token bucket Algorithm
	store, err := memorystore.New(&memorystore.Config{
		// Number of tokens/messages allowed per interval.
		Tokens: 60,

		// Interval until tokens reset, user can post messages again after tokens are available again after message quota expiry
		Interval: time.Minute,
	})
	room.ratelimitstore = store

	if err != nil {
		log.Fatal(err)
	}

	r.putRoom(room)

	go r.roomTopicEventHandler(room)
	go r.roomSubscriptionHandler(room)
	r.advertiseToRoom(room)
	logger.Debug("successfully joined room")
	return true, nil
}

// HasJoined returns whether the manager has joined a given room.
func (r *ChannelProcessor) HasJoined(roomName string) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	_, found := r.rooms[r.TopicName(roomName)]
	return found
}

// TopicName builds a string containing the name of the pubsub topic for a given room name.
func (r *ChannelProcessor) TopicName(roomName string) string {
	return fmt.Sprintf("chat/room/%s", roomName)
}

// SendChatMessage sends a chat message to a given room.
// Fails if it has not yet joined the given room.
func (r *ChannelProcessor) SendChatMessage(ctx context.Context, roomName string, msg message.Message) error {
	room, found := r.getRoom(roomName)
	if !found {
		return errors.New(fmt.Sprintf("must join the room before sending messages"))
	}

	rm := &RoomMessageOut{
		Type:    RoomMessageTypeChatMessage,
		Payload: msg,
	}

	if err := r.publishRoomMessage(ctx, room, rm); err != nil {
		return err
	}

	return nil
}

// SetNickname sets the node's nickname in a given room.
func (r *ChannelProcessor) SetNickname(roomName string, nickname string) error {
	room, found := r.getRoom(roomName)
	if !found {
		return errors.New("must join the room before setting nickname")
	}

	room.setNickname(r.node.ID(), nickname)
	return nil
}

// GetNickname tries to find the nickname of a peer in the DHT.
func (r *ChannelProcessor) GetNickname(
	roomName string,
	peerID peer.ID,
) (string, bool, error) {
	if room, found := r.getRoom(roomName); found {
		nickname, found := room.getNickname(peerID)
		return nickname, found, nil
	}

	return "", false, errors.New("must join the room before getting nicknames")
}

// GetRoomParticipants returns the list of peers in a room.
func (r *ChannelProcessor) GetRoomParticipants(roomName string) ([]participantsEntry, error) {
	room, found := r.getRoom(roomName)
	if !found {
		return nil, errors.New("must join the room before getting participants")
	}

	// always append the node to the participants list
	return room.getParticipants(), nil
}

func (r *ChannelProcessor) putRoom(room *Channel) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.rooms[room.topic.String()] = room
}

func (r *ChannelProcessor) getRoom(roomName string) (*Channel, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	room, found := r.rooms[r.TopicName(roomName)]
	return room, found
}

func (r *ChannelProcessor) roomTopicEventHandler(room *Channel) {
	handler, err := room.topic.EventHandler()
	if err != nil {
		r.logger.Error("failed getting room topic event handler", zap.Error(err))
	}

	for {
		peerEvt, err := handler.NextPeerEvent(context.Background())
		if err != nil {
			r.logger.Error("failed receiving room topic peer event", zap.Error(err))
			continue
		}

		var evt events.Event

		switch peerEvt.Type {
		case pubsub.PeerJoin:
			evt = &events.PeerJoined{
				PeerID:   peerEvt.Peer,
				RoomName: room.name,
			}
			room.addParticipant(peerEvt.Peer, "", roomParticipantsTTL)

		case pubsub.PeerLeave:
			evt = &events.PeerLeft{
				PeerID:   peerEvt.Peer,
				RoomName: room.name,
			}
			room.removeParticipant(peerEvt.Peer)
		}

		if evt == nil {
			continue
		}

		if err := r.eventPublisher.Publish(evt); err != nil {
			r.logger.Error("failed publishing room manager event", zap.Error(err))
		}
	}
}

func (r *ChannelProcessor) roomSubscriptionHandler(room *Channel) {
	for {
		subMsg, err := room.subscription.Next(context.Background())
		if err != nil {
			r.logger.Error("failed receiving room message", zap.Error(err))
			continue
		}

		var rm RoomMessageIn
		if err := json.Unmarshal(subMsg.Data, &rm); err != nil {
			r.logger.Warn("ignoring room message", zap.Error(err))
		}

		//var moderation bool

		switch rm.Type {
		case RoomMessageTypeChatMessage:
			var chatMessage message.Message
			if err := json.Unmarshal(rm.Payload, &chatMessage); err != nil {
				r.logger.Warn(
					"ignoring message",
					zap.Error(errors.Wrap(err, "unmarshalling payload")),
				)

				continue

			}
			/*
				if room.moderationdata[chatMessage.SenderID] != nil {

					if room.moderationdata[chatMessage.SenderID].blocked == true {
						continue
					}

					if room.moderationdata[chatMessage.SenderID].banned == true {
						continue
					}
				}
			*/
			moderationvalue, found := room.moderationdata.Load(chatMessage.SenderID)
			if found {
				if moderationvalue.(moderationData).blocked == true {
					continue
				}
				if moderationvalue.(moderationData).banned == true {
					continue
				}
			}

			/*
				re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
				fmt.Printf("Pattern: %v\n", re.String())
				submatchall := re.FindAllString(, -1)
				if r.bannedIPlist.Blocked(net.ParseIP(submatchall[0])) {

					msg := message.Message{
						SenderID:  "Moderator",
						Timestamp: time.Now(),
						Value:     chatMessage.SenderID.Pretty() + ":" + "User Banned",
					}

					err := r.eventPublisher.Publish(&events.NewMessage{
						Message:  msg,
						RoomName: room.name,
					})
					if err != nil {
						r.logger.Error("failed publishing room manager event", zap.Error(err))
						continue
					}
					continue
				}
			*/
			topic := boltdbstorage.Topic{ChatMessage: chatMessage.Value, SenderId: chatMessage.SenderID.String()}
			if r.boltdatabase != nil {
				err = r.boltdatabase.WriteTopic(room.name, &topic)
				r.logger.Debug("Data written to bolt db")
				if err != nil {
					r.logger.Debug("Error Writing chat history", zap.Error(err))
				}
			}

			ctx := context.Background()
			_, _, reset, ok, err := room.ratelimitstore.Take(ctx, chatMessage.SenderID.String())
			if err != nil {
				r.logger.Error("failed publishing room manager event", zap.Error(err))
			}
			timeFromTS := time.Unix(0, int64(reset))

			if !ok {
				msg := message.Message{
					SenderID:  "Moderator",
					Timestamp: time.Now(),
					Value:     chatMessage.SenderID.Pretty() + ":" + "User Rate Limited" + ":" + "Post message at:" + timeFromTS.String(),
				}

				err := r.eventPublisher.Publish(&events.NewMessage{
					Message:  msg,
					RoomName: room.name,
				})

				r.logger.Debug("User Rate limited")

				if err != nil {
					r.logger.Error("failed publishing room manager event", zap.Error(err))
					continue
				}
				continue
			}

			resp, err := http.Get("http://localhost:82/IAB/" + chatMessage.Value + "," + "abusive")
			if err != nil {
				log.Fatalln(err)
			}
			score, err := io.ReadAll(resp.Body)

			similarityscore, err := strconv.ParseFloat(strings.TrimRight(string(score), "!\n"), 32)

			//Moderation module
			resp, err = http.Get("http://localhost:82/IAB/" + chatMessage.Value + "," + room.name)
			if err != nil {
				log.Fatalln(err)
			}

			score1, err := io.ReadAll(resp.Body)
			topicrelevancescore, err := strconv.ParseFloat(strings.TrimRight(string(score1), "!\n"), 32)

			if err != nil {
				r.logger.Error("Error in calculating similarity", zap.Error(err))
			}
			r.logger.Debug("Abusive content score", zap.Float64("Abuse Score", similarityscore))
			r.logger.Debug("Topic relevance score", zap.Float64("Quality Score", topicrelevancescore))
			var moderation bool

			if topicrelevancescore > 0.4 {

				r.logger.Debug("Quality content detected")
				msg := message.Message{
					SenderID:  "Moderator",
					Timestamp: time.Now(),
					Value:     chatMessage.SenderID.Pretty() + ":" + "Kool content",
				}

				err := r.eventPublisher.Publish(&events.NewMessage{
					Message:  msg,
					RoomName: room.name,
				})
				if err != nil {
					r.logger.Error("failed publishing room manager event", zap.Error(err))
					continue
				}

			} else {

			}

			if similarityscore > 0.3 {
				moderation = true
				r.logger.Debug("Abusive content detected")
			} else {

				if subMsg.ReceivedFrom == r.node.ID() {
					continue
				}

			}

			if moderation {

				blockdata, found := room.usermoderationCounter.Load(chatMessage.SenderID)

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func(found bool) {
					defer wg.Done()
					if found {
						room.usermoderationCounter.Store(chatMessage.SenderID, room.counter+1)
						room.counter = blockdata.(int) + 1
					} else {
						room.usermoderationCounter.Store(chatMessage.SenderID, 1)
						room.counter = 1
					}
				}(found)

				wg.Wait()

				r.logger.Debug("Moderation Counter", zap.Int("Moderation", room.counter))

				if room.counter < 5 {

					msg := message.Message{
						SenderID:  "Moderator",
						Timestamp: time.Now(),
						Value:     chatMessage.SenderID.Pretty() + ":" + "Content Warning",
					}

					err := r.eventPublisher.Publish(&events.NewMessage{
						Message:  msg,
						RoomName: room.name,
					})
					if err != nil {
						r.logger.Error("failed publishing room manager event", zap.Error(err))
						continue
					}

				} else if room.counter >= 5 && room.counter < 10 {

					room.moderationdata.Store(chatMessage.SenderID, moderationData{blockedAt: time.Now(), blocked: true})
					msg := message.Message{
						SenderID:  "Moderator",
						Timestamp: time.Now(),
						Value:     chatMessage.SenderID.Pretty() + ":" + "User Blocked",
					}

					err := r.eventPublisher.Publish(&events.NewMessage{
						Message:  msg,
						RoomName: room.name,
					})

					if err != nil {
						r.logger.Error("failed publishing room manager event", zap.Error(err))
						continue
					}

				} else {

					room.moderationdata.Store(chatMessage.SenderID, moderationData{bannedAt: time.Now(), banned: true})

					msg := message.Message{
						SenderID:  "Moderator",
						Timestamp: time.Now(),
						Value:     chatMessage.SenderID.Pretty() + ":" + "User Banned",
					}

					err := r.eventPublisher.Publish(&events.NewMessage{
						Message:  msg,
						RoomName: room.name,
					})
					if err != nil {
						r.logger.Error("failed publishing room manager event", zap.Error(err))
						continue
					}

				}
			} else {
				err := r.eventPublisher.Publish(&events.NewMessage{
					Message:  chatMessage,
					RoomName: room.name,
				})
				if err != nil {
					r.logger.Error("failed publishing room manager event", zap.Error(err))
				}
			}

		case RoomMessageTypeModerateMessage:
			var moderateMessage message.Message
			if err := json.Unmarshal(rm.Payload, &moderateMessage); err != nil {
				r.logger.Warn(
					"ignoring message",
					zap.Error(errors.Wrap(err, "unmarshalling payload")),
				)
				continue
			}

			err := r.eventPublisher.Publish(&events.NewMessage{
				Message:  moderateMessage,
				RoomName: room.name,
			})
			if err != nil {
				r.logger.Error("failed publishing room manager event", zap.Error(err))
			}

		case RoomMessageTypeBlockMessage:
			var moderateMessage message.Message
			if err := json.Unmarshal(rm.Payload, &moderateMessage); err != nil {
				r.logger.Warn(
					"ignoring message",
					zap.Error(errors.Wrap(err, "unmarshalling payload")),
				)
				continue
			}

			err := r.eventPublisher.Publish(&events.NewMessage{
				Message:  moderateMessage,
				RoomName: room.name,
			})
			if err != nil {
				r.logger.Error("failed publishing room manager event", zap.Error(err))
			}

		case RoomMessageTypeBanMessage:
			var moderateMessage message.Message
			if err := json.Unmarshal(rm.Payload, &moderateMessage); err != nil {
				r.logger.Warn(
					"ignoring message",
					zap.Error(errors.Wrap(err, "unmarshalling payload")),
				)
				continue
			}

			err := r.eventPublisher.Publish(&events.NewMessage{
				Message:  moderateMessage,
				RoomName: room.name,
			})
			if err != nil {
				r.logger.Error("failed publishing room manager event", zap.Error(err))
			}
		case RoomMessageTypeAdvertise:
			var nickname string
			if err := json.Unmarshal(rm.Payload, &nickname); err != nil {
				r.logger.Warn("ignoring message", zap.Error(errors.Wrap(err, "unmarshalling payload")))
				continue
			}
			room.addParticipant(subMsg.ReceivedFrom, nickname, roomParticipantsTTL)

		default:
			r.logger.Warn(
				"ignoring room message",
				zap.Error(errors.New("unknown room message type")),
			)
		}
	}
}

func (r *ChannelProcessor) publishRoomMessage(
	ctx context.Context,
	room *Channel,
	rm *RoomMessageOut,
) error {
	rmJSON, err := json.Marshal(rm)
	if err != nil {
		return errors.Wrap(err, "marshalling message")
	}

	if err := room.topic.Publish(ctx, rmJSON); err != nil {
		return err
	}

	return nil
}

func (r *ChannelProcessor) advertise() {
	tick := time.Tick(time.Second * 5)

	for {
		<-tick

		func() {
			r.lock.RLock()
			defer r.lock.RUnlock()

			for _, room := range r.rooms {
				r.advertiseToRoom(room)
			}
		}()
	}
}

func (r *ChannelProcessor) advertiseToRoom(room *Channel) {
	// fetch this node's nickname
	thisNickname, _ := room.getNickname(r.node.ID())

	rm := RoomMessageOut{
		Type:    RoomMessageTypeAdvertise,
		Payload: thisNickname,
	}

	if err := r.publishRoomMessage(context.Background(), room, &rm); err != nil {
		r.logger.Error(
			"failed publishing room advertise",
			zap.Error(err),
			zap.String("room", room.topic.String()),
		)
	}
}

func (r *ChannelProcessor) refreshRoomsParticipants() {
	tick := time.Tick(time.Second)

	for {
		<-tick

		func() {
			r.lock.RLock()
			defer r.lock.RUnlock()

			for _, room := range r.rooms {
				room.refreshParticipants(func(peerID peer.ID) {
					// consider that if we haven't hear of this peer for a while, it disconnected from the room
					err := r.eventPublisher.Publish(&events.PeerLeft{
						PeerID:   peerID,
						RoomName: room.name,
					})
					if err != nil {
						r.logger.Error("failed publishing room manager event", zap.Error(err))
					}
				})
			}
		}()
	}
}

func (r *ChannelProcessor) unblockunbanRoomsParticipants() {
	tick := time.Tick(time.Second)

	for {
		<-tick

		func() {
			r.lock.RLock()
			defer r.lock.RUnlock()

			for _, room := range r.rooms {
				room.unblockunbanParticipants()
			}
		}()

	}
}
