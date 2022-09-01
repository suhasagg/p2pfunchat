package main

import (
	"fmt"
	"github.com/go-libp2p-chat/go-node/storage/boltdbstorage"
	bolt "go.etcd.io/bbolt"
)

func main() {

	db, err := bolt.Open("messagehistory.db", 0600, nil)
	if err != nil {

	}
	storage, err := boltdbstorage.New(db, []byte("topic"))
	if err != nil {

	}
	fmt.Println("DB Setup Done")

	topic := boltdbstorage.Topic{ChatMessage: "red fish", SenderId: "a"}
	err = storage.WriteTopic("whale", &topic)

	topic = boltdbstorage.Topic{ChatMessage: "blue fish", SenderId: "b"}
	err = storage.WriteTopic("whale", &topic)

	topic = boltdbstorage.Topic{ChatMessage: "green fish", SenderId: "c"}
	err = storage.WriteTopic("whale", &topic)

	//topicData, err := storage.ReadTopic("whale")
	if err != nil {

	}
	//fmt.Println(topicData.ChatMessage)
}
