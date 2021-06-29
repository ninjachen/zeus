package pubsub

import (
	"fmt"
	"sync"
	"testing"

	"github.com/dgraph-io/badger/v2"

)

func TestEvent(t *testing.T) {
	opt := badger.DefaultOptions("test_badger")
	opt.Truncate = true

	db, err := badger.Open(opt)
	if err != nil {
		t.Fatal(err)
	}
	hub, err := newHub(&Config{
		Type:        "pubsub",
		TopicPrefix: "test_sandwich",
		Setting: map[string]interface{}{
			"projectID":       "sandwich-311704",
			"credentialsFile": "sandwich-311704-0c506850c4a0.json",
		},
	}, db)
	if err != nil {
		t.Fatal(err)
	}
	err = hub.Start()
	if err != nil {
		t.Fatal(err)
	}
	topic := "testtopic8"

	wait := sync.WaitGroup{}
	wait.Add(5)
	err = hub.Sub(topic, func(msg *Message) {
		fmt.Println(msg.UUID(), msg.Payload())
		msg.Ack()
		wait.Done()
	})
	if err != nil {
		t.Fatal(err)
	}

	err = hub.AsyncPub(topic, NewMessage([]byte("data")))
	if err != nil {
		t.Fatal(err)
	}

	wait.Wait()
}
