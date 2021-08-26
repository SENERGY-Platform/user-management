package ctrl

import (
	"errors"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"io"
	"net/url"
	"strconv"
)

var BatchSize = 100

func DeleteWaitingRoomUser(token Token, conf configuration.Config) error {
	if conf.WaitingRoomUrl == "" || conf.WaitingRoomUrl == "-" {
		return nil
	}

	loopLimit := 10000
	loopCount := 0
	for {
		ids, err := getBatchOfWaitingRoomDeviceIds(token, conf, BatchSize, 0)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}

		err = deleteBatchOfWaitingRoomDevices(token, conf, ids)
		if err != nil {
			return err
		}

		loopCount = loopCount + 1
		if loopCount == loopLimit {
			return errors.New("DeleteWaitingRoomUser() reach loop limit")
		}
	}

}

func deleteBatchOfWaitingRoomDevices(token Token, conf configuration.Config, ids []string) error {
	resp, err := token.Impersonate().DeleteWithBody(conf.WaitingRoomUrl+"/devices", ids)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New("deleteBatchOfWaitingRoomDevices(): " + string(temp))
	}
	return nil
}

func getBatchOfWaitingRoomDeviceIds(token Token, config configuration.Config, limit int, offset int) (ids []string, err error) {
	query := url.Values{}
	query.Add("limit", strconv.Itoa(limit))
	query.Add("offset", strconv.Itoa(offset))
	query.Add("show_hidden", "true")
	temp := WaitingRoomListIdWrapper{}
	err = token.Impersonate().GetJSON(config.WaitingRoomUrl+"/devices?"+query.Encode(), &temp)
	if err != nil {
		return ids, err
	}
	for _, element := range temp.Result {
		ids = append(ids, element.Id)
	}
	return ids, err
}
