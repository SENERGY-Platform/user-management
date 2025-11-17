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

package ctrl

import (
	"errors"
	"github.com/SENERGY-Platform/analytics-pipeline/client"
	"github.com/SENERGY-Platform/analytics-pipeline/lib"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
)

func DeleteAnalyticsFlowEngineUser(token Token, conf configuration.Config) error {
	if conf.AnalyticsFlowEngineUrl == "" || conf.AnalyticsFlowEngineUrl == "-" {
		return nil
	}
	ids, err := getAnalyticsFlowEngineIds(token, conf)
	if err != nil {
		return err
	}
	for _, id := range ids {
		err = deleteAnalyticsFlowEngine(token, conf, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteAnalyticsFlowEngine(token Token, conf configuration.Config, id string) error {
	r, err, code := client.NewClient(conf.AnalyticsPipelineUrl).DeletePipeline(token.Token, token.GetUserId(), id)
	if err != nil || code >= 300 {
		return errors.New("deleteAnalyticsFlow(): " + r.Message)
	}
	return nil
}

func getAnalyticsFlowEngineIds(token Token, config configuration.Config) (ids []string, err error) {
	temp := []IdWrapper{}
	limit := 1000
	first := true
	c := client.NewClient(config.AnalyticsPipelineUrl)
	var pipelines lib.PipelinesResponse
	for first || len(pipelines.Data) == limit {
		first = false
		pipelines, err, _ = c.GetPipelines(token.Token, token.GetUserId(), limit, len(temp), "name", true)
		if err != nil {
			return ids, err
		}
		for _, p := range pipelines.Data {
			temp = append(temp, IdWrapper{Id: p.Id})
		}
	}
	return ids, err
}
