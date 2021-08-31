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
	resp, err := token.Impersonate().DeleteWithBody(conf.DatabaseExportsUrl+"/instances", ids)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New("deleteBatchOfDatabaseExports(): " + string(temp))
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
