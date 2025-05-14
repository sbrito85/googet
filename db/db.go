/*
Copyright 2025 Google Inc. All Rights Reserved.
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

// Package db manages the googet state sqlite database.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/googet/v2/client"
	"github.com/google/googet/v2/goolib"

	_ "modernc.org/sqlite" // Import the SQLite driver (unnamed)
)

const (
	stateQuery = `INSERT or REPLACE INTO state (PkgName, SourceRepo, DownloadURL, Checksum, LocalPath, UnpackDir) VALUES (
		?, ?, ?, ?, ?, ?)`
	specQuery = `INSERT or REPLACE INTO pkgspec (PkgName, Version, Arch, Description, License, Authors, Owners, Source, Replaces, Conflicts) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	installerQuery = `INSERT or REPLACE INTO pkgInstallers (PkgName, ScriptType, Path, Args, ExitCodes) VALUES (
		 ?, ?, ?, ?, ?)`
	filesQuery = `INSERT INTO pkgFiles (PkgName, FileName, Path) VALUES (
		?, ?, ?)`
	tagQuery = `INSERT INTO pkgTags (PkgName, Name, Description) VALUES (
		?, ?, ?, ?)`
	insFilesQuery = `INSERT INTO pkgInsFiles (PkgName, Name, Checksum) VALUES (
		?, ?, ?)`
)

type gooDB struct {
	db *sql.DB
}

// NewDB returns the googet DB object
func NewDB(dbFile string) (*gooDB, error) {
	var gdb gooDB
	var err error
	if _, err := os.Stat(dbFile); errors.Is(err, os.ErrNotExist) {
		gdb.db, _ = createDB(dbFile)
		return &gdb, nil
	}
	gdb.db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}
	return &gdb, nil
}

// Create db creates the initial googet database
func createDB(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	createDBQuery := `BEGIN;
	CREATE TABLE IF NOT EXISTS state (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEXT NOT NULL UNIQUE,
			SourceRepo TEXT NOT NULL,
			DownloadURL TEXT NOT NULL,
			Checksum TEXT,
			LocalPath TEXT,
			UnpackDir TEXT
		) STRICT;
	CREATE TABLE IF NOT EXISTS pkgspec (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEXT NOT NULL UNIQUE,
			Version TEXT NOT NULL,
			Arch TEXT,
			Description TEXT,
			License TEXT,
			Authors TEXT,
			Owners TEXT,
			Source TEXT,
			Replaces TEXT,
			Conflicts TEXT
		) STRICT;
	CREATE TABLE IF NOT EXISTS pkgInstallers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEXT NOT NULL,
			ScriptType TEXT NOT NULL,
			Path TEXT NOT NULL,
			Args TEXT,
			ExitCodes TEXT,
			UNIQUE(PkgName, ScriptType) ON CONFLICT REPLACE
		) STRICT;
	CREATE TABLE IF NOT EXISTS pkgFiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEXT NOT NULL,
			FileName TEXT NOT NULL,
			Path TEXT NOT NULL
		) STRICT;
	CREATE TABLE IF NOT EXISTS pkgDeps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEXT NOT NULL,
			Dependency TEXT NOT NULL,
			Version TEXT
		) STRICT;
	CREATE TABLE IF NOT EXISTS pkgInsFiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEXT NOT NULL,
			Name TEXT NOT NULL,
			Checksum TEXT NOT NULL
		) STRICT;
	CREATE TABLE IF NOT EXISTS pkgTags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEXT NOT NULL,
			Name TEXT NOT NULL,
			Description TEXT NOT NULL
		) STRICT;
	CREATE TABLE IF NOT EXISTS pkgMappings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		PkgName TEXT NOT NULL,
		InstalledApp TEXT NOT NULL
		) STRICT;
	COMMIT;
		`

	_, err = db.ExecContext(context.Background(), createDBQuery)
	if err != nil {
		fmt.Printf("%v", err)
		return nil, err
	}

	return db, nil
}

// WriteStateToDB writes new or partial state to the db.
func (g *gooDB) WriteStateToDB(gooState *client.GooGetState) error {
	for _, pkgState := range *gooState {
		g.addPkg(pkgState)
	}
	return nil
}

func (g *gooDB) addPkg(pkgState client.PackageState) {
	spec := pkgState.PackageSpec

	tx, err := g.db.Begin()
	if err != nil {
		fmt.Printf("Unable to update record %s: %v", spec.Name, err)
	}

	_, err = tx.ExecContext(context.Background(), stateQuery, spec.Name, pkgState.SourceRepo, pkgState.DownloadURL, pkgState.Checksum, pkgState.LocalPath, pkgState.UnpackDir)
	if err != nil {
		tx.Rollback()
		fmt.Printf("Unable to update record %s: %v", spec.Name, err)
	}

	_, err = tx.ExecContext(context.Background(), specQuery, spec.Name, spec.Version, spec.Arch, spec.Description, spec.License,
		spec.Authors, spec.Owners, spec.Source, strings.Join(spec.Replaces, ","), strings.Join(spec.Conflicts, ","))
	if err != nil {
		tx.Rollback()
		fmt.Printf("Unable to update record %s: %v", spec.Name, err)
	}

	_, err = tx.ExecContext(context.Background(), installerQuery, spec.Name, "Install", spec.Install.Path, strings.Join(spec.Install.Args, ","), strings.Trim(strings.Join(strings.Fields(fmt.Sprint(spec.Install.ExitCodes)), ","), "[]"))
	if err != nil {
		tx.Rollback()
		fmt.Printf("Unable to update record %s: %v", spec.Name, err)
	}
	_, err = tx.ExecContext(context.Background(), installerQuery, spec.Name, "Uninstall", spec.Uninstall.Path, strings.Join(spec.Uninstall.Args, ","), strings.Trim(strings.Join(strings.Fields(fmt.Sprint(spec.Uninstall.ExitCodes)), ","), "[]"))
	if err != nil {
		tx.Rollback()
		fmt.Printf("Unable to update record %s: %v", spec.Name, err)
	}
	_, err = tx.ExecContext(context.Background(), installerQuery, spec.Name, "Verify", spec.Verify.Path, strings.Join(spec.Verify.Args, ","), strings.Trim(strings.Join(strings.Fields(fmt.Sprint(spec.Verify.ExitCodes)), ","), "[]"))
	if err != nil {
		tx.Rollback()
		fmt.Printf("Unable to update record %s: %v", spec.Name, err)
	}

	for k, v := range spec.Files {
		_, err = tx.ExecContext(context.Background(), filesQuery, spec.Name, k, v)
		if err != nil {
			tx.Rollback()
			fmt.Printf("Unable to update record %s: %v", spec.Name, err)
		}
	}

	for k, v := range spec.Tags {
		_, err = tx.ExecContext(context.Background(), tagQuery, spec.Name, k, string(v))
		if err != nil {
			tx.Rollback()
			fmt.Printf("Unable to update record %s: %v", spec.Name, err)
		}
	}

	for k, v := range pkgState.InstalledFiles {
		_, err = tx.ExecContext(context.Background(), insFilesQuery, spec.Name, k, v)
		if err != nil {
			tx.Rollback()
			fmt.Printf("Unable to update record %s: %v", spec.Name, err)
		}
	}
	err = tx.Commit()
	if err != nil {
		fmt.Printf("Unable to update record %s: %v", spec.Name, err)
	}
}

// RemovePkg removes a single package from the googet database
func (g *gooDB) RemovePkg(packageName string) {
	removeQuery := fmt.Sprintf(`BEGIN;
	DELETE FROM state where PkgName = '%[1]v';
	DELETE FROM pkgspec where PkgName = '%[1]v';
	DELETE FROM pkgInstallers where PkgName = '%[1]v';
	DELETE FROM pkgInstallers where PkgName = '%[1]v';
	DELETE FROM pkgDeps where PkgName = '%[1]v';
	DELETE FROM pkgInsFiles where PkgName = '%[1]v';
	DELETE FROM pkgTags where PkgName = '%[1]v';
	COMMIT;`, packageName)

	_, err := g.db.ExecContext(context.Background(), removeQuery)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

// FetchPkg exports a sinfle package from the googet database
func (g *gooDB) FetchPkg(pkgName string) *client.PackageState {
	var pkgState client.PackageState
	var pkgSpec goolib.PkgSpec
	selectSpecQuery :=
		`SELECT 
			pkgspec.PkgName,
			pkgspec.Version,
			pkgspec.Arch,
			pkgspec.Description,
			pkgspec.License,
			pkgspec.Authors,
			pkgspec.Owners,
			pkgspec.Source,
			pkgspec.Replaces,
			pkgspec.Conflicts,
			state.SourceRepo,
			state.DownloadURL,
			state.Checksum,
			state.LocalPath,
			state.UnpackDir
		FROM
			pkgspec
			LEFT JOIN state ON state.PkgName = pkgspec.PkgName
		WHERE pkgspec.PkgName = ?
		ORDER BY pkgspec.PkgName
		`
	spec, err := g.db.Query(selectSpecQuery, pkgName)
	if err != nil {
		fmt.Printf("%v", err)
	}
	for spec.Next() {
		var replaces, conflicts string
		err = spec.Scan(
			&pkgSpec.Name,
			&pkgSpec.Version,
			&pkgSpec.Arch,
			&pkgSpec.Description,
			&pkgSpec.License,
			&pkgSpec.Authors,
			&pkgSpec.Owners,
			&pkgSpec.Source,
			&replaces,
			&conflicts,
			&pkgState.SourceRepo,
			&pkgState.DownloadURL,
			&pkgState.Checksum,
			&pkgState.LocalPath,
			&pkgState.UnpackDir,
		)
		if err != nil {
			fmt.Printf("%v", err)
		}
		if replaces != "" {
			pkgSpec.Replaces = strings.Split(replaces, ",")
		}
		if conflicts != "" {
			pkgSpec.Conflicts = strings.Split(conflicts, ",")
		}
	}
	selectInstallerQuery := `Select ScriptType, Path, Args, ExitCodes FROM pkgInstallers Where PkgName = ?`
	ins, err := g.db.Query(selectInstallerQuery, pkgName)
	if err != nil {
		fmt.Printf("%v", err)
	}
	for ins.Next() {
		var sType, path, args, eCodes string
		err = ins.Scan(
			&sType,
			&path,
			&args,
			&eCodes,
		)
		if path == "" {
			continue
		}
		switch sType {
		case "Install":
			pkgSpec.Install.Path = path
			pkgSpec.Install.Args = strings.Split(args, ",")
			if eCodes != "" {
				pkgSpec.Install.ExitCodes = processExitCodes(eCodes)
			}
		case "Uninstall":
			pkgSpec.Uninstall.Path = path
			pkgSpec.Uninstall.Args = strings.Split(args, ",")
			if eCodes != "" {
				pkgSpec.Uninstall.ExitCodes = processExitCodes(eCodes)
			}
		case "Verify":
			pkgSpec.Verify.Path = path
			pkgSpec.Verify.Args = strings.Split(args, ",")
			if eCodes != "" {
				pkgSpec.Verify.ExitCodes = processExitCodes(eCodes)
			}
		}
	}

	selectInsFilesQuery := `Select Name, Checksum FROM pkgInsFiles Where PkgName = ?`
	insFiles, err := g.db.Query(selectInsFilesQuery, pkgName)
	if err != nil {
		fmt.Printf("%v", err)
	}
	stateInsFiles := make(map[string]string)
	for insFiles.Next() {
		var name, checksum string
		err = insFiles.Scan(
			&name,
			&checksum,
		)
		stateInsFiles[name] = checksum
	}
	if len(stateInsFiles) != 0 {
		pkgState.InstalledFiles = stateInsFiles
	}

	selectFilesQuery := `Select FileName, Path FROM pkgFiles Where PkgName = ?`
	files, err := g.db.Query(selectFilesQuery, pkgName)
	if err != nil {
		fmt.Printf("%v", err)
	}
	specFiles := make(map[string]string)
	for files.Next() {
		var name, path string
		err = files.Scan(
			&name,
			&path,
		)
		specFiles[name] = path
	}
	pkgSpec.Files = specFiles

	selectDepsQuery := `Select Dependency, Version FROM pkgDeps Where PkgName = ?`
	deps, _ := g.db.Query(selectDepsQuery, pkgName)

	specDeps := make(map[string]string)
	for deps.Next() {
		var dep, ver string
		err = deps.Scan(
			&dep,
			&ver,
		)
		specDeps[dep] = ver
	}
	pkgSpec.PkgDependencies = specDeps

	selectTagsQuery := `Select Name, Description FROM pkgTags Where PkgName = ?`
	tags, _ := g.db.Query(selectTagsQuery, pkgName)

	specTags := make(map[string][]byte)
	for tags.Next() {
		var name, desc string
		err = tags.Scan(
			&name,
			&desc,
		)
		specTags[name] = []byte(desc)
	}
	pkgSpec.Tags = specTags
	pkgState.PackageSpec = &pkgSpec

	return &pkgState

}

// FetchPkgs exports all of the current packages in the googet database
func (g *gooDB) FetchPkgs() *client.GooGetState {
	var state client.GooGetState

	pkgs, err := g.db.Query(`Select PkgName from pkgspec`)
	if err != nil {
		fmt.Printf("%v", err)
	}
	defer pkgs.Close()
	for pkgs.Next() {
		var pkgName string
		err = pkgs.Scan(&pkgName)
		state = append(state, *g.FetchPkg(pkgName))
	}

	return &state
}

func processExitCodes(eCodes string) []int {
	e := strings.Split(eCodes, ",")
	var err error
	codes := make([]int, len(e))
	for i, v := range e {
		codes[i], err = strconv.Atoi(v)
		if err != nil {

		}
	}
	return codes
}
