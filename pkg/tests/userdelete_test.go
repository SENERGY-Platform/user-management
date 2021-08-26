package tests

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/user-management/pkg/api"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/ctrl"
	"github.com/SENERGY-Platform/user-management/pkg/tests/docker"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestUserDelete(t *testing.T) {
	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}

	config.ServerPort, err = docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err = docker.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = api.Start(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	/*
		defer log.Println("api closed")
		defer wg2.Wait()
		defer log.Println("wait for api close")
		defer apicancel()
		defer log.Println("cancel api")
	*/

	user1, err := ctrl.CreateToken("test", "user1")
	if err != nil {
		t.Error(err)
		return
	}
	user2, err := ctrl.CreateToken("test", "user2")
	if err != nil {
		t.Error(err)
		return
	}

	scheduleIds := []string{}
	dashboardIds := []string{}

	t.Run("init states", func(t *testing.T) {
		t.Run("init waiting room state", initWaitingRoomState(config, user1, user2))
		t.Run("init scheduler state", initSchedulerState(config, user1, user2, &scheduleIds))
		t.Run("init dashboard state", initDashboardState(config, user1, user2, &dashboardIds))
	})

	t.Run("remove user1", func(t *testing.T) {
		users := &kafka.Writer{
			Addr:        kafka.TCP(config.KafkaBootstrap),
			Topic:       config.UserTopic,
			MaxAttempts: 10,
			Logger:      log.New(os.Stdout, "[TEST-KAFKA-PRODUCER] ", 0),
		}
		cmd := ctrl.UserCommandMsg{
			Command: "DELETE",
			Id:      user1.GetUserId(),
		}
		message, err := json.Marshal(cmd)
		if err != nil {
			t.Error(err)
			return
		}
		err = users.WriteMessages(
			context.Background(),
			kafka.Message{
				Key:   []byte(user1.GetUserId()),
				Value: message,
				Time:  time.Now(),
			},
		)
		if err != nil {
			t.Error(err)
		}
	})

	time.Sleep(10 * time.Second)

	t.Run("check states after delete", func(t *testing.T) {
		t.Run("check waiting room state", checkWaitingRoomState(config, user1, user2))
		t.Run("check scheduler state", checkSchedulerState(config, user1, user2, scheduleIds))
		t.Run("check dashboard state", checkDashboardState(config, user1, user2, dashboardIds))
	})
}

func initWaitingRoomState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token) func(t *testing.T) {
	return func(t *testing.T) {
		err := user1.Impersonate().PutJSON(config.WaitingRoomUrl+"/devices", []map[string]interface{}{
			{
				"local_id":       "1",
				"name":           "1",
				"device_type_id": "test",
			},
			{
				"local_id":       "2",
				"name":           "2",
				"device_type_id": "test",
			},
		}, nil)
		if err != nil {
			t.Error(err)
			return
		}
		err = user2.Impersonate().PutJSON(config.WaitingRoomUrl+"/devices", []map[string]interface{}{
			{
				"local_id":       "3",
				"name":           "3",
				"device_type_id": "test",
			},
			{
				"local_id":       "4",
				"name":           "4",
				"device_type_id": "test",
			},
		}, nil)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func initSchedulerState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, schedulerIds *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.IdWrapper{}
		err := user1.Impersonate().PostJSON(
			config.ProcessSchedulerUrl+"/schedules",
			map[string]interface{}{
				"cron":                  "* * * ? *",
				"process_deployment_id": "foo1",
				"disabled":              true,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*schedulerIds = append(*schedulerIds, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user1.Impersonate().PostJSON(
			config.ProcessSchedulerUrl+"/schedules",
			map[string]interface{}{
				"cron":                  "* * * ? *",
				"process_deployment_id": "foo2",
				"disabled":              true,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*schedulerIds = append(*schedulerIds, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user2.Impersonate().PostJSON(
			config.ProcessSchedulerUrl+"/schedules",
			map[string]interface{}{
				"cron":                  "* * * ? *",
				"process_deployment_id": "foo3",
				"disabled":              true,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*schedulerIds = append(*schedulerIds, temp.Id)
	}
}

func initDashboardState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, dashboardIds *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.IdWrapper{}
		err := user1.Impersonate().PutJSON(
			config.DashboardServiceUrl+"/dashboard",
			map[string]interface{}{
				"name":  "foo1",
				"index": 0,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*dashboardIds = append(*dashboardIds, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user1.Impersonate().PutJSON(
			config.DashboardServiceUrl+"/dashboard",
			map[string]interface{}{
				"name":  "foo2",
				"index": 1,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*dashboardIds = append(*dashboardIds, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user2.Impersonate().PutJSON(
			config.DashboardServiceUrl+"/dashboard",
			map[string]interface{}{
				"name":  "foo3",
				"index": 0,
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*dashboardIds = append(*dashboardIds, temp.Id)

		user1Dashboards := []ctrl.IdWrapper{}
		err = user1.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboards", &user1Dashboards)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(user1Dashboards, []ctrl.IdWrapper{
			{Id: (*dashboardIds)[0]},
			{Id: (*dashboardIds)[1]},
		}) {
			t.Error(user1Dashboards, dashboardIds)
		}

		user2Dashboards := []ctrl.IdWrapper{}
		err = user2.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboards", &user2Dashboards)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(user2Dashboards, []ctrl.IdWrapper{
			{Id: (*dashboardIds)[2]},
		}) {
			t.Error(user2Dashboards, dashboardIds)
		}
	}
}

func checkWaitingRoomState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.WaitingRoomListIdWrapper{}
		err := user1.Impersonate().GetJSON(config.WaitingRoomUrl+"/devices", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Result) != 0 {
			t.Error(temp)
		}

		temp = ctrl.WaitingRoomListIdWrapper{}
		err = user2.Impersonate().GetJSON(config.WaitingRoomUrl+"/devices", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Result) != 2 {
			t.Error(temp)
		}
		if temp.Result[0].Id != "3" || temp.Result[1].Id != "4" {
			t.Error(temp)
		}
	}
}

func checkSchedulerState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 3 {
			t.Error(ids)
		}
		temp := []ctrl.IdWrapper{}
		err := user1.Impersonate().GetJSON(config.ProcessSchedulerUrl+"/schedules", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 0 {
			t.Error(temp)
		}

		temp = []ctrl.IdWrapper{}
		err = user2.Impersonate().GetJSON(config.ProcessSchedulerUrl+"/schedules", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 1 {
			t.Error(temp)
		}
		if temp[0].Id != ids[2] {
			t.Error(temp)
		}
	}
}

func checkDashboardState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 3 {
			t.Error(ids)
		}
		temp := []ctrl.IdWrapper{}
		err := user1.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboards", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 1 {
			t.Error(temp)
		}
		if ctrl.Contains(ids, temp[0].Id) {
			t.Error(temp, ids)
		}

		temp = []ctrl.IdWrapper{}
		err = user2.Impersonate().GetJSON(config.DashboardServiceUrl+"/dashboards", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 1 {
			t.Error(temp)
		}
		if temp[0].Id != ids[2] {
			t.Error(temp, ids)
		}
	}
}
