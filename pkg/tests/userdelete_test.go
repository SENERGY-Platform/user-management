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
	importIds := []string{}
	brokerExportsIds := []string{}

	t.Run("init states", func(t *testing.T) {
		t.Run("init waiting room state", initWaitingRoomState(config, user1, user2))
		t.Run("init scheduler state", initSchedulerState(config, user1, user2, &scheduleIds))
		t.Run("init dashboard state", initDashboardState(config, user1, user2, &dashboardIds))
		t.Run("init imports state", initImportState(config, user1, user2, &importIds))
		t.Run("init broker exports state", initBrokerExportState(config, user1, user2, &brokerExportsIds))
	})

	time.Sleep(30 * time.Second)

	users := &kafka.Writer{
		Addr:        kafka.TCP(config.KafkaBootstrap),
		Topic:       config.UserTopic,
		MaxAttempts: 10,
		Logger:      log.New(os.Stdout, "[TEST-KAFKA-PRODUCER] ", 0),
	}

	t.Run("remove user1", func(t *testing.T) {
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

	time.Sleep(60 * time.Second)

	t.Run("check states after delete", func(t *testing.T) {
		t.Run("check waiting room state", checkWaitingRoomState(config, user1, user2))
		t.Run("check scheduler state", checkSchedulerState(config, user1, user2, scheduleIds))
		t.Run("check dashboard state", checkDashboardState(config, user1, user2, dashboardIds))
		t.Run("check imports state", checkImportsState(config, user1, user2, importIds))
		t.Run("check broker exports state", checkBrokerExportsState(config, user1, user2, brokerExportsIds))
	})

	t.Run("remove user2 for cleanup", func(t *testing.T) {
		cmd := ctrl.UserCommandMsg{
			Command: "DELETE",
			Id:      user2.GetUserId(),
		}
		message, err := json.Marshal(cmd)
		if err != nil {
			t.Error(err)
			return
		}
		err = users.WriteMessages(
			context.Background(),
			kafka.Message{
				Key:   []byte(user2.GetUserId()),
				Value: message,
				Time:  time.Now(),
			},
		)
		if err != nil {
			t.Error(err)
		}
	})
	time.Sleep(30 * time.Second)
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

func initImportState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.IdWrapper{}
		err := user1.Impersonate().PostJSON(
			config.ImportsDeploymentUrl+"/instances",
			map[string]interface{}{
				"name":           "1",
				"import_type_id": "1",
				"image":          "docker.io/library/hello-world",
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user1.Impersonate().PostJSON(
			config.ImportsDeploymentUrl+"/instances",
			map[string]interface{}{
				"name":           "2",
				"import_type_id": "2",
				"image":          "docker.io/library/hello-world",
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.IdWrapper{}
		err = user2.Impersonate().PostJSON(
			config.ImportsDeploymentUrl+"/instances",
			map[string]interface{}{
				"name":           "3",
				"import_type_id": "3",
				"image":          "docker.io/library/hello-world",
			}, &temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)
	}
}

func getBrokerExportElement(name string) map[string]interface{} {
	return map[string]interface{}{
		"name":       name,
		"FilterType": "deviceId",
		"Filter":     name,
	}
}

func initBrokerExportState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids *[]string) func(t *testing.T) {
	return func(t *testing.T) {
		temp := ctrl.ExportIdWrapper{}
		err := user1.Impersonate().PostJSON(
			config.BrokerExportsUrl+"/instances",
			getBrokerExportElement("1"),
			&temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.ExportIdWrapper{}
		err = user1.Impersonate().PostJSON(
			config.BrokerExportsUrl+"/instances",
			getBrokerExportElement("2"),
			&temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)

		temp = ctrl.ExportIdWrapper{}
		err = user2.Impersonate().PostJSON(
			config.BrokerExportsUrl+"/instances",
			getBrokerExportElement("3"),
			&temp)
		if err != nil {
			t.Error(err)
			return
		}
		*ids = append(*ids, temp.Id)
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
			return
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
			return
		}
		if temp[0].Id != ids[2] {
			t.Error(temp)
		}
	}
}

func checkImportsState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 3 {
			t.Error(ids)
		}
		temp := []ctrl.IdWrapper{}
		err := user1.Impersonate().GetJSON(config.ImportsDeploymentUrl+"/instances", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 0 {
			t.Error(temp)
		}

		temp = []ctrl.IdWrapper{}
		err = user2.Impersonate().GetJSON(config.ImportsDeploymentUrl+"/instances", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp) != 1 {
			t.Error(temp)
			return
		}
		if temp[0].Id != ids[2] {
			t.Error(temp)
		}
	}
}

func checkBrokerExportsState(config configuration.Config, user1 ctrl.Token, user2 ctrl.Token, ids []string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(ids) != 3 {
			t.Error(ids)
		}
		temp := ctrl.ExportListIdWrapper{}
		err := user1.Impersonate().GetJSON(config.BrokerExportsUrl+"/instances", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Instances) != 0 {
			t.Error(temp)
		}

		temp = ctrl.ExportListIdWrapper{}
		err = user2.Impersonate().GetJSON(config.BrokerExportsUrl+"/instances", &temp)
		if err != nil {
			t.Error(err)
			return
		}
		if len(temp.Instances) != 1 {
			t.Error(temp)
			return
		}
		if temp.Instances[0].Id != ids[2] {
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
			return
		}
		if temp[0].Id != ids[2] {
			t.Error(temp, ids)
		}
	}
}
