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

package github

import (
	"log"
	"net/http"

	gh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// NewClient creates a new Github client.
func NewClient(token string) *gh.Client {
	var tc *http.Client
	if token != "" {
		log.Printf("[DEBUG] Github token: %s", token)
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token,
		})
		tc = oauth2.NewClient(oauth2.NoContext, ts)
	}
	return gh.NewClient(tc)
}
