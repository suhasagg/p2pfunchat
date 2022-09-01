package boltdbstorage

import (
	"encoding/json"
)

// Spec contains fields for serialising topic information
type Topic struct {
	ChatMessage        string
	SenderId           string
	UUId               string
	ModeratorMessage   string
	AdvertisingMessage string
	ReactionMessage    string
	UserList           []string
	BlockuserList      []string
	BanuserList        []string
}

type jsontopicSpec struct {
	ChatMessage        string
	UUId               string
	SenderId           string
	ModeratorMessage   string
	AdvertisingMessage string
	ReactionMessage    string
	UserList           []string
	BlockuserList      []string
	BanuserList        []string
}

// MarshalJSON converts the Spec to a JSON string.
func (t *Topic) MarshalJSON() ([]byte, error) {
	j := jsontopicSpec{
		ChatMessage:        t.ChatMessage,
		UUId:               t.UUId,
		SenderId:           t.SenderId,
		ModeratorMessage:   t.ModeratorMessage,
		AdvertisingMessage: t.AdvertisingMessage,
		ReactionMessage:    t.ReactionMessage,
		UserList:           t.UserList,
		BlockuserList:      t.BlockuserList,
		BanuserList:        t.BanuserList,
	}
	return json.Marshal(j)
}

// UnmarshalJSON fills the Spec from a JSON string.
func (t *Topic) UnmarshalJSON(b []byte) error {
	var j jsontopicSpec
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	t.ChatMessage = j.ChatMessage
	t.UUId = j.UUId
	t.SenderId = j.SenderId
	t.ModeratorMessage = j.ModeratorMessage
	t.AdvertisingMessage = j.AdvertisingMessage
	t.ReactionMessage = j.ReactionMessage
	t.UserList = j.UserList
	t.BlockuserList = j.BlockuserList
	t.BanuserList = j.BanuserList
	return nil
}

//contains fields for serialising user semantics

type User struct {
	Geography                string
	UserContentCategories    []string
	UserPersonality          string
	AverageUserSessionLength int
	UserFrequency            string
	PeerId                   string
	UserHandle               string
}

type jsonuserSpec struct {
	Geography                string
	UserContentCategories    []string
	UserPersonality          string
	AverageUserSessionLength int
	UserFrequency            string
	PeerId                   string
	UserHandle               string
}

// MarshalJSON converts the Spec to a JSON string.
func (u *User) MarshalJSON() ([]byte, error) {
	j := jsonuserSpec{
		Geography:                u.Geography,
		UserContentCategories:    u.UserContentCategories,
		UserPersonality:          u.UserPersonality,
		AverageUserSessionLength: u.AverageUserSessionLength,
		UserFrequency:            u.UserFrequency,
		PeerId:                   u.PeerId,
		UserHandle:               u.UserHandle,
	}
	return json.Marshal(j)
}

// UnmarshalJSON fills the Spec from a JSON string.
func (u *User) UnmarshalJSON(b []byte) error {
	var j jsonuserSpec
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	u.Geography = j.Geography
	u.UserContentCategories = j.UserContentCategories
	u.UserPersonality = j.UserPersonality
	u.AverageUserSessionLength = j.AverageUserSessionLength
	u.UserFrequency = j.UserFrequency
	u.PeerId = j.PeerId
	u.UserHandle = j.UserHandle
	return nil
}
