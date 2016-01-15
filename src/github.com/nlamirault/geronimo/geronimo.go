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
	"flag"
	"fmt"
	"log"

	"github.com/nlamirault/geronimo/config"
	"github.com/nlamirault/geronimo/logging"
	"github.com/nlamirault/geronimo/version"
)

var (
	vrsn  bool
	debug bool
)

func init() {
	// parse flags
	flag.BoolVar(&vrsn, "version", false, "print version and exit")
	flag.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")

	flag.Usage = func() {
		flag.PrintDefaults()
	}

	flag.Parse()
}

func getConfigurationFile() string {
	return "/home/nlamirault/.config/geronimo/geronimo.toml"
}

func setup(filename string, debug bool) (*config.Configuration, error) {
	if debug {
		logging.SetLogging("DEBUG")
	} else {
		logging.SetLogging("INFO")
	}
	return config.Load(filename)
}

func main() {
	if vrsn {
		fmt.Printf("Geronimo v%s\n", version.Version)
		return
	}
	conf, err := setup(getConfigurationFile(), debug)
	if err != nil {
		log.Printf("[ERROR] Can't setup : %s", err.Error())
		return
	}
	synchronize(conf)

}
