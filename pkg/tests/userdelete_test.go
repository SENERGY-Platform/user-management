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

package tests

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/user-management/pkg/api"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/ctrl"
	"github.com/SENERGY-Platform/user-management/pkg/tests/docker"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestSwagger(t *testing.T) {
	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}
	config.RemoveExportDatabaseMetadataOnUserDelete = true

	config.ServerPort, err = docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	defer log.Println("done waiting")
	defer wg.Wait()
	defer log.Println("wait for wg")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err = api.Start(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("swagger", func(t *testing.T) {
		resp, err := http.Get("http://localhost:" + config.ServerPort + "/doc")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			t.Error(resp.StatusCode)
		}
	})
}

func TestUserDelete(t *testing.T) {
	old := ctrl.BatchSize
	ctrl.BatchSize = 1
	defer func() {
		ctrl.BatchSize = old
	}()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}
	config.RemoveExportDatabaseMetadataOnUserDelete = true

	config.ServerPort, err = docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	defer log.Println("done waiting")
	defer wg.Wait()
	defer log.Println("wait for wg")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, getDeviceRepoCalls, err := docker.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = api.Start(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	user1, err := ctrl.CreateToken("test", "user1")
	if err != nil {
		t.Error(err)
		return
	}
	user2, err := ctrl.CreateToken("test", "user2")
	if err != nil {
		t.Error(err)
		return
	}

	scheduleIds := []string{}
	dashboardIds := []string{}
	importIds := []string{}
	brokerExportsIds := []string{}
	dbExportsIds := []string{}
	user2Databases := []string{}
	operatorIds := []string{}
	flowIds := []string{}
	flowEngineIds := []string{}
	notificationIds := []string{}
	brokerIds := []string{}
	var dbId string

	t.Run("init states", func(t *testing.T) {
		t.Run("init waiting room state", initWaitingRoomState(config, user1, user2))
		t.Run("init scheduler state", initSchedulerState(config, user1, user2, &scheduleIds))
		t.Run("init dashboard state", initDashboardState(config, user1, user2, &dashboardIds))
		t.Run("init imports state", initImportState(config, user1, user2, &importIds))
		t.Run("init broker exports state", initBrokerExportState(config, user1, user2, &brokerExportsIds))
		t.Run("init public export db", initPublicExportDatabase(config, user1, &dbId))
		user2Databases = append(user2Databases, dbId)
		t.Run("init exports databases", initExportDatabases(config, user1, user2, &user2Databases))
		t.Run("init database exports state", initDatabaseExportState(config, user1, user2, dbId, &dbExportsIds))
		t.Run("init operators state", initOperatorState(config, user1, user2, &operatorIds))
		t.Run("init flow state", initFlowState(config, user1, user2, &flowIds))
		t.Run("init flow engine state", initFlowEngineState(config, user1, user2, &flowEngineIds))
		t.Run("init notification state", initNotificationState(config, user1, user2, &notificationIds, &brokerIds))
	})

	time.Sleep(30 * time.Second)

	users := &kafka.Writer{
		Addr:        kafka.TCP(config.KafkaBootstrap),
		Topic:       config.UserTopic,
		MaxAttempts: 10,
		Logger:      log.New(os.Stdout, "[TEST-KAFKA-PRODUCER] ", 0),
	}

	t.Run("remove user1", func(t *testing.T) {
		cmd := ctrl.UserCommandMsg{
			Command: "DELETE",
			Id:      user1.GetUserId(),
		}
		message, err := json.Marshal(cmd)
		if err != nil {
			t.Error(err)
			return
		}
		err = users.WriteMessages(
			context.Background(),
			kafka.Message{
				Key:   []byte(user1.GetUserId()),
				Value: message,
				Time:  time.Now(),
			},
		)
		if err != nil {
			t.Error(err)
		}
	})

	time.Sleep(60 * time.Second)

	t.Run("check states after delete", func(t *testing.T) {
		t.Run("check waiting room state", checkWaitingRoomState(config, user1, user2))
		t.Run("check scheduler state", checkSchedulerState(config, user1, user2, scheduleIds))
		t.Run("check dashboard state", checkDashboardState(config, user1, user2, dashboardIds))
		t.Run("check imports state", checkImportsState(config, user1, user2, importIds))
		t.Run("check broker exports state", checkBrokerExportsState(config, user1, user2, brokerExportsIds))
		t.Run("check export databases", checkExportsDatabases(config, user1, user2, user2Databases))
		t.Run("check database exports state", checkDatabaseExportsState(config, user1, user2, dbExportsIds))
		t.Run("check operators state", checkOperatorState(config, user1, user2, operatorIds))
		t.Run("check flows state", checkFlowState(config, user1, user2, flowIds))
		t.Run("check flows engine state", checkFlowEngineState(config, user1, user2, flowEngineIds))
		t.Run("check notification state", checkNotificationState(config, user1, user2, notificationIds, brokerIds))
		t.Run("check device-repo call", func(t *testing.T) {
			if !slices.Contains(getDeviceRepoCalls(), "DELETE /users/"+url.PathEscape(user1.GetUserId())) {
				t.Errorf("calls = %#v", getDeviceRepoCalls())
			}
		})
	})
}
