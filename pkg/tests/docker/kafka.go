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

package docker

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Kafka(ctx context.Context, wg *sync.WaitGroup) (kafkaUrl string, err error) {
	kafkaport, err := GetFreePort()
	if err != nil {
		return kafkaUrl, err
	}
	provider, err := testcontainers.NewDockerProvider(testcontainers.DefaultNetwork("bridge"))
	if err != nil {
		return kafkaUrl, err
	}
	hostIp, err := provider.GetGatewayIP(ctx)
	if err != nil {
		return kafkaUrl, err
	}
	kafkaUrl = hostIp + ":" + kafkaport
	log.Println("host ip: ", hostIp)
	log.Println("host port: ", kafkaport)
	log.Println("kafkaUrl url: ", kafkaUrl)
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "apache/kafka:3.9.1",
			Tmpfs: map[string]string{},
			WaitingFor: wait.ForAll(
				wait.ForLog("INFO Awaiting socket connections on"),
				wait.ForListeningPort("9092/tcp"),
			),
			ExposedPorts:    []string{kafkaport + ":9092"},
			AlwaysPullImage: true,
			Env: map[string]string{
				"KAFKA_NODE_ID":                                  "1",
				"CLUSTER_ID":                                     "IrzuggcFT-mWom7mj7PgtA",
				"KAFKA_PROCESS_ROLES":                            "controller,broker",
				"KAFKA_LISTENERS":                                "BROKER://:9092,CONTROLLER://:9093",
				"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
				"KAFKA_INTER_BROKER_LISTENER_NAME":               "BROKER",
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "CONTROLLER:PLAINTEXT,BROKER:PLAINTEXT",
				"KAFKA_ADVERTISED_LISTENERS":                     "BROKER://" + kafkaUrl,
				"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@localhost:9093",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
				"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
				"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
				"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
				"KAFKA_NUM_PARTITIONS":                           "1",
			},
		},
		Started: true,
	})
	if err != nil {
		return kafkaUrl, err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("DEBUG: remove container kafka", c.Terminate(context.Background()))
	}()

	containerPort, err := c.MappedPort(ctx, "9092/tcp")
	if err != nil {
		return kafkaUrl, err
	}
	log.Println("KAFKA_TEST: container-port", containerPort, kafkaport)

	err = retry(1*time.Minute, func() error {
		return tryKafkaConn(kafkaUrl)
	})
	if err != nil {
		return kafkaUrl, err
	}

	return kafkaUrl, err
}

func tryKafkaConn(kafkaUrl string) error {
	log.Println("try kafka connection to " + kafkaUrl + "...")
	conn, err := kafka.Dial("tcp", kafkaUrl)
	if err != nil {
		log.Println(err)
		return err
	}
	defer conn.Close()
	brokers, err := conn.Brokers()
	if err != nil {
		log.Println(err)
		return err
	}
	if len(brokers) == 0 {
		err = errors.New("missing brokers")
		log.Println(err)
		return err
	}
	p, err := conn.ReadPartitions("__consumer_offsets")
	if err != nil {
		log.Println(err)
		return err
	}
	if len(p) == 0 {
		err = errors.New("__consumer_offsets not ready")
		return err
	}
	log.Println("kafka connection ok")
	return nil
}
