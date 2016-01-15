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

package storage

import (
	"log"

	"gopkg.in/olivere/elastic.v3"
)

// NewClient creates a new Elasticsearch client.
func NewClient(uri string) (*elastic.Client, error) {
	return elastic.NewClient(elastic.SetURL(uri))
}

func CreateIndex(client *elastic.Client, name string) error {
	log.Printf("[INFO] Search index %s", name)
	exists, err := client.IndexExists(name).Do()
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("[INFO] Create index %s", name)
		_, err := client.CreateIndex(name).Do()
		if err != nil {
			return err
		}
	}
	return nil

}

func Save(client *elastic.Client, index string, typename string, id string,
	body interface{}) (*elastic.IndexResponse, error) {
	return client.Index().
		Index(index).
		Type(typename).
		Id(id).
		BodyJson(body).
		Do()
}
