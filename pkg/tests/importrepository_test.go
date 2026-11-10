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

type importRepositoryMock struct {
	mux   sync.Mutex
	calls []string
}

func (this *importRepositoryMock) getCalls() []string {
	this.mux.Lock()
	defer this.mux.Unlock()
	return slices.Clone(this.calls)
}

// startImportRepositoryMock answers DELETE /users/{id} like import-repository: 403 without the
// admin role, 204 otherwise; with failing, 500
func startImportRepositoryMock(t *testing.T, failing bool) (string, *importRepositoryMock) {
	mock := &importRepositoryMock{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.mux.Lock()
		mock.calls = append(mock.calls, r.Method+" "+r.URL.Path)
		mock.mux.Unlock()
		token, err := api.GetParsedToken(r)
		if err != nil || !token.IsAdmin() {
			http.Error(w, "admin role required", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodDelete || !strings.HasPrefix(r.URL.Path, "/users/") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if failing {
			http.Error(w, "permissions-v2 unavailable", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	return server.URL, mock
}

func TestImportRepositoryUserDeletion(t *testing.T) {
	t.Run("delete", func(t *testing.T) {
		repoUrl, repo := startImportRepositoryMock(t, false)
		err := ctrl.DeleteImportRepositoryUser("user1", configuration.Config{ImportRepositoryUrl: repoUrl})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(repo.getCalls(), []string{"DELETE /users/user1"}) {
			t.Errorf("calls = %#v", repo.getCalls())
		}
	})

	t.Run("error is returned", func(t *testing.T) {
		repoUrl, _ := startImportRepositoryMock(t, true)
		err := ctrl.DeleteImportRepositoryUser("user1", configuration.Config{ImportRepositoryUrl: repoUrl})
		if err == nil || !strings.Contains(err.Error(), "permissions-v2 unavailable") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		for _, repoUrl := range []string{"", "-"} {
			if err := ctrl.DeleteImportRepositoryUser("user1", configuration.Config{ImportRepositoryUrl: repoUrl}); err != nil {
				t.Error(err)
			}
		}
	})
}
