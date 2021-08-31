package ctrl

import (
	"errors"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"io"
	"net/url"
	"strconv"
)

func DeleteBrokerExportsUser(token Token, conf configuration.Config) error {
	if conf.BrokerExportsUrl == "" || conf.BrokerExportsUrl == "-" {
		return nil
	}

	loopLimit := 10000
	loopCount := 0
	for {
		ids, err := getBatchOfBrokerExportIds(token, conf, BatchSize, 0)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}

		for _, id := range ids {
			err = deleteBrokerExport(token, conf, id)
			if err != nil {
				return err
			}
		}

		loopCount = loopCount + 1
		if loopCount == loopLimit {
			return errors.New("DeleteBrokerExportUser() reach loop limit")
		}
	}

}

func deleteBrokerExport(token Token, conf configuration.Config, id string) error {
	resp, err := token.Impersonate().DeleteWithBody(conf.BrokerExportsUrl+"/instances/"+url.QueryEscape(id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New("deleteBrokerExport(): " + string(temp))
	}
	return nil
}

func getBatchOfBrokerExportIds(token Token, config configuration.Config, limit int, offset int) (ids []string, err error) {
	query := url.Values{}
	query.Add("limit", strconv.Itoa(limit))
	query.Add("offset", strconv.Itoa(offset))
	temp := ExportListIdWrapper{}
	err = token.Impersonate().GetJSON(config.BrokerExportsUrl+"/instances?"+query.Encode(), &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp.Instances {
		ids = append(ids, element.Id)
	}
	return ids, err
}
