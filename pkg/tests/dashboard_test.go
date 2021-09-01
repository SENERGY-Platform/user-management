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
	"reflect"
	"testing"
)

func initDashboardState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, dashboardIds *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.IdWrapper{}
		err := user1.Impersonate().PutJSON(
			config.DashboardServiceUrl+"/dashboard",
			map[string]interface{}{
				"name":  "foo1",
				"index": 0,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*dashboardIds = append(*dashboardIds, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user1.Impersonate().PutJSON(
			config.DashboardServiceUrl+"/dashboard",
			map[string]interface{}{
				"name":  "foo2",
				"index": 1,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*dashboardIds = append(*dashboardIds, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user2.Impersonate().PutJSON(
			config.DashboardServiceUrl+"/dashboard",
			map[string]interface{}{
				"name":  "foo3",
				"index": 0,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*dashboardIds = append(*dashboardIds, temp.Id)

		user1Dashboards := []ctrl.IdWrapper{}
		err = user1.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboards", &user1Dashboards)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(user1Dashboards, []ctrl.IdWrapper{
			{Id: (*dashboardIds)[0]},
			{Id: (*dashboardIds)[1]},
		}) {
			t.Error(user1Dashboards, dashboardIds)
		}

		user2Dashboards := []ctrl.IdWrapper{}
		err = user2.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboards", &user2Dashboards)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(user2Dashboards, []ctrl.IdWrapper{
			{Id: (*dashboardIds)[2]},
		}) {
			t.Error(user2Dashboards, dashboardIds)
		}
	}
}

func checkDashboardState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 3 {
			t.Error(ids)
		}
		temp := []ctrl.IdWrapper{}
		err := user1.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboards", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 1 {
			t.Error(temp)
		}
		if ctrl.Contains(ids, temp[0].Id) {
			t.Error(temp, ids)
		}

		temp = []ctrl.IdWrapper{}
		err = user2.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboards", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 1 {
			t.Error(temp)
			return
		}
		if temp[0].Id != ids[2] {
			t.Error(temp, ids)
		}
	}
}
