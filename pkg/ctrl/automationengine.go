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

package ctrl

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SENERGY-Platform/user-management/pkg/configuration"
)

// ErrUserDeletionBlocked is returned when the automation-engine rejects the deletion of a user,
// because resources of other owners depend on resources of the user and their namespace does not
// allow to delete them with the user.
var ErrUserDeletionBlocked = errors.New("user deletion blocked by automation-engine")

// AutomationEngineUserDependents is the dry run result of GET /users/{id}/dependents.
type AutomationEngineUserDependents struct {
	UserId     string                      `json:"user_id"`
	Owned      map[string]int              `json:"owned"`
	Dependents []AutomationEngineDependent `json:"dependents"`
	Blocked    bool                        `json:"blocked"`
}

type AutomationEngineDependent struct {
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Id        string `json:"id"`
	Owner     string `json:"owner"`
	Reason    string `json:"reason"`
}

func automationEngineEnabled(conf configuration.Config) bool {
	return conf.AutomationEngineUrl != "" && conf.AutomationEngineUrl != "-"
}

// the engine accepts the user endpoints only with the admin role
func automationEngineToken() (JwtImpersonate, error) {
	token, err := CreateTokenWithRoles("users-service", "user-management", []string{"admin"})
	if err != nil {
		return JwtImpersonate{}, err
	}
	return token.Impersonate(), nil
}

// CheckAutomationEngineUser calls the dry run of the user deletion and returns ErrUserDeletionBlocked
// if the engine would reject it. It deletes nothing.
func CheckAutomationEngineUser(userId string, conf configuration.Config) error {
	if !automationEngineEnabled(conf) {
		return nil
	}
	token, err := automationEngineToken()
	if err != nil {
		return err
	}
	deps := AutomationEngineUserDependents{}
	err = token.GetJSON(conf.AutomationEngineUrl+"/users/"+url.PathEscape(userId)+"/dependents", &deps)
	if err != nil {
		return err
	}
	if deps.Blocked {
		list := make([]string, 0, len(deps.Dependents))
		for _, d := range deps.Dependents {
			list = append(list, d.Namespace+"/"+d.Kind+"/"+d.Id+" (owner "+d.Owner+": "+d.Reason+")")
		}
		return fmt.Errorf("%w: resources of other owners depend on the user: %v", ErrUserDeletionBlocked, strings.Join(list, ", "))
	}
	return nil
}

// DeleteAutomationEngineUser deletes the processes, deployments, instances, schedules and stored
// values of the user in the automation-engine.
func DeleteAutomationEngineUser(userId string, conf configuration.Config) error {
	if !automationEngineEnabled(conf) {
		return nil
	}
	token, err := automationEngineToken()
	if err != nil {
		return err
	}
	// Delete returns the response together with an error that contains the body for status >= 300
	resp, err := token.Delete(conf.AutomationEngineUrl+"/users/"+url.PathEscape(userId), nil)
	if resp != nil && resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("%w: %v", ErrUserDeletionBlocked, err)
	}
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
