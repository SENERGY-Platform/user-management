/*
 * Copyright 2021 InfAI (CC SES)
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

package api

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"net/http"
	"strings"
)

func GetAuthToken(req *http.Request) string {
	return req.Header.Get("Authorization")
}

func GetParsedToken(req *http.Request) (token Token, err error) {
	return parse(GetAuthToken(req))
}

type Token struct {
	Token       string              `json:"-"`
	Sub         string              `json:"sub,omitempty"`
	RealmAccess map[string][]string `json:"realm_access,omitempty"`
}

func (this *Token) String() string {
	return this.Token
}

func (this *Token) Jwt() string {
	return this.Token
}

func (this *Token) Valid() error {
	if this.Sub == "" {
		return errors.New("missing subject")
	}
	return nil
}

func parse(token string) (claims Token, err error) {
	orig := token
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}
	_, _, err = new(jwt.Parser).ParseUnverified(token, &claims)
	if err == nil {
		claims.Token = orig
	}
	return
}

func (this *Token) IsAdmin() bool {
	return contains(this.RealmAccess["roles"], "admin")
}

func (this *Token) GetUserId() string {
	return this.Sub
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
