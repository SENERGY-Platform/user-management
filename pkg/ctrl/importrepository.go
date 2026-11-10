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
	"net/url"

	"github.com/SENERGY-Platform/user-management/pkg/configuration"
)

// DeleteImportRepositoryUser deletes the import types the user administers alone and removes the
// user's permissions from all other import types.
func DeleteImportRepositoryUser(userId string, conf configuration.Config) error {
	if conf.ImportRepositoryUrl == "" || conf.ImportRepositoryUrl == "-" {
		return nil
	}
	token, err := adminToken()
	if err != nil {
		return err
	}
	// Delete returns the response together with an error that contains the body for status >= 300
	resp, err := token.Delete(conf.ImportRepositoryUrl+"/users/"+url.PathEscape(userId), nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
