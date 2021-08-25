package ctrl

import (
	"errors"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"io"
	"net/url"
)

func DeleteDashboardUser(token Token, conf configuration.Config) error {
	if conf.DashboardServiceUrl == "" || conf.DashboardServiceUrl == "-" {
		return nil
	}
	ids, err := getDashboardIds(token, conf)
	if err != nil {
		return err
	}
	for _, id := range ids {
		err = deleteDashboard(token, conf, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteDashboard(token Token, conf configuration.Config, id string) error {
	resp, err := token.Impersonate().Delete(conf.DashboardServiceUrl + "/dashboard/" + url.QueryEscape(id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New("deleteProcessSchedule(): " + string(temp))
	}
	return nil
}

func getDashboardIds(token Token, config configuration.Config) (ids []string, err error) {
	temp := []IdWrapper{}
	err = token.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboard", &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp {
		ids = append(ids, element.Id)
	}
	return ids, err
}
