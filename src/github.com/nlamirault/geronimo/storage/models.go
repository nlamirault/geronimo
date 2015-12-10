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

package storage

// User is the structure used for serializing/deserializing user in Elasticsearch.
type User struct {
	Login    string `json:"user"`
	Name     string `json:"name"`
	Company  string `json:"company"`
	Email    string `json:"email"`
	Location string `json:"location"`
}

// Repository is the structure used for serializing/deserializing repository in Elasticsearch.
type Repository struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Created          string `json:"created"`
	Language         string `json:"language"`
	ForksCount       int    `json:"fork_count"`
	StarsCount       int    `json:"star_count"`
	SubscribersCount int    `json:"subscriber_count"`
	WatchersCount    int    `json:"watcher_count"`
	OpenIssuesCount  int    `json:"open_issue_count"`
}
