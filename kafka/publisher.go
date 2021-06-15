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
	"github.com/Shopify/sarama"
	"log"
	"runtime/debug"
	"sync"
)

type Publisher struct {
	kafkaBootstrap string
	syncProducer   sarama.SyncProducer
}

func NewPublisher(kafkaBootstrap string, ctx context.Context, wg *sync.WaitGroup) (*Publisher, error) {
	p := &Publisher{kafkaBootstrap: kafkaBootstrap}
	var err error
	p.syncProducer, err = p.ensureConnection()
	wg.Add(1)
	go func() {
		<-ctx.Done()
		if p.syncProducer != nil {
			_ = p.syncProducer.Close()
		}
		log.Println("Kafka Publisher closed")
		wg.Done()
	}()
	return p, err
}

func (publisher *Publisher) ensureConnection() (syncProducer sarama.SyncProducer, err error) {
	if publisher.syncProducer != nil {
		return publisher.syncProducer, nil
	}
	kafkaConf := sarama.NewConfig()
	kafkaConf.Producer.Return.Successes = true
	syncP, err := sarama.NewSyncProducer([]string{publisher.kafkaBootstrap}, kafkaConf)
	if err != nil {
		publisher.syncProducer = syncP
	}
	return syncP, err
}

func (this *Publisher) Publish(topic string, key string, payload []byte) error {
	if this.syncProducer == nil {
		var err error
		this.syncProducer, err = this.ensureConnection()
		if err != nil {
			return err
		}
	}
	_, _, err := this.syncProducer.SendMessage(&sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(payload), Key: sarama.StringEncoder(key)})
	if err != nil {
		debug.PrintStack()
	}
	return err
}
