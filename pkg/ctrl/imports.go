package ctrl

import (
	"errors"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"io"
	"net/url"
	"strconv"
)

func DeleteImportsUser(token Token, conf configuration.Config) error {
	if conf.ImportsDeploymentUrl == "" || conf.ImportsDeploymentUrl == "-" {
		return nil
	}

	loopLimit := 10000
	loopCount := 0
	for {
		ids, err := getBatchOfImportIds(token, conf, BatchSize, 0)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}

		for _, id := range ids {
			err = deleteImport(token, conf, id)
			if err != nil {
				return err
			}
		}

		loopCount = loopCount + 1
		if loopCount == loopLimit {
			return errors.New("DeleteImportUser() reach loop limit")
		}
	}

}

func deleteImport(token Token, conf configuration.Config, id string) error {
	resp, err := token.Impersonate().Delete(conf.ImportsDeploymentUrl + "/instances/" + url.QueryEscape(id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New("deleteImport(): " + string(temp))
	}
	return nil
}

func getBatchOfImportIds(token Token, config configuration.Config, limit int, offset int) (ids []string, err error) {
	query := url.Values{}
	query.Add("limit", strconv.Itoa(limit))
	query.Add("offset", strconv.Itoa(offset))
	temp := []IdWrapper{}
	err = token.Impersonate().GetJSON(config.ImportsDeploymentUrl+"/instances?"+query.Encode(), &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp {
		ids = append(ids, element.Id)
	}
	return ids, err
}
