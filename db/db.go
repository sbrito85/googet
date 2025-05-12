// Package db manages the googet state sqlite database.
package db

import (
	"database/sql"
	"context"
	"errors"
	"strings"
	"strconv"
	"os"
	"io/fs"
	"fmt"
	_ "modernc.org/sqlite" // Import the SQLite driver (unnamed)

	"github.com/google/googet/v2/client"
	"github.com/google/googet/v2/goolib"
)


type gooDB struct {
	db *sql.DB
}

// NewDB returns the googet DB object
func NewDB(dbFile string) (*gooDB, error) {
	var gdb gooDB
	if _, err := os.Stat(dbFile); errors.Is(err, fs.ErrNotExist) {
		gdb.db, _ = createDB(dbFile)
		return &gdb, nil
	}
	goodb, err := sql.Open("sqlite", dbFile)
	if err != nil {
      return nil, err
	}
	gdb.db = goodb
	return &gdb, nil
}

// Create db creates the initial googet database
func createDB(dbFile string) (*sql.DB, error) {
	goodb, err := sql.Open("sqlite", dbFile)
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
		);
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
		);
	CREATE TABLE IF NOT EXISTS pkgInstallers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEXT NOT NULL,
			ScriptType TEXT NOT NULL,
			Path TEXT NOT NULL,
			Args TEXT,
			ExitCodes TEXT,
			UNIQUE(PkgName, ScriptType) ON CONFLICT REPLACE
		);
	CREATE TABLE IF NOT EXISTS pkgFiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEST NOT NULL,
			FileName TEXT NOT NULL,
			Path TEXT NOT NULL
		);
	CREATE TABLE IF NOT EXISTS pkgDeps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEST NOT NULL,
			Dependency TEXT NOT NULL,
			Version TEXT
		);
	CREATE TABLE IF NOT EXISTS pkgInsFiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEST NOT NULL,
			Name NOT NULL,
			Hash NOT NULL
		);
	CREATE TABLE IF NOT EXISTS pkgTags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			PkgName TEST NOT NULL,
			Name NOT NULL,
			Description NOT NULL
		);
	COMMIT;
		`

	_, err = goodb.ExecContext(context.Background(), createDBQuery)
	if err != nil {
		fmt.Println("%v", err)
	}

	return goodb, nil
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
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	stmt, err := tx.PrepareContext(context.Background(), `
	INSERT or REPLACE INTO state (PkgName, SourceRepo, DownloadURL, Checksum, LocalPath, UnpackDir) VALUES (
		?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(context.Background(), spec.Name, pkgState.SourceRepo, pkgState.DownloadURL, pkgState.Checksum, pkgState.LocalPath, pkgState.UnpackDir, spec.Name)
	if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	stmt, err = tx.PrepareContext(context.Background(), `
	INSERT or REPLACE INTO pkgspec (PkgName, Version, Arch, Description, License, Authors, Owners, Source, Replaces, Conflicts) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	_, err = stmt.ExecContext(context.Background(), spec.Name, spec.Version, spec.Arch, spec.Description, spec.License, 
										   spec.Authors, spec.Owners, spec.Source, strings.Join(spec.Replaces, ","), strings.Join(spec.Conflicts, ","))
	if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
    stmt, err = tx.PrepareContext(context.Background(), `
	INSERT or REPLACE INTO pkgInstallers (PkgName, ScriptType, Path, Args, ExitCodes) VALUES (
		 ?, ?, ?, ?, ?) 
	`)
	_, err = stmt.ExecContext(context.Background(), spec.Name, "Install", spec.Install.Path, strings.Join(spec.Install.Args, ","), strings.Trim(strings.Join(strings.Fields(fmt.Sprint(spec.Install.ExitCodes)), ","), "[]"))
	if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	_, err = stmt.ExecContext(context.Background(), spec.Name, "Uninstall", spec.Uninstall.Path, strings.Join(spec.Uninstall.Args, ","), strings.Trim(strings.Join(strings.Fields(fmt.Sprint(spec.Uninstall.ExitCodes)), ","), "[]"))
	if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	stmt.ExecContext(context.Background(), spec.Name, "Verify", spec.Verify.Path, strings.Join(spec.Verify.Args, ","), strings.Trim(strings.Join(strings.Fields(fmt.Sprint(spec.Verify.ExitCodes)), ","), "[]"))
	if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}

	stmt, err = tx.PrepareContext(context.Background(), `
	INSERT INTO pkgFiles (PkgName, FileName, Path) VALUES (
		?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	for k, v := range spec.Files {
		_, err = stmt.ExecContext(context.Background(), spec.Name, k, v)
		if err != nil {
			tx.Rollback()
			fmt.Println("Unable to update record %s: %v", spec.Name, err)
		}
	}

	stmt, err = tx.PrepareContext(context.Background(), `
	INSERT INTO pkgTags (PkgName, Name, Description) VALUES (
		?, ?, ?, ?)
	`)
    if err != nil {
		tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	for k, v := range spec.Tags {
		_, err = stmt.ExecContext(context.Background(), spec.Name, k, v)
		if err != nil {
			tx.Rollback()
			fmt.Println("Unable to update record %s: %v", spec.Name, err)
		}
	}
	stmt, err = tx.PrepareContext(context.Background(), `
	INSERT INTO pkgInsFiles (PkgName, Name, Hash) VALUES (
		?, ?, ?)
	`)
    if err != nil {
    	tx.Rollback()
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
	}
	for k, v := range pkgState.InstalledFiles {
		_, err = stmt.ExecContext(context.Background(), spec.Name, k, v)
		if err != nil {
			tx.Rollback()
			fmt.Println("Unable to update record %s: %v", spec.Name, err)
		}
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println("Unable to update record %s: %v", spec.Name, err)
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
		fmt.Println("%v", err)
	}
}

// FetchPkg exports a sinfle package from the googet database
func (g *gooDB) FetchPkg(pkgName string) *client.PackageState {
	var pkgState client.PackageState
	var pkgSpec goolib.PkgSpec
	specquery := 		
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
	spec, err := g.db.Query(specquery, pkgName)
	if err != nil {
		fmt.Println("%v", err)
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
			fmt.Println("%v", err)
		}
		pkgSpec.Replaces = strings.Split(replaces, " ")
		pkgSpec.Conflicts = strings.Split(replaces, " ")
	}
	installerQuery := `Select ScriptType, Path, Args, ExitCodes FROM pkgInstallers Where PkgName = ?`
	ins, err := g.db.Query(installerQuery, pkgName)
	if err != nil {
		fmt.Println("%v", err)
	}
	for ins.Next() {
		var sType, path, args, eCodes string
		err = ins.Scan(
			&sType,
			&path,
			&args,
			&eCodes,
		)
		
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
				pkgSpec.Uninstall.ExitCodes =  processExitCodes(eCodes)
			}
	    case "Verify":
	    	pkgSpec.Verify.Path = path
			pkgSpec.Verify.Args = strings.Split(args, ",")
			if eCodes != "" {
				pkgSpec.Verify.ExitCodes = processExitCodes(eCodes)
			}
		}
	}

	insFilesQuery := `Select Name, Hash FROM pkgInsFiles Where PkgName = ?`
	insFiles, err := g.db.Query(insFilesQuery, pkgName)
    if err != nil {
			fmt.Println("%v", err)
		}
	stateInsFiles := make(map[string]string)
	for insFiles.Next() {
		var name, hash string
		err = insFiles.Scan(
			&name,
			&hash,
		)
		stateInsFiles[name] = hash
	}
	pkgState.InstalledFiles = stateInsFiles

	filesQuery := `Select FileName, Path FROM pkgFiles Where PkgName = ?`
	files, err := g.db.Query(filesQuery, pkgName)
    if err != nil {
			fmt.Println("%v", err)
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

	depsQuery := `Select Dependency, Version FROM pkgDeps Where PkgName = ?`
	deps, _ := g.db.Query(depsQuery, pkgName)

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