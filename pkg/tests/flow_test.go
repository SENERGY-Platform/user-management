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
	"slices"
	"strings"
	"testing"

	"github.com/SENERGY-Platform/analytics-flow-repo-v2/client"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/ctrl"
)

type Flow struct {
	Id   string `json:"_id"`
	Name string `json:"name"`
}
type FlowListResult struct {
	Flows []Flow `json:"flows"`
}

func initFlowState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		err := user1.Impersonate().PutJSON(
			config.AnalyticsFlowRepoUrl+"/flow/",
			map[string]interface{}{
				"name": "1",
				"model": map[string]interface{}{
					"cells": []interface{}{},
				},
			}, nil)
		if err != nil {
			t.Error(err)
			return
		}
		err = user1.Impersonate().PutJSON(
			config.AnalyticsFlowRepoUrl+"/flow/",
			map[string]interface{}{
				"name": "2",
				"model": map[string]interface{}{
					"cells": []interface{}{},
				},
			}, nil)
		if err != nil {
			t.Error(err)
			return
		}

		list := []Flow{}
		temp := FlowListResult{}
		err = user1.Impersonate().GetJSON(config.AnalyticsFlowRepoUrl+"/flow", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		for _, item := range temp.Flows {
			list = append(list, item)
		}
		if len(list) != 2 {
			t.Error(len(list))
		}

		err = user2.Impersonate().PutJSON(
			config.AnalyticsFlowRepoUrl+"/flow/",
			map[string]interface{}{
				"name": "3",
				"model": map[string]interface{}{
					"cells": []interface{}{},
				},
				"share": map[string]interface{}{
					"list": true,
				},
			}, nil)
		if err != nil {
			t.Error(err)
			return
		}
		err = user2.Impersonate().PutJSON(
			config.AnalyticsFlowRepoUrl+"/flow/",
			map[string]interface{}{
				"name": "4",
				"model": map[string]interface{}{
					"cells": []interface{}{},
				},
				"pub": false,
			}, nil)
		if err != nil {
			t.Error(err)
			return
		}

		temp = FlowListResult{}
		err = user2.Impersonate().GetJSON(config.AnalyticsFlowRepoUrl+"/flow", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Flows) != 2 {
			t.Error(len(temp.Flows))
		}
		for _, item := range temp.Flows {
			list = append(list, item)
		}

		slices.SortFunc(list, func(a, b Flow) int {
			return strings.Compare(a.Name, b.Name)
		})

		if len(list) != 4 {
			t.Error(len(list))
		}

		for _, item := range list {
			*ids = append(*ids, item.Id)
		}

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
		//one public flow from user2, but the pub field has lost its relevance for permissions
		if len(temp.Flows) != 0 {
			t.Error(len(temp.Flows), temp)
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
		if temp.Flows[0].Id.Hex() != ids[2] {
			t.Errorf("\nexp=%#v\nact=%#v\nids=%#v\nresult=%#v\n", ids[2], temp.Flows[0].Id.Hex(), ids, temp)
		}
		if temp.Flows[1].Id.Hex() != ids[3] {
			t.Errorf("\nexp=%#v\nact=%#v\nids=%#v\nresult=%#v\n", ids[3], temp.Flows[1].Id.Hex(), ids, temp)
		}
	}
}
