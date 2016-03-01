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

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/google/googet/client"
	"github.com/google/googet/goolib"
)

func TestRepoList(t *testing.T) {
	testRepo := "https://foo.com/googet/bar"

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	f, err := os.Create(path.Join(tempDir, "test.repo"))
	if err != nil {
		t.Fatalf("error creating repo file: %v", err)
	}
	if _, err := f.Write([]byte("url: " + testRepo)); err != nil {
		t.Fatalf("error writing repo: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("error closing repo file: %v", err)
	}

	got, err := repoList(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, []string{testRepo}) {
		t.Errorf("returned repo does not match expected repo: got %v, want %v", got, testRepo)
	}
}

func TestInstalledPackages(t *testing.T) {
	state := []client.PackageState{
		{
			PackageSpec: &goolib.PkgSpec{
				Name:    "foo",
				Version: "1.2.3@4",
				Arch:    "noarch",
			},
		},
		{
			PackageSpec: &goolib.PkgSpec{
				Name:    "bar",
				Version: "0.1.0@1",
				Arch:    "noarch",
			},
		},
	}

	want := packageMap{"foo.noarch": "1.2.3@4", "bar.noarch": "0.1.0@1"}
	got := installedPackages(state)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("returned map does not match expected map: got %v, want %v", got, want)
	}
}

func TestReadConf(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	confPath := path.Join(tempDir, "test.conf")
	f, err := os.Create(confPath)
	if err != nil {
		t.Fatalf("error creating conf file: %v", err)
	}

	content := []byte("archs: [noarch, x86_64]\ncachelife: 10m")
	if _, err := f.Write(content); err != nil {
		t.Fatalf("error writing conf file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("error closing conf file: %v", err)
	}

	readConf(confPath)

	ea := []string{"noarch", "x86_64"}
	if !reflect.DeepEqual(archs, ea) {
		t.Errorf("readConf did not create expected arch list, want: %s, got: %s", ea, archs)
	}

	ecl := time.Duration(10 * time.Minute)
	if cacheLife != ecl {
		t.Errorf("readConf did not create expected cacheLife, want: %s, got: %s", ecl, cacheLife)
	}
}

func TestRotateLog(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	table := []struct {
		name    string
		size    int64
		rotated bool
	}{
		{"test1.log", 10 * 1024, true},
		{"test2.log", 9 * 1024, false},
	}

	for _, tt := range table {
		logPath := path.Join(tempDir, tt.name)
		f, err := os.Create(logPath)
		if err != nil {
			t.Fatalf("error creating log file: %v", err)
		}

		if err := f.Truncate(tt.size); err != nil {
			t.Fatalf("error truncating log file: %v", err)
		}

		if err := f.Close(); err != nil {
			t.Fatalf("error closing log file: %v", err)
		}

		if err := rotateLog(logPath, 10*1024); err != nil {
			t.Errorf("error running rotateLog: %v", err)
		}

		switch tt.rotated {
		case true:
			if _, err := os.Stat(logPath); err == nil {
				t.Error("rotateLog did not rotate log as expected, old log file still exists")
			}
			if _, err := os.Stat(logPath + ".old"); err != nil {
				t.Error("rotateLog did not rotate log as expected, .old file does not exist")
			}
		case false:
			if _, err := os.Stat(logPath); err != nil {
				t.Error("rotateLog rotated a log we didn't expect")
			}
		}
	}
}

func TestWriteReadState(t *testing.T) {
	want := &client.GooGetState{
		client.PackageState{PackageSpec: &goolib.PkgSpec{Name: "test"}},
	}

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sf := path.Join(tempDir, "test.state")

	if err := writeState(want, sf); err != nil {
		t.Errorf("error running writeState: %v", err)
	}

	got, err := readState(sf)
	if err != nil {
		t.Errorf("error running readState: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("did not get expected state, got: %+v, want %+v", got, want)
	}
}
