/*
 * Copyright 2021 InfAI (CC SES)
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
	"log"
	"sync"
	"time"
)

type Kafka struct {
	group     string
	broker    string
	publisher *Publisher
	ctx       context.Context
	wg        *sync.WaitGroup
	debug     bool
}

func Init(ctx context.Context, wg *sync.WaitGroup, kafkaBootstrap string, group string, debug bool) (Interface, error) {
	publisher, err := NewPublisher(kafkaBootstrap, ctx, wg)
	k := Kafka{broker: kafkaBootstrap, group: group, debug: debug, publisher: publisher, ctx: ctx, wg: wg}
	return &k, err
}

func (k *Kafka) Consume(topic string, listener func(topic string, delivery []byte, t time.Time) error) (err error) {
	_, err = NewConsumer(k.ctx, k.wg, k.broker, []string{topic}, k.group, Earliest, listener, errorhandler, k.debug)
	return
}

func (k *Kafka) Publish(topic string, key string, payload []byte) error {
	return k.publisher.Publish(topic, key, payload)
}

func errorhandler(err error, consumer *Consumer) {
	log.Println("ERROR: Encountered error on consumer with topic", consumer.topics, err.Error())
}
