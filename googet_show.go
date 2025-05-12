package main

// The update subcommand handles bulk updating of packages.

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"encoding/json"

	"github.com/google/googet/v2/db"
	"github.com/google/subcommands"
	"github.com/google/logger"
)

type showCmd struct{}

func (*showCmd) Name() string     { return "show" }
func (*showCmd) Synopsis() string { return "Show the package spec information of a single package" }
func (*showCmd) Usage() string {
	return fmt.Sprintf("%s show <package>\n", filepath.Base(os.Args[0]))
}

func (cmd *showCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *showCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	goodb, err := db.NewDB(filepath.Join(rootDir, dbFile))
	if err != nil {
		logger.Fatal(err)
	} 
	var name string
	switch f.NArg() {
	case 0:
		fmt.Fprintln(os.Stderr, "Not enough arguments")
		f.Usage()
		return subcommands.ExitUsageError
	case 1:
		name = f.Arg(0)
	default:
		fmt.Fprintln(os.Stderr, "Excessive arguments")
		f.Usage()
		return subcommands.ExitUsageError
	}
	pkg := goodb.FetchPkg(name)
	marshaled, err := json.MarshalIndent(pkg, "", "  ")
   if err != nil {
      logger.Fatal("marshaling error: %s", err)
   }
   fmt.Println(string(marshaled))
	return subcommands.ExitSuccess
}