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

func DeleteExportDatabasesUser(token Token, conf configuration.Config) error {
	if conf.DatabaseExportsUrl == "" || conf.DatabaseExportsUrl == "-" {
		return nil
	}

	loopLimit := 10000
	loopCount := 0
	offset := 0
	for {
		toBeRemoved, publicIds, err := getBatchOfExportDatabasesIds(token, conf, BatchSize, offset)
		if err != nil {
			return err
		}
		if len(toBeRemoved) == 0 && len(publicIds) == 0 {
			return nil
		}

		err = deleteBatchOfExportDatabases(token, conf, toBeRemoved)
		if err != nil {
			return err
		}

		offset = offset + len(publicIds)

		loopCount = loopCount + 1
		if loopCount == loopLimit {
			return errors.New("DeleteDatabaseExportUser() reach loop limit")
		}
	}

}

func deleteBatchOfExportDatabases(token Token, conf configuration.Config, ids []string) error {
	for _, id := range ids {
		resp, err := token.Impersonate().Delete(conf.DatabaseExportsUrl+"/databases/"+url.PathEscape(id), nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			temp, _ := io.ReadAll(resp.Body)
			return errors.New("deleteBatchOfExportDatabases(): " + string(temp))
		}
	}
	return nil
}

func getBatchOfExportDatabasesIds(token Token, config configuration.Config, limit int, offset int) (toBeRemoved []string, publicIds []string, err error) {
	query := url.Values{}
	query.Add("limit", strconv.Itoa(limit))
	query.Add("offset", strconv.Itoa(offset))
	temp := []ExportDatabase{}
	err = token.Impersonate().GetJSON(config.DatabaseExportsUrl+"/databases?"+query.Encode(), &temp)
	if err != nil {
		return toBeRemoved, publicIds, err
	}
	for _, element := range temp {
		if element.Public {
			publicIds = append(publicIds, element.ID)
		} else {
			toBeRemoved = append(toBeRemoved, element.ID)
		}
	}
	return toBeRemoved, publicIds, err
}

type ExportDatabaseRequest struct {
	Name          string `json:"Name" validate:"required"`
	Description   string `json:"Description"`
	Type          string `json:"Type" validate:"required"`
	Deployment    string `json:"deployment" validate:"required"`
	Url           string `json:"Url" validate:"required"`
	EwFilterTopic string `json:"EwFilterTopic" validate:"required"`
	Public        bool   `json:"Public"`
}

type ExportDatabase struct {
	ID            string `gorm:"primary_key;type:varchar(255);column:id"`
	Name          string `gorm:"type:varchar(255)"`
	Description   string `gorm:"type:varchar(255)"`
	Type          string `gorm:"type:varchar(255)"`
	Deployment    string `gorm:"type:varchar(255)"`
	Url           string `gorm:"type:varchar(255)"`
	EwFilterTopic string `gorm:"type:varchar(255)"`
	UserId        string `gorm:"type:varchar(255)"`
	Public        bool   `gorm:"type:bool;DEFAULT:false"`
}
