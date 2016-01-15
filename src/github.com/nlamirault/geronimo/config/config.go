// Copyright (C) 2015, 2016 Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	//"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

// NSQConfig is the configuration for NSQ.
type NSQConfig struct {
	Topic   string `json:"topic"`
	Channel string `json:"channel"`
	Lookupd string `json:"lookup_address"`
}

// GithubConfig is the Github configuration
type GithubConfig struct {
	APIToken string `toml:"api_token"`
	User     string `toml:"user"`
}

// ElasticsearchConfig is the Elasticsearch configuration
type ElasticsearchConfig struct {
	Host string `toml:"host"`
}

// Configuration is the Geronimo configuration
type Configuration struct {
	NSQ           NSQConfig           `toml:"nsq"`
	Github        GithubConfig        `toml:"github"`
	ElasticSearch ElasticsearchConfig `toml:"elasticsearch"`
}

// Load read the configuration
func Load(filename string) (*Configuration, error) {
	var config Configuration
	log.Printf("[DEBUG] Load configuration from %s\n", filename)
	if _, err := toml.DecodeFile(filename, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
