/*
 * Copyright 2026 InfAI (CC SES)
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

package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/user-management/pkg/api"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/SENERGY-Platform/user-management/pkg/ctrl"
)

type automationEngineMock struct {
	mux   sync.Mutex
	calls []string
}

func (this *automationEngineMock) getCalls() []string {
	this.mux.Lock()
	defer this.mux.Unlock()
	return slices.Clone(this.calls)
}

// startAutomationEngineMock answers the user endpoints of the automation-engine; with blocked, the
// dry run reports a foreign dependent and the deletion answers 409 like the engine does
func startAutomationEngineMock(t *testing.T, blocked bool) (string, *automationEngineMock) {
	mock := &automationEngineMock{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.mux.Lock()
		mock.calls = append(mock.calls, r.Method+" "+r.URL.Path)
		mock.mux.Unlock()
		token, err := api.GetParsedToken(r)
		if err != nil || !token.IsAdmin() {
			http.Error(w, "admin role required", http.StatusForbidden)
			return
		}
		userId := strings.Split(strings.TrimPrefix(r.URL.Path, "/users/"), "/")[0]
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/dependents"):
			deps := ctrl.AutomationEngineUserDependents{UserId: userId, Owned: map[string]int{"processes_models": 1}}
			if blocked {
				deps.Blocked = true
				deps.Dependents = []ctrl.AutomationEngineDependent{{Namespace: "processes", Kind: "deployments", Id: "d1", Owner: "user2", Reason: "uses model m1"}}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(deps)
		case r.Method == http.MethodDelete:
			if blocked {
				http.Error(w, "user "+userId+" owns resources that resources of other owners depend on", http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server.URL, mock
}

func TestAutomationEngineUserDeletion(t *testing.T) {
	t.Run("dry run not blocked", func(t *testing.T) {
		engineUrl, engine := startAutomationEngineMock(t, false)
		err := ctrl.CheckAutomationEngineUser("user1", configuration.Config{AutomationEngineUrl: engineUrl})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(engine.getCalls(), []string{"GET /users/user1/dependents"}) {
			t.Errorf("calls = %#v", engine.getCalls())
		}
	})

	t.Run("dry run blocked", func(t *testing.T) {
		engineUrl, engine := startAutomationEngineMock(t, true)
		err := ctrl.CheckAutomationEngineUser("user1", configuration.Config{AutomationEngineUrl: engineUrl})
		if !errors.Is(err, ctrl.ErrUserDeletionBlocked) {
			t.Fatalf("err = %v", err)
		}
		if !strings.Contains(err.Error(), "processes/deployments/d1") {
			t.Errorf("err does not name the dependent: %v", err)
		}
		if !slices.Equal(engine.getCalls(), []string{"GET /users/user1/dependents"}) {
			t.Errorf("calls = %#v", engine.getCalls())
		}
	})

	t.Run("delete", func(t *testing.T) {
		engineUrl, engine := startAutomationEngineMock(t, false)
		err := ctrl.DeleteAutomationEngineUser("user1", configuration.Config{AutomationEngineUrl: engineUrl})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(engine.getCalls(), []string{"DELETE /users/user1"}) {
			t.Errorf("calls = %#v", engine.getCalls())
		}
	})

	t.Run("delete blocked", func(t *testing.T) {
		engineUrl, _ := startAutomationEngineMock(t, true)
		err := ctrl.DeleteAutomationEngineUser("user1", configuration.Config{AutomationEngineUrl: engineUrl})
		if !errors.Is(err, ctrl.ErrUserDeletionBlocked) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("blocked deletion leaves other services untouched", func(t *testing.T) {
		engineUrl, engine := startAutomationEngineMock(t, true)
		deviceRepoCalls := []string{}
		deviceRepo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deviceRepoCalls = append(deviceRepoCalls, r.Method+" "+r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer deviceRepo.Close()
		err := ctrl.DeleteUser("user1", configuration.Config{AutomationEngineUrl: engineUrl, DeviceRepositoryUrl: deviceRepo.URL})
		if !errors.Is(err, ctrl.ErrUserDeletionBlocked) {
			t.Fatalf("err = %v", err)
		}
		if len(deviceRepoCalls) != 0 {
			t.Errorf("device-repository calls = %#v", deviceRepoCalls)
		}
		if !slices.Equal(engine.getCalls(), []string{"DELETE /users/user1"}) {
			t.Errorf("calls = %#v", engine.getCalls())
		}
	})

	t.Run("disabled", func(t *testing.T) {
		for _, engineUrl := range []string{"", "-"} {
			conf := configuration.Config{AutomationEngineUrl: engineUrl}
			if err := ctrl.CheckAutomationEngineUser("user1", conf); err != nil {
				t.Error(err)
			}
			if err := ctrl.DeleteAutomationEngineUser("user1", conf); err != nil {
				t.Error(err)
			}
		}
	})
}
