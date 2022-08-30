// Package boltdbstorage provides a Resumer implementation that uses a Bolt database file as storage.
package boltdbstorage

import (
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
	"runtime/debug"
	"strconv"
)

// Keys for the persistent storage.
// Chat room data storage
var TopicKeys = struct {
	ChatMessage        []byte
	ModeratorMessage   []byte
	AdvertisingMessage []byte
	ReactionMessage    []byte
	UserList           []byte
	BlockuserList      []byte
	BanuserList        []byte
}{
	ChatMessage:        []byte("chatmessage"),
	ModeratorMessage:   []byte("moderatormessage"),
	AdvertisingMessage: []byte("advertisingmessage"),
	ReactionMessage:    []byte("reactionmessage"),
	UserList:           []byte("userlist"),
	BlockuserList:      []byte("blockuser"),
	BanuserList:        []byte("banuser"),
}

// Keys for User semantics
// Peer semantics data storage
var UserKeys = struct {
	Geography                []byte
	UserContentCategories    []byte
	UserPersonality          []byte
	AverageUserSessionLength []byte
	UserFrequency            []byte
	PeerId                   []byte
	UserHandle               []byte
}{
	Geography:                []byte("geography"),
	UserContentCategories:    []byte("usercontentcategories"),
	UserPersonality:          []byte("userpersonality"),
	AverageUserSessionLength: []byte("averageusersessionlength"),
	UserFrequency:            []byte("userlist"),
}

// Storage contains methods for saving/loading chat history to a BoltDB database (optimum for peer and p2p) (Very similar to blockchain databases used in peers)
type Storage struct {
	db     *bbolt.DB
	bucket []byte
}

// New returns a new Resumer.
func New(db *bbolt.DB, bucket []byte) (*Storage, error) {
	err := db.Update(func(tx *bbolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists(bucket)
		return err2
	})
	if err != nil {
		return nil, err
	}
	return &Storage{
		db:     db,
		bucket: bucket,
	}, nil
}

// Write the chat session information for `topic`.
func (s *Storage) WriteTopic(topic string, t *Topic) error {
	userList, err := json.Marshal(t.UserList)
	if err != nil {
		return err
	}
	blockuserList, err := json.Marshal(t.BlockuserList)
	if err != nil {
		return err
	}
	banuserList, err := json.Marshal(t.BanuserList)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.Bucket(s.bucket).CreateBucketIfNotExists([]byte(topic))
		if err != nil {
			return err
		}
		//Torrent information which is serialised to disk and is used to restore torrent state in next session
		_ = b.Put(TopicKeys.ChatMessage, []byte(t.ChatMessage))
		_ = b.Put(TopicKeys.AdvertisingMessage, []byte(t.AdvertisingMessage))
		_ = b.Put(TopicKeys.ModeratorMessage, []byte(t.ModeratorMessage))
		_ = b.Put(TopicKeys.ReactionMessage, []byte(t.ReactionMessage))
		_ = b.Put(TopicKeys.UserList, []byte(userList))
		_ = b.Put(TopicKeys.BanuserList, []byte(banuserList))
		_ = b.Put(TopicKeys.BlockuserList, []byte(blockuserList))

		return nil
	})
}

// Write the Semantics information for `peer`.
func (s *Storage) WritePeerInfo(user string, u *User) error {
	usercontentcategories, err := json.Marshal(u.UserContentCategories)
	sessionlength := strconv.Itoa(u.AverageUserSessionLength)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.Bucket(s.bucket).CreateBucketIfNotExists([]byte(user))
		if err != nil {
			return err
		}
		//Torrent information which is serialised to disk and is used to restore torrent state in next session
		_ = b.Put(UserKeys.Geography, []byte(u.Geography))
		_ = b.Put(UserKeys.AverageUserSessionLength, []byte(sessionlength))
		_ = b.Put(UserKeys.UserPersonality, []byte(u.UserPersonality))
		_ = b.Put(UserKeys.UserFrequency, []byte(u.UserFrequency))
		_ = b.Put(UserKeys.UserContentCategories, []byte(usercontentcategories))

		return nil
	})
}

func (s *Storage) ReadTopic(topics string) (topic *Topic, err error) {
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot read torrent %q from db: %s", topic, r)
		}
	}()
	err = s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket).Bucket([]byte(topics))
		if b == nil {
			return fmt.Errorf("bucket not found: %q", topic)
		}
		tspec := new(Topic)
		value := b.Get(TopicKeys.ChatMessage)
		tspec.ChatMessage = string(value)

		value = b.Get(TopicKeys.AdvertisingMessage)
		tspec.AdvertisingMessage = string(value)

		value = b.Get(TopicKeys.ModeratorMessage)
		tspec.ModeratorMessage = string(value)

		value = b.Get(TopicKeys.ReactionMessage)
		tspec.ReactionMessage = string(value)

		value = b.Get(TopicKeys.UserList)
		userlist := make([]string, 0)
		if value != nil {
			err = json.Unmarshal(value, &userlist)
			if err != nil {
				return err
			}
		}
		tspec.UserList = userlist

		value = b.Get(TopicKeys.BlockuserList)
		blockuserlist := make([]string, 0)
		if value != nil {
			err = json.Unmarshal(value, &blockuserlist)
			if err != nil {
				return err
			}
		}
		tspec.UserList = userlist

		value = b.Get(TopicKeys.BanuserList)
		banuserlist := make([]string, 0)
		if value != nil {
			err = json.Unmarshal(value, &banuserlist)
			if err != nil {
				return err
			}
		}
		tspec.BanuserList = banuserlist

		return nil
	})
	return
}

func (s *Storage) ReadUser(users string) (user *User, err error) {
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot read torrent %q from db: %s", users, r)
		}
	}()
	err = s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucket).Bucket([]byte(users))
		if b == nil {
			return fmt.Errorf("bucket not found: %q", users)
		}

		uspec := new(User)
		value := b.Get(UserKeys.Geography)
		uspec.Geography = string(value)

		value = b.Get(UserKeys.UserPersonality)
		uspec.UserPersonality = string(value)

		value = b.Get(UserKeys.PeerId)
		uspec.PeerId = string(value)

		value = b.Get(UserKeys.UserHandle)
		uspec.UserHandle = string(value)

		value = b.Get(UserKeys.UserContentCategories)
		usercontentcategories := make([]string, 0)
		if value != nil {
			err = json.Unmarshal(value, &usercontentcategories)
			if err != nil {
				return err
			}
		}
		uspec.UserContentCategories = usercontentcategories

		value = b.Get(UserKeys.UserFrequency)
		uspec.UserFrequency = string(value)

		value = b.Get(UserKeys.AverageUserSessionLength)
		uspec.AverageUserSessionLength, err = strconv.Atoi(string(value))
		if err != nil {
			return err
		}

		return nil
	})
	return
}
