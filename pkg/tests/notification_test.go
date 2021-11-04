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

func initNotificationState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, notificationIds *[]string, brokerIds *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.UnderscoreIdWrapper{}
		err := user1.Impersonate().PutJSON(
			config.NotifierUrl+"/notifications",
			map[string]interface{}{
				"title":   "1",
				"message": "1",
				"userId":  user1.GetUserId(),
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*notificationIds = append(*notificationIds, temp.Id)

		temp = ctrl.UnderscoreIdWrapper{}
		err = user1.Impersonate().PutJSON(
			config.NotifierUrl+"/notifications",
			map[string]interface{}{
				"title":   "2",
				"message": "2",
				"userId":  user1.GetUserId(),
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*notificationIds = append(*notificationIds, temp.Id)

		temp = ctrl.UnderscoreIdWrapper{}
		err = user2.Impersonate().PutJSON(
			config.NotifierUrl+"/notifications",
			map[string]interface{}{
				"title":   "3",
				"message": "3",
				"userId":  user2.GetUserId(),
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*notificationIds = append(*notificationIds, temp.Id)

		tempB := ctrl.IdWrapper{}
		err = user1.Impersonate().PostJSON(
			config.NotifierUrl+"/brokers",
			map[string]interface{}{
				"address": "1",
			}, &tempB)
		if err != nil {
			t.Error(err)
			return
		}
		*brokerIds = append(*brokerIds, tempB.Id)

		tempB = ctrl.IdWrapper{}
		err = user1.Impersonate().PostJSON(
			config.NotifierUrl+"/brokers",
			map[string]interface{}{
				"address": "2",
			}, &tempB)
		if err != nil {
			t.Error(err)
			return
		}
		*brokerIds = append(*brokerIds, tempB.Id)

		tempB = ctrl.IdWrapper{}
		err = user2.Impersonate().PostJSON(
			config.NotifierUrl+"/brokers",
			map[string]interface{}{
				"address": "3",
			}, &tempB)
		if err != nil {
			t.Error(err)
			return
		}
		*brokerIds = append(*brokerIds, tempB.Id)
	}
}

func checkNotificationState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, notificationIds []string, brokerIds []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(notificationIds) != 3 {
			t.Error(notificationIds)
			return
		}
		temp := ctrl.NotificationList{}
		err := user1.Impersonate().GetJSON(config.NotifierUrl+"/notifications?limit=0", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Notifications) != 0 {
			t.Error(temp)
			return
		}

		temp = ctrl.NotificationList{}
		err = user2.Impersonate().GetJSON(config.NotifierUrl+"/notifications?limit=0", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Notifications) != 1 {
			t.Error(temp)
			return
		}
		if temp.Notifications[0].Id != notificationIds[2] {
			t.Error(temp)
		}

		tempB := ctrl.BrokerList{}
		err = user1.Impersonate().GetJSON(config.NotifierUrl+"/brokers?limit=0", &tempB)
		if err != nil {
			t.Error(err)
			return
		}
		if len(tempB.Brokers) != 0 {
			t.Error(tempB)
			return
		}

		tempB = ctrl.BrokerList{}
		err = user2.Impersonate().GetJSON(config.NotifierUrl+"/brokers?limit=0", &tempB)
		if err != nil {
			t.Error(err)
			return
		}
		if len(tempB.Brokers) != 1 {
			t.Error(temp)
			return
		}
		if tempB.Brokers[0].Id != brokerIds[2] {
			t.Error(temp)
		}
	}
}
