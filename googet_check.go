/*
Copyright 2016 Google Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// The check subcommand searches the repo for packages using the filter provided. The default
// filter is an empty string and will return all packages.

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/googet/v2/client"
	"github.com/google/googet/v2/goolib"
	"github.com/google/googet/v2/googetdb"
	"github.com/google/googet/v2/system"
	"github.com/google/logger"
	"github.com/google/subcommands"
)

type checkCmd struct {
	sources string
	dryRun bool
}

func (*checkCmd) Name() string     { return "check" }
func (*checkCmd) Synopsis() string { return "check and take over exsiting packages" }
func (*checkCmd) Usage() string {
	return fmt.Sprintf(`%s check [-sources repo1,repo2...] [-dry_run=true]`, filepath.Base(os.Args[0]))
}

func (cmd *checkCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.dryRun, "dry_run", false, "Don't make any changes to the DB.")
	f.StringVar(&cmd.sources, "sources", "", "comma separated list of sources, setting this overrides local .repo files")
}

func (cmd *checkCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	exitCode := subcommands.ExitFailure
	cache := filepath.Join(rootDir, cacheDir)
	var filteredState []goolib.RepoSpec
	db, err := googetdb.NewDB(filepath.Join(rootDir, dbFile))
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	state, err := db.FetchPkgs()
	//var newPkgs client.GooGetState 
	downloader, err := client.NewDownloader(proxyServer)
	if err != nil {
		logger.Fatal(err)
	}
	repos, err := buildSources(cmd.sources)
	if err != nil {
		logger.Fatal(err)
	}
	rm := downloader.AvailableVersions(ctx, repos, cache, cacheLife)
	for _, repo := range rm {
		for _, p := range repo.Packages {
			match := false
			for _, k := range state {
				if p.PackageSpec.Name == k.PackageSpec.Name {
					match = true
					break
				}
			}
			if !match {
				filteredState = append(filteredState, p)
			}
		}
	}
	fmt.Println("Searching for unmanaged software...")
	unmanaged := make(map[string]string)
	for _, v := range filteredState {
		app, _ := system.AppAssociation(v.PackageSpec.Authors, "", v.PackageSpec.Name, filepath.Ext(v.PackageSpec.Install.Path))
		if app != "" {
			/*pi := goolib.PackageInfo {
					Name: v.PackageSpec.Name, 
					Arch: v.PackageSpec.Arch, 
					Ver: v.PackageSpec.Version
				}
			if err := install.FromRepo(ctx, pi, r, cache, rm, archs, &newPkgs, true, downloader); err != nil {
				logger.Errorf("Error installing %s.%s.%s: %v", pi.Name, pi.Arch, pi.Ver, err)
				exitCode = subcommands.ExitFailure
				continue
			}*/
			logger.Infof("Unmanaged software found(packagename: application name): %v: %v\n", v.PackageSpec.Name, app)
			unmanaged[v.PackageSpec.Name] = app
		}
	}
	if len(unmanaged) > 0 {
		fmt.Println("Found the following unmanaged software (Package: Software name ...)")
		for k, v := range unmanaged {
			fmt.Printf(" %v: %v\n", k, v)
		}	
	}
	return exitCode
}
