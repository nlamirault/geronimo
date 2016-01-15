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
	"io/ioutil"
	"os"
	"testing"
)

func createConfiguration(t *testing.T, data []byte) *os.File {
	configFile, err := ioutil.TempFile("", "geronimo")
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(configFile.Name(), data, 0700)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	return configFile
}

func TestValidConfiguration(t *testing.T) {
	data := []byte(`
[nsq]
channel = "geronimo"
lookupd = "lookupd:4161"

[github]
api_token = "azerty2468"
user = "nlamirault"

[elasticsearch]
host = "localhost:9200"
`)
	configFile := createConfiguration(t, data)
	defer os.RemoveAll(configFile.Name())
	conf, err := Load(configFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if conf.ElasticSearch.Host != "localhost:9200" {
		t.Fatalf("Invalid Elasticsearch conf: %#v", conf)
	}
	if conf.Github.APIToken != "azerty2468" ||
		conf.Github.User != "nlamirault" {
		t.Fatalf("Invalid Github conf: %#v", conf)
	}
	if conf.NSQ.Channel != "geronimo" ||
		conf.NSQ.Lookupd != "lookupd:4161" {
		t.Fatalf("Invalid NSQ conf: %#v", conf)
	}

}
