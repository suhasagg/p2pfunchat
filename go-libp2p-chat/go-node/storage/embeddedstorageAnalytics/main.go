package main

import (
	"fmt"
	"github.com/go-libp2p-chat/go-node/storage/boltdbstorage"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"io"
	"log"
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
	kvs, err := storage.ReadTopicBucket("encyclopedia")
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
	for _, v := range messages {
		resp, err := http.Get("http://localhost:81/IABSegments/" + v.Message)
		if err != nil {
			log.Fatalln(err)
		}

		entities, err := io.ReadAll(resp.Body)
		entity := strings.Replace(string(entities), "\n", "", -1)
		topic := boltdbstorage.Topic{ChatMessage: v.Message, SenderId: v.PeerId, Persona: entity}
		err = storage.WriteTopic("persona", &topic)
		logger.Debug("Data written to bolt db")
		if err != nil {
			logger.Debug("Error Writing chat history", zap.Error(err))
		}
		logger.Debug("Entities", zap.String("entities", string(entities)))
	}
}
