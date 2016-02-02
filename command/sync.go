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

package command

import (
	//"encoding/json"
	"flag"
	//"fmt"
	"log"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"
	"gopkg.in/olivere/elastic.v3"

	//"github.com/nlamirault/geronimo/config"
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

// SyncCommand defines the synchronization command
type SyncCommand struct {
	UI      cli.Ui
	toFetch chan *github.Repository
	toIndex chan *github.Repository
	wgFetch sync.WaitGroup
	wgIndex sync.WaitGroup
	options syncOptions
}

// Help return the command's help message
func (c *SyncCommand) Help() string {
	helpText := `
Usage: geronimo sync [options]
	Synchronize open source projects
Options:
	--debug                       Debug mode enabled

Action :
`
	return strings.TrimSpace(helpText)
}

// Synopsis return the command's synopsis message
func (c *SyncCommand) Synopsis() string {
	return "Syncialize storage"
}

// Run represents the command execution process
func (c *SyncCommand) Run(args []string) int {
	var debug bool
	f := flag.NewFlagSet("init", flag.ContinueOnError)
	f.Usage = func() { c.UI.Output(c.Help()) }
	f.BoolVar(&debug, "debug", false, "Debug mode enabled")

	if err := f.Parse(args); err != nil {
		return 1
	}
	conf, err := getConfig(getConfigurationFile(), debug)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	log.Printf("[DEBUG] Configuration : %v", conf)
	githubClient := gh.NewClient(conf.Github.APIToken)
	esClient, err := storage.NewClient(conf.ElasticSearch.Host)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	info, _, err := esClient.Ping(conf.ElasticSearch.Host).Do()
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	log.Printf("[DEBUG] Elasticsearch: %s ", info.Version.Number)
	user, _, err := githubClient.Users.Get(conf.Github.User)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	log.Printf("[DEBUG] Github user: %s", user)
	c.options = syncOptions{
		NumFetchProcs: DefaultNumFetchProcs,
		NumIndexProcs: DefaultNumIndexProcs,
		From:          DefaultFrom,
		PerPage:       DefaultPerPage,
	}
	c.toFetch = make(chan *github.Repository, c.options.NumFetchProcs)
	c.toIndex = make(chan *github.Repository, c.options.NumIndexProcs)
	return c.execute(user, githubClient, esClient)

}

func (c *SyncCommand) retrieveUserRepositories(client *github.Client, user *github.User) ([]github.Repository, *github.Response, error) {
	return client.Repositories.List(
		*user.Login,
		&github.RepositoryListOptions{
			ListOptions: github.ListOptions{
				PerPage: c.options.PerPage,
				Page:    c.options.From,
			},
			Type: "owner"})
}

func (c *SyncCommand) execute(user *github.User, ghClient *github.Client, esClient *elastic.Client) int {
	repos, _, err := c.retrieveUserRepositories(ghClient, user)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	username := strings.ToLower(*user.Login)

	// err := createIndex(esClient, username)
	// c.indexingRepository(esClient, username)
	// if err != nil {
	// 	c.UI.Error(err.Error())
	// 	return 1
	// }

	// for i := 0; i != c.options.NumFetchProcs; i++ {
	// 	c.wgFetch.Add(1)
	// 	go c.fetchingRepository(ghClient, username)
	// }
	// for i := 0; i != c.options.NumIndexProcs; i++ {
	// 	c.wgIndex.Add(1)
	// 	go c.indexingRepository(esClient, username)
	// }

	go c.fetchingRepository(ghClient, username)
	go c.indexingRepository(esClient, username)

	for _, repo := range repos {
		log.Printf("[INFO] --> Repository: %s", *repo.Name)

		// log.Printf("[INFO] Repository: %v", *repo.Name)
		// err = createIndex(
		// 	esClient,
		// 	fmt.Sprintf(
		// 		"%s_%s", username, strings.ToLower(*repo.Name)))
		// if err != nil {
		// 	c.UI.Error(err.Error())
		// 	return 1
		// }

		c.toFetch <- &repo
		c.toIndex <- &repo
	}

	close(c.toFetch)
	close(c.toIndex)

	// c.wgFetch.Wait()

	// log.Printf("[INFO] Waiting")

	// c.wgIndex.Wait()

	log.Printf("[INFO] Done indexing repositories in ElasticSearch")

	// storeRepositoriesData(esClient, username, repos)
	// searchRepositories(esClient, "emacs")
	return 0
}

// func (c *SyncCommand) indexingRepository(client *elastic.Client, username string) {
// 	name := username
// 	if repo != nil {
// 		name = fmt.Sprintf(
// 			"%s_%s", username, strings.ToLower(*repo.Name))
// 	}
// 	exists, err := client.IndexExists(name).Do()
// 	if err != nil {
// 		log.Printf("[ERROR] Failed to check if index exists: %s", err.Error())
// 	}
// 	if !exists {
// 		log.Printf("[INFO] Create index %s", name)
// 		_, err := client.CreateIndex(name).Do()
// 		if err != nil {
// 			log.Printf("[ERROR] Failed to create index: %s", err.Error())
// 		}
// 	}

// }

func (c *SyncCommand) fetchingRepository(client *github.Client, username string) {
	// for repo := range c.toFetch {
	for {
		repo := <-c.toFetch
		log.Printf("[INFO] Fetch repository: %s", *repo.Name)
	}
	//c.wgFetch.Done()
}

func (c *SyncCommand) indexingRepository(client *elastic.Client, username string) {
	// for repo := range c.toIndex {
	for {
		repo := <-c.toIndex
		log.Printf("[INFO] Index repository: %s", *repo.Name)
		// lang := "None"
		// if repo.Language != nil {
		// 	lang = *repo.Language
		// }
		// forks := 0
		// if repo.ForksCount != nil {
		// 	forks = *repo.ForksCount
		// }
		// stars := 0
		// if repo.StargazersCount != nil {
		// 	stars = *repo.StargazersCount
		// }
		// subscribers := 0
		// if repo.SubscribersCount != nil {
		// 	subscribers = *repo.SubscribersCount
		// }
		// watchers := 0
		// if repo.WatchersCount != nil {
		// 	watchers = *repo.WatchersCount
		// }
		// openissues := 0
		// if repo.OpenIssuesCount != nil {
		// 	openissues = *repo.OpenIssuesCount
		// }
		// data := storage.Repository{
		// 	Name:             *repo.Name,
		// 	Description:      *repo.Description,
		// 	Created:          fmt.Sprintf("%s", *repo.CreatedAt),
		// 	ForksCount:       forks,
		// 	StarsCount:       stars,
		// 	SubscribersCount: subscribers,
		// 	WatchersCount:    watchers,
		// 	OpenIssuesCount:  openissues,
		// 	Language:         lang,
		// }
		// log.Printf("[INFO] Store data : %#v", data)
		// put, err := client.Index().
		// 	Index(username).
		// 	Type("repository").
		// 	Id(fmt.Sprintf("%d", *repo.ID)).
		// 	BodyJson(data).
		// 	Do()
		// if err != nil {
		// 	log.Printf("[ERROR] %s", err.Error())
		// }
		// log.Printf("[INFO] Indexed repository %s to index %s, type %s\n",
		// 	put.Id, put.Index, put.Type)
	}
	// c.wgIndex.Done()
}

// func storeRepositoryData(client *elastic.Client, username string, repo *github.Repository) {
// 	log.Printf("Repo: %s %v", *repo.Name, repo.Homepage)
// 	lang := "None"
// 	if repo.Language != nil {
// 		lang = *repo.Language
// 	}
// 	forks := 0
// 	if repo.ForksCount != nil {
// 		forks = *repo.ForksCount
// 	}
// 	stars := 0
// 	if repo.StargazersCount != nil {
// 		stars = *repo.StargazersCount
// 	}
// 	subscribers := 0
// 	if repo.SubscribersCount != nil {
// 		subscribers = *repo.SubscribersCount
// 	}
// 	watchers := 0
// 	if repo.WatchersCount != nil {
// 		watchers = *repo.WatchersCount
// 	}
// 	openissues := 0
// 	if repo.OpenIssuesCount != nil {
// 		openissues = *repo.OpenIssuesCount
// 	}
// 	data := storage.Repository{
// 		Name:             *repo.Name,
// 		Description:      *repo.Description,
// 		Created:          fmt.Sprintf("%s", *repo.CreatedAt),
// 		ForksCount:       forks,
// 		StarsCount:       stars,
// 		SubscribersCount: subscribers,
// 		WatchersCount:    watchers,
// 		OpenIssuesCount:  openissues,
// 		Language:         lang,
// 	}
// 	log.Printf("[INFO] Store data : %#v", data)
// 	put, err := client.Index().
// 		Index(username).
// 		Type("repository").
// 		Id(fmt.Sprintf("%d", *repo.ID)).
// 		BodyJson(data).
// 		Do()
// 	if err != nil {
// 		log.Printf("[ERROR] %s", err.Error())
// 	}
// 	log.Printf("[INFO] Indexed repository %s to index %s, type %s\n",
// 		put.Id, put.Index, put.Type)
// }

// func storeRepositoriesData(client *elastic.Client, username string, repos []github.Repository) {
// 	log.Printf("Repo: %s %v", *repo.Name, repo.Homepage)
// 	lang := "None"
// 	if repo.Language != nil {
// 		lang = *repo.Language
// 	}
// 	forks := 0
// 	if repo.ForksCount != nil {
// 		forks = *repo.ForksCount
// 	}
// 	stars := 0
// 	if repo.StargazersCount != nil {
// 		stars = *repo.StargazersCount
// 	}
// 	subscribers := 0
// 	if repo.SubscribersCount != nil {
// 		subscribers = *repo.SubscribersCount
// 	}
// 	watchers := 0
// 	if repo.WatchersCount != nil {
// 		watchers = *repo.WatchersCount
// 	}
// 	openissues := 0
// 	if repo.OpenIssuesCount != nil {
// 		openissues = *repo.OpenIssuesCount
// 	}
// 	data := storage.Repository{
// 		Name:             *repo.Name,
// 		Description:      *repo.Description,
// 		Created:          fmt.Sprintf("%s", *repo.CreatedAt),
// 		ForksCount:       forks,
// 		StarsCount:       stars,
// 		SubscribersCount: subscribers,
// 		WatchersCount:    watchers,
// 		OpenIssuesCount:  openissues,
// 		Language:         lang,
// 	}
// 	log.Printf("[INFO] Store data : %#v", data)
// 	put, err := client.Index().
// 		Index(username).
// 		Type("repository").
// 		Id(fmt.Sprintf("%d", *repo.ID)).
// 		BodyJson(data).
// 		Do()
// 	if err != nil {
// 		log.Printf("[WARN] %s", err.Error())
// 	}
// 	log.Printf("[INFO] Indexed repository %s to index %s, type %s\n",
// 		put.Id, put.Index, put.Type)
// }

// func searchRepositories(client *elastic.Client, term string) error {
// 	log.Printf("Search for %s", term)
// 	// Search with a term query
// 	termQuery := elastic.NewTermQuery("name", term)
// 	searchResult, err := client.Search().
// 		Index("nlamirault").
// 		Query(termQuery).
// 		//Sort("name", true).
// 		From(0).Size(10).
// 		Pretty(true).
// 		Do()
// 	if err != nil {
// 		return nil
// 	}

// 	// searchResult is of type SearchResult and returns hits, suggestions,
// 	// and all kinds of other information from Elasticsearch.
// 	log.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

// 	// Number of hits
// 	if searchResult.Hits != nil {
// 		log.Printf("Found a total of %d results\n",
// 			searchResult.Hits.TotalHits)

// 		// Iterate through results
// 		for _, hit := range searchResult.Hits.Hits {
// 			// hit.Index contains the name of the index
// 			var r storage.Repository
// 			err := json.Unmarshal(*hit.Source, &r)
// 			if err != nil {
// 				return nil
// 			}
// 			log.Printf("Repository %s: %s\n", r.Name, r.Description)
// 		}
// 	} else {
// 		log.Printf("Found no tweets\n")
// 	}

// 	return nil
// }
