package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dgraph-io/badger/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/segmentio/ksuid"
	"go.yym.plus/zeus/pkg/log"
	"google.golang.org/api/option"
)

type Hub struct {
	subscriber  message.Subscriber
	publisher   message.Publisher
	conf        *Config
	middlewares []MiddlewareFunc
	logger      watermill.LoggerAdapter
	db          *badger.DB
	pushBuffer  *Buffer
}

type Config struct {
	Type        string
	GroupID     string
	TopicPrefix string
	Debug       bool
	Setting     map[string]interface{}
}

type PushItem struct {
	topic string
	msg   *Message
}

type Buffer struct {
	cond     *sync.Cond
	capacity int
	items    []*PushItem
	closed   bool
}

type Message struct {
	original *message.Message
}

type PubSubConfig struct {
	ProjectID       string
	CredentialsFile string
}

type MiddlewareFunc func(*Message) error

func NewHub(conf *Config, db *badger.DB) (*Hub, error) {
	hub, err := newHub(conf, db)

	if err != nil {
		return nil, err
	}

	return hub, nil
}

func NewMessage() *Message {
	return newMessage(ksuid.New().String(), nil)
}

func newMessage(uuid string, payload []byte) *Message {
	return &Message{
		original: message.NewMessage(uuid, payload),
	}
}

func newHub(conf *Config, db *badger.DB) (*Hub, error) {
	hub := Hub{
		conf:   conf,
		logger: watermill.NewStdLogger(conf.Debug, false),
		db:     db,
		pushBuffer: &Buffer{
			cond:     sync.NewCond(&sync.Mutex{}),
			capacity: 1000,
			closed:   false,
		},
	}

	err := hub.init()
	if err != nil {
		return nil, err
	}
	return &hub, nil
}

type Handler func(msg *Message)

func (self *Hub) Pub(topic string, msg *Message) error {
	return self.publisher.Publish(self.conf.TopicPrefix+"_"+topic, msg.original)
}

func (self *Hub) AsyncPub(topic string, msg *Message) error {
	err := self.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(msg.original)
		if err != nil {
			return err
		}
		return txn.Set([]byte(fmt.Sprintf("push_%s:%s", topic, msg.original.UUID)), data)
	})
	if err != nil {
		return err
	}
	return self.pushBuffer.Push(&PushItem{
		topic: topic,
		msg:   msg,
	})
}

func (self *Hub) Sub(topic string, handler Handler) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	messages, err := self.subscriber.Subscribe(ctx, self.conf.TopicPrefix+"_"+topic)
	if err != nil {
		return err
	}
	go func() {
		for msg := range messages {
			wrapperMsg := &Message{original: msg}
			for _, middleware := range self.middlewares {
				err = middleware(wrapperMsg)
				if err != nil {
					log.WithError(err).Error("exec sub message middleware error")
					break
				}
			}
			if err == nil {
				handler(wrapperMsg)
			}
		}
	}()
	return nil
}

func (self *Hub) init() error {
	var err error
	self.publisher, err = self.createPublisher()
	if err != nil {
		return err
	}
	self.subscriber, err = self.createSubscriber()
	if err != nil {
		return err
	}
	err = self.loadUnFinishPush(true)
	if err != nil {
		return err
	}
	return nil
}

func (self *Hub) Start() error {
	go self.runAsyncPub()
	return nil
}

func (self *Hub) loadUnFinishPush(enable bool) error {
	if !enable {
		return nil
	}
	prefix := []byte("push_")
	items := []*PushItem{}
	err := self.db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().Key()
			key = key[5:]
			topicAndUid := strings.Split(string(key), ":")
			msg := message.Message{}
			value, err := it.Item().ValueCopy(nil)
			if err != nil {
				return err
			}
			err = json.Unmarshal(value, &msg)
			if err != nil {
				return err
			}
			items = append(items, &PushItem{
				topic: topicAndUid[0],
				msg:   &Message{original: &msg},
			})
		}
		return nil
	})

	if err != nil {
		return err
	}
	return self.pushBuffer.Push(items...)
}

func (self *Hub) runAsyncPub() {
	var item *PushItem
	var err error
	for {
		if item == nil {
			item, err = self.pushBuffer.Pop()
			if err != nil {
				return
			}
		}
		err = self.Pub(item.topic, item.msg)
		if err != nil {
			log.WithError(err).Errorw("push message error", "topic", item.topic)
			time.Sleep(time.Second)
			continue
		} else {
			err = self.db.Update(func(txn *badger.Txn) error {
				return txn.Delete([]byte(fmt.Sprintf("push_%s:%s", item.topic, item.msg.original.UUID)))
			})
			if err != nil {
				log.WithError(err).Errorw("delete async push message error", "topic", item.topic)
			}
			item = nil
		}
	}
}

