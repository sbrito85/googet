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

// The installed subcommand lists out all installed packages that match the filter.
// The default filter is an empty string and will return all packages.

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/googet/v2/client"
	"github.com/google/googet/v2/db"
	"github.com/google/googet/v2/goolib"
	"github.com/google/logger"
	"github.com/google/subcommands"
)

type installedCmd struct {
	info   bool
	files  bool
	format string
}

func (*installedCmd) Name() string     { return "installed" }
func (*installedCmd) Synopsis() string { return "list installed packages" }
func (*installedCmd) Usage() string {
	return fmt.Sprintf(`%s installed [-info] [-files] [<initial>]:
	List installed packages beginning with an initial string,
	if no initial string is provided all installed packages will be listed.
`, filepath.Base(os.Args[0]))
}

func (cmd *installedCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.info, "info", false, "display package info")
	f.BoolVar(&cmd.files, "files", false, "display package file list")
	f.StringVar(&cmd.format, "format", "simple", "Formatting of the output. Supported outputs: simple, json")
}

func (cmd *installedCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var state client.GooGetState
	var exitCode subcommands.ExitStatus
	var displayText string
	goodb, err := db.NewDB(filepath.Join(rootDir, dbFile))
	if err != nil {
		logger.Fatal(err)
	}
	switch f.NArg() {
	case 0:
		state = *goodb.FetchPkgs()
		displayText = "Installed packages:\n"
	case 1:
		pkg := *goodb.FetchPkg(f.Arg(0))
		if pkg.PackageSpec.Name != "" {
			state = append(state, pkg)
		}
		displayText = fmt.Sprintf("Installed packages matching %q:\n", f.Arg(0))
		if len(state) == 0 {
			displayText = fmt.Sprintf("No package matching filter %q installed.\n", f.Arg(0))
		}
	default:
		fmt.Fprintln(os.Stderr, "Excessive arguments")
		f.Usage()
		return subcommands.ExitUsageError
	}

	switch cmd.format {
	case "simple":
		exitCode = cmd.formatSimple(&state, displayText)
	case "json":
		exitCode = cmd.formatJson(&state)
	default:
		fmt.Fprintln(os.Stderr, "Invalid format")
		f.Usage()
		return subcommands.ExitUsageError
	}
	return exitCode
}

func (cmd *installedCmd) formatSimple(state *client.GooGetState, displayText string) subcommands.ExitStatus {
	pm := installedPackages(*state)
	var pl []string
	for p, v := range pm {
		pl = append(pl, p+"."+v)
	}

	sort.Strings(pl)
	fmt.Printf(displayText)

	exitCode := subcommands.ExitFailure
	for _, p := range pl {
		exitCode = subcommands.ExitSuccess
		pi := goolib.PkgNameSplit(p)

		if cmd.info {
			local(pi, *state)
			continue
		}
		fmt.Println(" ", pi.Name+"."+pi.Arch+" "+pi.Ver)

		if cmd.files {
			ps, err := state.GetPackageState(pi)
			if err != nil {
				logger.Errorf("Unable to get file list for package %q.", p)
				continue
			}
			if len(ps.InstalledFiles) == 0 {
				fmt.Println("  - No files directly managed by GooGet.")
			}
			for file := range ps.InstalledFiles {
				fmt.Println("  -", file)
			}
		}
	}
	return exitCode
}

func (cmd *installedCmd) formatJson(state *client.GooGetState) subcommands.ExitStatus {
	marshaled, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		logger.Fatalf("marshaling error: %s", err)
	}
	fmt.Println(string(marshaled))
	return subcommands.ExitSuccess
}

func local(pi goolib.PackageInfo, state client.GooGetState) {
	for _, p := range state {
		if p.Match(pi) {
			info(p.PackageSpec, "installed")
			return
		}
	}
}
