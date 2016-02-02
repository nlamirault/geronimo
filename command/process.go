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
	"flag"
	//"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/mitchellh/cli"

	gh "github.com/nlamirault/geronimo/providers/github"
)

type ProcessCommand struct {
	UI cli.Ui
}

func (c *ProcessCommand) Help() string {
	helpText := `
Usage: geronimo process [options]
	Start the daemon
Options:
	--debug                       Debug mode enabled

Action :
`
	return strings.TrimSpace(helpText)
}

func (c *ProcessCommand) Synopsis() string {
	return "Manage open source analyse"
}

func (c *ProcessCommand) Run(args []string) int {
	var debug bool
	f := flag.NewFlagSet("process", flag.ContinueOnError)
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
	user, _, err := githubClient.Users.Get(conf.Github.User)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	log.Printf("[DEBUG] Github user: %s", user)
	events, _, err := githubClient.Activity.ListEventsPerformedByUser(
		*user.Login, true, &github.ListOptions{})
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	for _, event := range events {
		log.Printf("[INFO] Repository: %v", *event.Type)
	}
	return 0
}
