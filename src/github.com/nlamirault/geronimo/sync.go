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

package main

import (
	"fmt"
	"log"
	"strings"
	//"sync"
	"time"

	"github.com/google/go-github/github"
	"gopkg.in/olivere/elastic.v3"

	"github.com/nlamirault/geronimo/config"
	gh "github.com/nlamirault/geronimo/providers/github"
	"github.com/nlamirault/geronimo/storage"
)

const (
	// DefaultNumFetchProcs is the default number of goroutines fetching data
	// from the GitHub API in parallel.
	DefaultNumFetchProcs = 10

	// DefaultNumIndexProcs is the default number of goroutines indexing data
	// into Elastic Search in parallel.
	DefaultNumIndexProcs = 4

	// DefaultFrom is the default starting number for syncing repository items.
	DefaultFrom = 1

	// DefaultPerPage is the default number of items per page in GitHub API
	// requests.
	DefaultPerPage = 100

	// DefaultSleepPerPage is the default number of seconds to sleep between
	// each GitHub page queried.
	DefaultSleepPerPage = 0
)

var (
	toFetch chan *github.Repository
	toIndex chan *github.Repository
	// wgFetch sync.WaitGroup
	// wgIndex sync.WaitGroup
	options syncOptions
)

type syncOptions struct {

	// NumFetchProcs is the number of goroutines fetching GitHub data in
	// parallel.
	NumFetchProcs int

	// NumIndexProcs is the number of goroutines storing data into the Elastic
	// Search backend in parallel.
	NumIndexProcs int

	// PerPage is the number of items per page in GitHub API requests.
	PerPage int

	// From is the index to start syncing from
	From int
}

func synchronize(conf *config.Configuration) {
	log.Printf("[DEBUG] Configuration : %v", conf)
	githubClient := gh.NewClient(conf.Github.APIToken)
	esClient, err := storage.NewClient(conf.ElasticSearch.Host)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	info, _, err := esClient.Ping(conf.ElasticSearch.Host).Do()
	if err != nil {
		log.Printf(err.Error())
		return
	}
	log.Printf("[DEBUG] Elasticsearch: %s ", info.Version.Number)
	user, _, err := githubClient.Users.Get(conf.Github.User)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	log.Printf("[DEBUG] Github user: %s", user)
	options = syncOptions{
		NumFetchProcs: DefaultNumFetchProcs,
		NumIndexProcs: DefaultNumIndexProcs,
		From:          DefaultFrom,
		PerPage:       DefaultPerPage,
	}
	toFetch = make(chan *github.Repository, options.NumFetchProcs)
	toIndex = make(chan *github.Repository, options.NumIndexProcs)
	execute(user, githubClient, esClient)
}

func retrieveUserRepositories(client *github.Client, user *github.User) ([]github.Repository, error) {
	var repos []github.Repository
	log.Printf("Options: %#v", options)
	for page := options.From/options.PerPage + 1; page != 0; {
		r, resp, err := client.Repositories.List(
			*user.Login,
			&github.RepositoryListOptions{
				ListOptions: github.ListOptions{
					PerPage: options.PerPage,
					Page:    page,
				},
				Type: "owner"})
		page = resp.NextPage
		if err != nil {
			log.Printf("[ERROR] Retrieve repositories: %s", err.Error())
			continue
		}
		repos = append(repos, r...)
		time.Sleep(time.Duration(1) * time.Second)
	}
	return repos, nil
}

func execute(user *github.User, ghClient *github.Client, esClient *elastic.Client) error {
	repos, err := retrieveUserRepositories(ghClient, user)
	if err != nil {
		return err
	}

	err = indexingUser(esClient, user)
	if err != nil {
		return err
	}

	username := strings.ToLower(*user.Login)
	for _, repo := range repos {
		log.Printf("[INFO] Repository: %s", *repo.Name)
		// toFetch <- &repo
		// toIndex <- &repo
		fetchingRepository(ghClient, username, &repo)
		indexingRepository(esClient, username, &repo)
	}

	log.Printf("[INFO] Done indexing repositories in ElasticSearch")
	return nil
}

func indexingUser(client *elastic.Client, user *github.User) error {
	data := storage.User{
		Login:    *user.Login,
		Name:     *user.Name,
		Company:  *user.Company,
		Email:    *user.Email,
		Location: *user.Location,
	}
	_, err := storage.Save(
		client, *user.Login, "user", fmt.Sprintf("%d", *user.ID), data)
	return err
}

func fetchingRepository(client *github.Client, username string, repo *github.Repository) {
	//repo := <-toFetch
	log.Printf("[INFO] Fetch repository: %s", *repo.Name)
}

func indexingRepository(client *elastic.Client, username string, repo *github.Repository) {
	//repo := <-toIndex
	log.Printf("[INFO] Index repository: %s", *repo.Name)
	err := storage.CreateIndex(
		client, strings.ToLower(fmt.Sprintf("%s_%s", username, *repo.Name)))
	if err != nil {
		log.Printf("[ERROR] Can't create index for repository %s: %s",
			*repo.Name, err.Error())
		return
	}
	saveRepository(client, username, repo)
}

func saveRepository(client *elastic.Client, username string, repo *github.Repository) {
	lang := "None"
	if repo.Language != nil {
		lang = *repo.Language
	}
	forks := 0
	if repo.ForksCount != nil {
		forks = *repo.ForksCount
	}
	stars := 0
	if repo.StargazersCount != nil {
		stars = *repo.StargazersCount
	}
	subscribers := 0
	if repo.SubscribersCount != nil {
		subscribers = *repo.SubscribersCount
	}
	watchers := 0
	if repo.WatchersCount != nil {
		watchers = *repo.WatchersCount
	}
	openissues := 0
	if repo.OpenIssuesCount != nil {
		openissues = *repo.OpenIssuesCount
	}
	data := storage.Repository{
		Name:             *repo.Name,
		Description:      *repo.Description,
		Created:          fmt.Sprintf("%s", *repo.CreatedAt),
		ForksCount:       forks,
		StarsCount:       stars,
		SubscribersCount: subscribers,
		WatchersCount:    watchers,
		OpenIssuesCount:  openissues,
		Language:         lang,
	}
	log.Printf("[INFO] Store data : %#v", data)
	put, err := storage.Save(
		client, username, "repository", fmt.Sprintf("%d", *repo.ID), data)
	if err != nil {
		log.Printf("[ERROR] %s", err.Error())
	}
	log.Printf("[INFO] Indexed repository %s to index %s, type %s\n",
		put.Id, put.Index, put.Type)
}
