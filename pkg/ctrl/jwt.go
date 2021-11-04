/*
 * Copyright 2018 InfAI (CC SES)
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
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"github.com/golang-jwt/jwt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"net/url"

	"io/ioutil"
)

type JwtImpersonate struct {
	Token   string
	XUserId string
}

func (this JwtImpersonate) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	req.Header.Set("Content-Type", contentType)
	if this.XUserId != "" {
		req.Header.Set("X-UserId", this.XUserId)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		resp.Body.Close()
	}
	return
}

func (this JwtImpersonate) Put(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	req.Header.Set("Content-Type", contentType)
	if this.XUserId != "" {
		req.Header.Set("X-UserId", this.XUserId)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		resp.Body.Close()
	}
	return
}

func (this JwtImpersonate) Delete(url string, body interface{}) (resp *http.Response, err error) {
	var req *http.Request
	if body != nil {
		b := new(bytes.Buffer)
		err = json.NewEncoder(b).Encode(body)
		if err != nil {
			return
		}
		req, err = http.NewRequest("DELETE", url, b)
	} else {
		req, err = http.NewRequest("DELETE", url, nil)
		if err != nil {
			return nil, err
		}
	}
	req.Header.Set("Authorization", this.Token)
	if this.XUserId != "" {
		req.Header.Set("X-UserId", this.XUserId)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		resp.Body.Close()
	}
	return
}

func (this JwtImpersonate) DeleteWithBody(url string, body interface{}) (resp *http.Response, err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return
	}
	req, err := http.NewRequest("DELETE", url, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	req.Header.Set("Content-Type", "application/json")
	if this.XUserId != "" {
		req.Header.Set("X-UserId", this.XUserId)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		resp.Body.Close()
	}
	return
}

func (this JwtImpersonate) PostJSON(url string, body interface{}, result interface{}) (err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return
	}
	resp, err := this.Post(url, "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
	}
	return
}

func (this JwtImpersonate) PutJSON(url string, body interface{}, result interface{}) (err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return
	}
	resp, err := this.Put(url, "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
	}
	return
}

func (this JwtImpersonate) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", this.Token)
	if this.XUserId != "" {
		req.Header.Set("X-UserId", this.XUserId)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		resp.Body.Close()
	}
	return
}

func (this JwtImpersonate) GetJSON(url string, result interface{}) (err error) {
	resp, err := this.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(payload, result)
	if err != nil {
		log.Println("ERROR:", string(payload))
		debug.PrintStack()
	}
	return
}

type OpenidToken struct {
	AccessToken      string    `json:"access_token"`
	ExpiresIn        float64   `json:"expires_in"`
	RefreshExpiresIn float64   `json:"refresh_expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	RequestTime      time.Time `json:"-"`
}

var openid *OpenidToken

func EnsureAccess(conf configuration.Config) (token JwtImpersonate, err error) {
	if openid == nil {
		openid = &OpenidToken{}
	}
	duration := time.Now().Sub(openid.RequestTime).Seconds()

	if openid.AccessToken != "" && openid.ExpiresIn-conf.AuthExpirationTimeBuffer > duration {
		token = JwtImpersonate{Token: "Bearer " + openid.AccessToken}
		return
	}

	if openid.RefreshToken != "" && openid.RefreshExpiresIn-conf.AuthExpirationTimeBuffer < duration {
		log.Println("refresh token", openid.RefreshExpiresIn, duration)
		err = refreshOpenidToken(openid, conf)
		if err != nil {
			log.Println("WARNING: unable to use refreshtoken", err)
		} else {
			token = JwtImpersonate{Token: "Bearer " + openid.AccessToken}
			return
		}
	}

	log.Println("get new access token")
	err = getOpenidToken(openid, conf)
	if err != nil {
		log.Println("ERROR: unable to get new access token", err)
		openid = &OpenidToken{}
	}
	token = JwtImpersonate{Token: "Bearer " + openid.AccessToken}
	return
}

func getOpenidToken(token *OpenidToken, conf configuration.Config) (err error) {
	requesttime := time.Now()
	resp, err := http.PostForm(conf.KeycloakUrl+"/auth/realms/"+conf.KeycloakRealm+"/protocol/openid-connect/token", url.Values{
		"client_id":     {conf.AuthClientId},
		"client_secret": {conf.AuthClientSecret},
		"grant_type":    {"client_credentials"},
	})

	if err != nil {
		debug.PrintStack()
		log.Println("ERROR: getOpenidToken::PostForm()", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: getOpenidToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	err = json.NewDecoder(resp.Body).Decode(token)
	token.RequestTime = requesttime
	return
}

func refreshOpenidToken(token *OpenidToken, conf configuration.Config) (err error) {
	requesttime := time.Now()
	resp, err := http.PostForm(conf.KeycloakUrl+"/auth/realms/"+conf.KeycloakRealm+"/protocol/openid-connect/token", url.Values{
		"client_id":     {conf.AuthClientId},
		"client_secret": {conf.AuthClientSecret},
		"refresh_token": {token.RefreshToken},
		"grant_type":    {"refresh_token"},
	})

	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: refreshOpenidToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	err = json.NewDecoder(resp.Body).Decode(token)
	token.RequestTime = requesttime
	return
}

type Token struct {
	Token       string      `json:"-"`
	Sub         string      `json:"sub,omitempty"`
	RealmAccess RealmAccess `json:"realm_access,omitempty"`
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}

func (this *Token) IsAdmin() bool {
	return Contains(this.RealmAccess.Roles, "admin")
}

func (this *Token) GetUserId() string {
	return this.Sub
}

func (this *Token) Impersonate() JwtImpersonate {
	return JwtImpersonate{Token: this.Token, XUserId: this.Sub}
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type KeycloakClaims struct {
	RealmAccess RealmAccess `json:"realm_access"`
	jwt.StandardClaims
}

func CreateToken(issuer string, userId string) (token Token, err error) {
	return CreateTokenWithRoles(issuer, userId, []string{})
}

func CreateTokenWithRoles(issuer string, userId string, roles []string) (token Token, err error) {
	realmAccess := RealmAccess{Roles: roles}
	claims := KeycloakClaims{
		realmAccess,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
			Issuer:    issuer,
			Subject:   userId,
		},
	}

	jwtoken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	unsignedTokenString, err := jwtoken.SigningString()
	if err != nil {
		return token, err
	}
	tokenString := strings.Join([]string{unsignedTokenString, ""}, ".")
	token.Token = "Bearer " + tokenString
	token.Sub = userId
	token.RealmAccess = realmAccess
	return token, nil
}
