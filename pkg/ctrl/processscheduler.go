package ctrl

import (
	"errors"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"io"
	"net/url"
)

func DeleteProcessSchedulerUser(token Token, conf configuration.Config) error {
	if conf.ProcessSchedulerUrl == "" || conf.ProcessSchedulerUrl == "-" {
		return nil
	}
	ids, err := getProcessScheduleIds(token, conf)
	if err != nil {
		return err
	}
	for _, id := range ids {
		err = deleteProcessSchedule(token, conf, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteProcessSchedule(token Token, conf configuration.Config, id string) error {
	resp, err := token.Impersonate().Delete(conf.ProcessSchedulerUrl + "/schedules/" + url.QueryEscape(id))
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

func getProcessScheduleIds(token Token, config configuration.Config) (ids []string, err error) {
	temp := []IdWrapper{}
	err = token.Impersonate().GetJSON(config.ProcessSchedulerUrl+"/schedules", &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp {
		ids = append(ids, element.Id)
	}
	return ids, err
}
