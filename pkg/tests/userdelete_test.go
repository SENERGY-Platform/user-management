package tests

import (
	"context"
	"github.com/SENERGY-Platform/user-management/pkg/api"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/tests/docker"
	"sync"
	"testing"
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

	apictx, apicancel := context.WithCancel(ctx)
	wg2, err := api.Start(apictx, config)
	if err != nil {
		t.Error(err)
		apicancel()
		return
	}
	defer wg2.Wait()
	defer apicancel()
}
