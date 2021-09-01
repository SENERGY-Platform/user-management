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
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/ctrl"
	"testing"
)

func initWaitingRoomState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token) func(t *testing.T) {
	return func(t *testing.T) {
		err := user1.Impersonate().PutJSON(config.WaitingRoomUrl+"/devices", []map[string]interface{}{
			{
				"local_id":       "1",
				"name":           "1",
				"device_type_id": "test",
			},
			{
				"local_id":       "2",
				"name":           "2",
				"device_type_id": "test",
			},
		}, nil)
		if err != nil {
			t.Error(err)
			return
		}
		err = user2.Impersonate().PutJSON(config.WaitingRoomUrl+"/devices", []map[string]interface{}{
			{
				"local_id":       "3",
				"name":           "3",
				"device_type_id": "test",
			},
			{
				"local_id":       "4",
				"name":           "4",
				"device_type_id": "test",
			},
		}, nil)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func checkWaitingRoomState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.WaitingRoomListIdWrapper{}
		err := user1.Impersonate().GetJSON(config.WaitingRoomUrl+"/devices", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Result) != 0 {
			t.Error(temp)
		}

		temp = ctrl.WaitingRoomListIdWrapper{}
		err = user2.Impersonate().GetJSON(config.WaitingRoomUrl+"/devices", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Result) != 2 {
			t.Error(temp)
			return
		}
		if temp.Result[0].Id != "3" || temp.Result[1].Id != "4" {
			t.Error(temp)
		}
	}
}
