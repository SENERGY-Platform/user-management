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
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(kafkaUrl string, topic string, debug bool) (*Producer, error) {
	writer, err := GetKafkaWriter(kafkaUrl, topic, debug)
	if err != nil {
		return nil, err
	}
	return &Producer{writer: writer, topic: topic}, nil
}

func GetKafkaWriter(broker string, topic string, debug bool) (writer *kafka.Writer, err error) {
	logger := slog.NewLogLogger(slog.Default().Handler(), slog.LevelDebug)
	logger.SetPrefix("[KAFKA-PRODUCER] ")
	writer = &kafka.Writer{
		Addr:        kafka.TCP(broker),
		Topic:       topic,
		MaxAttempts: 10,
		Logger:      logger,
		Async:       false,
		BatchSize:   1,
		Balancer:    &kafka.Hash{},
	}
	return writer, err
}

func (this *Producer) Produce(key []byte, msg []byte) error {
	slog.Debug("produce message to kafka", "topic", this.topic, "key", string(key), "msg", string(msg))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   key,
			Value: msg,
			Time:  time.Now(),
		},
	)
}
