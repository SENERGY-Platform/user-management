/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"strings"
	"sync"
	"time"
)

// const Latest = sarama.OffsetNewest
const Earliest = sarama.OffsetOldest

func NewConsumer(ctx context.Context, wg *sync.WaitGroup, kafkaBootstrap string, topics []string, groupId string, offset int64, listener func(topic string, msg []byte, time time.Time) error, errorhandler func(err error, consumer *Consumer), debug bool) (consumer *Consumer, err error) {
	consumer = &Consumer{ctx: ctx, wg: wg, kafkaBootstrap: kafkaBootstrap, topics: topics, listener: listener, errorhandler: errorhandler, offset: offset, ready: make(chan bool), groupId: groupId, debug: debug}
	err = consumer.start()
	if err != nil {
		go func(err2 error) {
			for err2 != nil {
				time.Sleep(10 * time.Second)
				err2 = consumer.start()
				if err2 != nil {
					log.Println("WARN: Consumer still not ready:", err2)
				} else {
					log.Println("Consumer initiated successfully")
				}
			}
		}(err)
	}
	return
}

type Consumer struct {
	count          int
	kafkaBootstrap string
	topics         []string
	ctx            context.Context
	wg             *sync.WaitGroup
	listener       func(topic string, msg []byte, time time.Time) error
	errorhandler   func(err error, consumer *Consumer)
	mux            sync.Mutex
	offset         int64
	groupId        string
	ready          chan bool
	debug          bool
}

func (this *Consumer) start() error {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = this.offset

	client, err := sarama.NewConsumerGroup(strings.Split(this.kafkaBootstrap, ","), this.groupId, config)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-this.ctx.Done():
				log.Println("close kafka reader")
				return
			default:
				if err := client.Consume(this.ctx, this.topics, this); err != nil {
					log.Panicf("Error from consumer: %v", err)
				}
				// check if context was cancelled, signaling that the consumer should stop
				if this.ctx.Err() != nil {
					return
				}
				this.ready = make(chan bool)
			}
		}
	}()

	<-this.ready // Await till the consumer has been set up
	log.Println("Kafka consumer up and running...")

	return err
}

func (this *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(this.ready)
	this.wg.Add(1)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (this *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Cleaned up kafka session")
	this.wg.Done()
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (this *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		select {
		case <-this.ctx.Done():
			log.Println("Ignoring queued kafka messages for faster shutdown")
			return nil
		default:
			if this.debug {
				log.Println(message.Topic, message.Timestamp, string(message.Value))
			}
			err := this.listener(message.Topic, message.Value, message.Timestamp)
			if err != nil {
				this.errorhandler(err, this)
			}
			session.MarkMessage(message, "")
		}
	}

	return nil
}
