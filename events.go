/*
 * Copyright 2018 InfAI (CC SES)
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

package main

import (
	"context"
	"github.com/SENERGY-Platform/user-management/kafka"
	"log"
	"sync"
	"time"

	"encoding/json"
	"errors"
)

type UserCommandMsg struct {
	Command string `json:"command"`
	Id      string `json:"id"`
}

type EventHandler struct {
	conf Config
	conn kafka.Interface
}

func InitEventConn(ctx context.Context, wg *sync.WaitGroup, conf Config) (handler *EventHandler, err error) {
	conn, err := kafka.Init(ctx, wg, conf.KafkaBootstrap, conf.ConsumerGroup, conf.Debug)
	if err != nil {
		log.Println("WARN: problem initializing kafka connection: ", err)
		log.Println("WARN: client will retry until successful")
	}
	handler = &EventHandler{
		conf: conf,
		conn: conn,
	}
	log.Println("init permissions handler")
	err = conn.Consume(conf.UserTopic, handler.handleUserCommand)
	if err != nil {
		log.Println("WARN: problem initializing kafka connection: ", err)
		log.Println("WARN: client will retry until successful")
		err = nil
	}
	return
}

func (handler *EventHandler) sendEvent(topic string, key string, command interface{}) error {
	payload, err := json.Marshal(command)
	if err != nil {
		log.Println("ERROR: event marshaling:", err)
		return err
	}
	return handler.conn.Publish(topic, key, payload)
}

func (handler *EventHandler) DeleteUser(id string) error {
	user, err := GetUserById(id, handler.conf)
	if err != nil {
		return err
	}
	if user.Id != id {
		return errors.New("no matching user found")
	}
	return handler.sendEvent(handler.conf.UserTopic, "DELETE_"+id, UserCommandMsg{
		Command: "DELETE",
		Id:      id,
	})
}

func (handler *EventHandler) handleUserCommand(_ string, msg []byte, _ time.Time) (err error) {
	log.Println(handler.conf.UserTopic, string(msg))
	command := UserCommandMsg{}
	err = json.Unmarshal(msg, &command)
	if err != nil {
		return
	}
	switch command.Command {
	case "DELETE":
		return DeleteKeycloakUser(command.Id, handler.conf)
	}
	return errors.New("unable to handle permission command: " + string(msg))
}
