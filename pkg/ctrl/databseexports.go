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
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"io"
	"net/url"
	"strconv"
)

func DeleteDatabaseExportsUser(token Token, conf configuration.Config) error {
	if conf.DatabaseExportsUrl == "" || conf.DatabaseExportsUrl == "-" {
		return nil
	}

	loopLimit := 10000
	loopCount := 0
	for {
		ids, err := getBatchOfDatabaseExportIds(token, conf, BatchSize, 0)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}

		err = deleteBatchOfDatabaseExports(token, conf, ids)
		if err != nil {
			return err
		}

		loopCount = loopCount + 1
		if loopCount == loopLimit {
			return errors.New("DeleteDatabaseExportUser() reach loop limit")
		}
	}

}

func deleteBatchOfDatabaseExports(token Token, conf configuration.Config, ids []string) error {
	if len(ids) > 0 {
		resp, err := token.Impersonate().DeleteWithBody(conf.DatabaseExportsUrl+"/instances", ids)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			temp, _ := io.ReadAll(resp.Body)
			return errors.New("deleteBatchOfDatabaseExports(): " + string(temp))
		}
	}
	return nil
}

func getBatchOfDatabaseExportIds(token Token, config configuration.Config, limit int, offset int) (ids []string, err error) {
	query := url.Values{}
	query.Add("limit", strconv.Itoa(limit))
	query.Add("offset", strconv.Itoa(offset))
	temp := ExportListIdWrapper{}
	err = token.Impersonate().GetJSON(config.DatabaseExportsUrl+"/instance?"+query.Encode(), &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp.Instances {
		ids = append(ids, element.Id)
	}
	return ids, err
}
