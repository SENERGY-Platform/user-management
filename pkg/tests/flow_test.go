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
	"github.com/SENERGY-Platform/analytics-flow-repo-v2/client"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/ctrl"
	"testing"
)

func initFlowState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.UnderscoreIdWrapper{}
		err := user1.Impersonate().PutJSON(
			config.AnalyticsFlowRepoUrl+"/flow",
			map[string]interface{}{
				"name": "1",
				"model": map[string]interface{}{
					"cells": []interface{}{},
				},
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.UnderscoreIdWrapper{}
		err = user1.Impersonate().PutJSON(
			config.AnalyticsFlowRepoUrl+"/flow",
			map[string]interface{}{
				"name": "2",
				"model": map[string]interface{}{
					"cells": []interface{}{},
				},
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.UnderscoreIdWrapper{}
		err = user2.Impersonate().PutJSON(
			config.AnalyticsFlowRepoUrl+"/flow",
			map[string]interface{}{
				"name": "3",
				"model": map[string]interface{}{
					"cells": []interface{}{},
				},
				"share": map[string]interface{}{
					"list": true,
				},
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)
		temp = ctrl.UnderscoreIdWrapper{}
		err = user2.Impersonate().PutJSON(
			config.AnalyticsFlowRepoUrl+"/flow",
			map[string]interface{}{
				"name": "4",
				"model": map[string]interface{}{
					"cells": []interface{}{},
				},
				"pub": false,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)
	}
}

func checkFlowState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 4 {
			t.Error(ids)
			return
		}
		temp, _, err := client.NewClient(config.AnalyticsFlowRepoUrl).GetFlows(user1.Token, user1.GetUserId())
		if err != nil {
			t.Error(err)
			return
		}
		//one public flow from user2
		if len(temp.Flows) != 1 {
			t.Error(temp)
			return
		}
		if temp.Flows[0].Id.String() != ids[2] {
			t.Error(temp)
		}

		temp, _, err = client.NewClient(config.AnalyticsFlowRepoUrl).GetFlows(user2.Token, user2.GetUserId())
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Flows) != 2 {
			t.Error(temp)
			return
		}
		if temp.Flows[0].Id.String() != ids[2] {
			t.Error(temp)
		}
		if temp.Flows[1].Id.String() != ids[3] {
			t.Error(temp)
		}
	}
}
