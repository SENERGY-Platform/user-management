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
	"github.com/SENERGY-Platform/user-management/kafka"
	"log"

	"encoding/json"
	"errors"
)

var conn kafka.Interface

type UserCommandMsg struct {
	Command string `json:"command"`
	Id      string `json:"id"`
}

func InitEventConn() {
	var err error
	conn, err = kafka.Init(Config.ZookeeperUrl, Config.ConsumerGroup, Config.Debug)
	if err != nil {
		log.Fatal("ERROR: while initializing amqp connection", err)
	}

	log.Println("init permissions handler")
	err = conn.Consume(Config.UserTopic, handleUserCommand)
	if err != nil {
		log.Fatal("ERROR: while initializing perm consumer", err)
		return
	}
}

func StopEventConn() {
	conn.Close()
}

func sendEvent(topic string, key string, command interface{}) error {
	payload, err := json.Marshal(command)
	if err != nil {
		log.Println("ERROR: event marshaling:", err)
		return err
	}
	return conn.Publish(topic, key, payload)
}

func DeleteUser(id string) error {
	user, err := GetUserById(id)
	if err != nil {
		return err
	}
	if user.Id != id {
		return errors.New("no matching user found")
	}
	return sendEvent(Config.UserTopic, "DELETE_"+id, UserCommandMsg{
		Command: "DELETE",
		Id:      id,
	})
}

func handleUserCommand(msg []byte) (err error) {
	log.Println(Config.UserTopic, string(msg))
	command := UserCommandMsg{}
	err = json.Unmarshal(msg, &command)
	if err != nil {
		return
	}
	switch command.Command {
	case "DELETE":
		return DeleteKeycloakUser(command.Id)
	}
	return errors.New("unable to handle permission command: " + string(msg))
}
