// Copyright (C) 2015 Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"encoding/json"
	//"fmt"
	//"io"
	"io/ioutil"
	"log"
	"net/http"
)

func DecodeResponse(resp *http.Response, v interface{}) error {
	log.Printf("[DEBUG] Decode response")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Body response: %d / %s",
		resp.StatusCode, string(body))
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}
	return nil
}

func GetResponseBody(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
