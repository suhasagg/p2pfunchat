// Package boltdbstorage provides a Resumer implementation that uses a Bolt database file as storage.
package boltdbstorage

import (
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"go.etcd.io/bbolt"
	"strconv"
)

// Keys for the persistent storage.
// Chat room data storage
var TopicKeys = struct {
	ChatMessage        string
	UUId               string
	SenderId           string
	ModeratorMessage   string
	AdvertisingMessage string
	ReactionMessage    string
	UserList           string
	BlockuserList      string
	BanuserList        string
	Persona            string
}{
	ChatMessage:        "chatmessage",
	UUId:               "uuid",
	SenderId:           "senderId",
	ModeratorMessage:   "moderatormessage",
	AdvertisingMessage: "advertisingmessage",
	ReactionMessage:    "reactionmessage",
	UserList:           "userid",
	BlockuserList:      "blockuser",
	BanuserList:        "banuser",
	Persona:            "persona",
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
	Db     *bbolt.DB
	Bucket []byte
}

// Storage contains methods for saving/loading chat history to a BoltDB database (optimum for peer and p2p) (Very similar to blockchain databases used in peers)
type KV struct {
	Key   string
	Value string
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
		Db:     db,
		Bucket: bucket,
	}, nil
}

// Write the chat session information for `topic`.
func (s *Storage) WriteTopic(topic string, t *Topic) error {
	var userList, blockuserList, banuserList []byte
	var err error
	node, err := snowflake.NewNode(1)
	id := node.Generate().String()
	if t.UserList != nil {
		userList, err = json.Marshal(t.UserList)
		if err != nil {
			return err
		}
	}
	if t.BlockuserList != nil {
		blockuserList, err = json.Marshal(t.BlockuserList)
		if err != nil {
			return err
		}
	}
	if t.BanuserList != nil {
		banuserList, err = json.Marshal(t.BanuserList)
		if err != nil {
			return err
		}
	}

	return s.Db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.Bucket(s.Bucket).CreateBucketIfNotExists([]byte(topic))
		if err != nil {
			return err
		}
		//Message which is serialised to disk and is used to restore chat state in next session
		if t.ChatMessage != "" {
			_ = b.Put([]byte(TopicKeys.ChatMessage+":"+id), []byte(t.ChatMessage))
		}

		if t.SenderId != "" {
			_ = b.Put([]byte(TopicKeys.SenderId+":"+id), []byte(t.SenderId))
		}

		if t.Persona != "" {
			_ = b.Put([]byte(TopicKeys.Persona+":"+id), []byte(t.Persona))
		}

		if t.AdvertisingMessage != "" {
			_ = b.Put([]byte(TopicKeys.AdvertisingMessage+":"+id), []byte(t.AdvertisingMessage))
		}
		if t.ModeratorMessage != "" {
			_ = b.Put([]byte(TopicKeys.ModeratorMessage+":"+id), []byte(t.ModeratorMessage))
		}
		if t.ReactionMessage != "" {
			_ = b.Put([]byte(TopicKeys.ReactionMessage+":"+id), []byte(t.ReactionMessage))
		}
		if userList != nil {
			_ = b.Put([]byte(TopicKeys.UserList+":"+id), userList)
		}
		if banuserList != nil {
			_ = b.Put([]byte(TopicKeys.BanuserList+":"+id), banuserList)
		}
		if blockuserList != nil {
			_ = b.Put([]byte(TopicKeys.BlockuserList+":"+id), blockuserList)
		}
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
	return s.Db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.Bucket(s.Bucket).CreateBucketIfNotExists([]byte(user))
		if err != nil {
			return err
		}
		//Message which is serialised to disk and is used to restore chat state in next session
		_ = b.Put(UserKeys.Geography, []byte(u.Geography))
		_ = b.Put(UserKeys.AverageUserSessionLength, []byte(sessionlength))
		_ = b.Put(UserKeys.UserPersonality, []byte(u.UserPersonality))
		_ = b.Put(UserKeys.UserFrequency, []byte(u.UserFrequency))
		_ = b.Put(UserKeys.UserContentCategories, []byte(usercontentcategories))

		return nil
	})
}

func (s *Storage) ReadTopicBucket(bucket string) ([]KV, error) {
	messageKeyValue := []KV{}
	err := s.Db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("topic")).Bucket([]byte(bucket))
		// we need cursor for iteration
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			value := KV{Key: string(k), Value: string(v)}
			messageKeyValue = append(messageKeyValue, value)
		}
		// should return nil to complete the transaction
		return nil
	})
	return messageKeyValue, err
}

/*
func (s *Storage) ReadTopic(topics string) (topic *Topic, err error) {
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot read torrent %q from Db: %s", topic, r)
		}
	}()
	err = s.Db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.Bucket).Bucket([]byte(topics))
		if b == nil {
			return fmt.Errorf("Bucket not found: %q", topic)
		}
		tspec := new(Topic)
		value := b.Get(TopicKeys.ChatMessage)
		tspec.ChatMessage = string(value)

		value = b.Get(TopicKeys.SenderId)
		tspec.SenderId = string(value)

		value = b.Get(TopicKeys.UUId)
		tspec.UUId = string(value)

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
*/

/*
func (s *Storage) ReadUser(users string) (user *User, err error) {
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot read torrent %q from Db: %s", users, r)
		}
	}()
	err = s.Db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.Bucket).Bucket([]byte(users))
		if b == nil {
			return fmt.Errorf("Bucket not found: %q", users)
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
*/
