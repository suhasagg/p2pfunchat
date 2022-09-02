package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-libp2p-chat/go-node/storage/boltdbstorage"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

//Module to sync boltdb message history db with hbase p2p chat database tracker

type Messages struct {
	ChatRoomId int       `json:" chatRoomId"`
	Messages   []Message `json:"messages"`
}

type Message struct {
	PeerId    string `json:"peerId"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
	Index     string `json:"index"`
}

type Room struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Created      string   `json:"created"`
	Participants []string `json:"participants"`
	Meta         RoomMeta `json:"meta"`
}

type RoomMeta struct {
	MessageCount         int `json:"messagecount"`
	Created              int `json:"created"`
	LastMessageTimestamp int `json:"lastmessagetimestamp"`
}

func main() {
	logger, err := zap.NewDevelopment()
	db, err := bolt.Open("4.db", 0600, nil)
	fmt.Println(err)
	storage, err := boltdbstorage.New(db, []byte("topic"))
	kvs, err := storage.ReadTopicBucket("whale")
	messages := make(map[string]Message)
	for _, v := range kvs {
		value := strings.Split(v.Key, ":")
		timestamp, err := strconv.Atoi(value[1])
		fmt.Println(err)
		if value[0] == "chatmessage" {
			messages[value[1]] = Message{
				Message:   v.Value,
				Timestamp: timestamp,
				Index:     "1",
			}
		}
		if value[0] == "senderId" {
			messageData := messages[value[1]]
			messageData.PeerId = v.Value
			messages[value[1]] = messageData

		}
	}

	meta := RoomMeta{
		MessageCount:         10,
		Created:              10,
		LastMessageTimestamp: 1565339958141521920,
	}

	room := Room{
		Id:           "1",
		Name:         "whale",
		Created:      "1565339958141521920",
		Participants: []string{"aaa"},
		Meta:         meta,
	}

	client := &http.Client{}
	// marshal User to json
	jsonData, err := json.Marshal(room)
	if err != nil {
		panic(err)
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, "http://localhost:9991/room", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.StatusCode)

	var messageData []Message
	for _, v := range messages {
		messageData = append(messageData, v)
		logger.Debug(v.Message + "," + v.Index + "," + v.PeerId + "," + strconv.Itoa(v.Timestamp))
	}

	messagelist := Messages{Messages: messageData, ChatRoomId: 1}

	jsonData, err = json.Marshal(messagelist)
	if err != nil {
		panic(err)
	}

	// set the HTTP method, url, and request body
	req, err = http.NewRequest(http.MethodPut, "http://localhost:9991/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.StatusCode)

}