func (self *Hub) Stop() error {
	self.subscriber.Close()
	self.publisher.Close()
	self.pushBuffer.Close()
	return nil
}

func (self *Hub) createSubscriber() (message.Subscriber, error) {
	if self.conf.Type == "pubsub" {
		conf := PubSubConfig{}
		err := self.decode(self.conf.Setting, &conf)
		if err != nil {
			return nil, err
		}

		return googlecloud.NewSubscriber(googlecloud.SubscriberConfig{
			ProjectID: conf.ProjectID,
			ClientOptions: []option.ClientOption{
				option.WithCredentialsFile(conf.CredentialsFile),
			},
			GenerateSubscriptionName: func(topic string) string {
				return topic + "_" + self.conf.GroupID
			},
		}, self.logger)
	} else {
		return nil, fmt.Errorf("not support")
	}
}

func (self *Hub) createPublisher() (message.Publisher, error) {
	if self.conf.Type == "pubsub" {
		conf := PubSubConfig{}
		err := self.decode(self.conf.Setting, &conf)
		if err != nil {
			return nil, err
		}

		return googlecloud.NewPublisher(googlecloud.PublisherConfig{
			ProjectID: conf.ProjectID,
			ClientOptions: []option.ClientOption{
				option.WithCredentialsFile(conf.CredentialsFile),
			},
		}, self.logger)
	} else {
		return nil, fmt.Errorf("not support")
	}
}
func (self *Hub) decode(input, output interface{}) error {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	})
	if err != nil {
		return err
	}

	return d.Decode(input)
}

func (self *Buffer) Push(item ...*PushItem) error {
	self.cond.L.Lock()
	defer self.cond.L.Unlock()

	for self.capacity > 0 && len(self.items) >= self.capacity && !self.closed {
		self.cond.Wait()
	}
	if self.closed {
		return fmt.Errorf("closed")
	}
	self.items = append(self.items, item...)

	self.cond.Broadcast()
	return nil
}

func (self *Buffer) PushFront(item ...*PushItem) error {
	self.cond.L.Lock()
	defer self.cond.L.Unlock()

	for self.capacity > 0 && len(self.items) >= self.capacity && !self.closed {
		self.cond.Wait()
	}
	if self.closed {
		return fmt.Errorf("closed")
	}
	self.items = append(item, self.items...)

	self.cond.Broadcast()
	return nil
}

func (self *Buffer) Pop() (*PushItem, error) {
	self.cond.L.Lock()
	defer self.cond.L.Unlock()

	for len(self.items) == 0 && !self.closed {
		self.cond.Wait()
	}

	if self.closed {
		return nil, fmt.Errorf("closed")
	}

	item := self.items[0]
	self.items = self.items[1:]
	self.cond.Broadcast()

	return item, nil
}

func (self *Buffer) PopN(size int) ([]*PushItem, error) {
	self.cond.L.Lock()
	defer self.cond.L.Unlock()

	for len(self.items) == 0 && !self.closed {
		self.cond.Wait()
	}

	if self.closed {
		return nil, fmt.Errorf("closed")
	}

	self.cond.Broadcast()
	if size <= 0 || size >= len(self.items) {
		self.items = nil
		return self.items, nil
	}

	self.items = self.items[size:]
	return self.items[0:size], nil
}

func (self *Buffer) Close() {
	self.closed = true
	self.cond.Broadcast()
}

func (self *Message) Payload() []byte {
	return self.original.Payload
}

func (self *Message) SetPayloadData(data []byte) {
	self.original.Payload = data
}

func (self *Message) SetPayload(value interface{}) error {
	if v, ok := value.([]byte); ok {
		self.SetPayloadData(v)
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	self.original.Payload = data
	return nil
}

func (self *Message) UnmarshalPayload(out interface{}) error {
	return json.Unmarshal(self.original.Payload, out)
}

func (self *Message) UUID() string {
	return self.original.UUID
}

func (self *Message) Ack() bool {
	return self.original.Nack()
}

func (self *Message) Nack() bool {
	return self.original.Nack()
}

func (self *Message) GetMeta(key string) string {
	return self.original.Metadata.Get(key)
}

func (self *Message) SetMeta(key string, value string) {
	self.original.Metadata.Set(key, value)
}
