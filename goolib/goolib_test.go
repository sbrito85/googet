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

package goolib

import (
	"testing"
)

func TestScriptInterpreter(t *testing.T) {
	table := []struct {
		script string
		eitp   string
	}{
		{"/file/path/script.ps1", "powershell"},
		{"/file/path/script.cmd", "cmd"},
		{"/file/path/script.bat", "cmd"},
	}
	for _, tt := range table {
		itp, err := scriptInterpreter(tt.script)
		if err != nil {
			t.Errorf("error parsing interpreter: %v", err)
		}
		if itp != tt.eitp {
			t.Errorf("did not get expected interpreter: got %v, want %v", itp, tt.eitp)
		}
	}
}

func TestBadScriptInterpreter(t *testing.T) {
	if _, err := scriptInterpreter("/file/path/script.ext"); err == nil {
		t.Errorf("got no error from scriptInterpreter when processing bad extension, want error")
	}
	if _, err := scriptInterpreter("/file/path/script"); err == nil {
		t.Errorf("got no error from scriptInterpreter when processing no extension, want error")
	}
}

func TestContainsInt(t *testing.T) {
	table := []struct {
		a     int
		slice []int
		want  bool
	}{
		{1, []int{1, 2}, true},
		{3, []int{1, 2}, false},
	}
	for _, tt := range table {
		if got, want := ContainsInt(tt.a, tt.slice), tt.want; got != want {
			t.Errorf("Contains(%d, %v) incorrect return: got %v, want %t", tt.a, tt.slice, got, want)
		}
	}
}

func TestContainsString(t *testing.T) {
	table := []struct {
		a     string
		slice []string
		want  bool
	}{
		{"a", []string{"a", "b"}, true},
		{"c", []string{"a", "b"}, false},
	}
	for _, tt := range table {
		if got, want := ContainsString(tt.a, tt.slice), tt.want; got != want {
			t.Errorf("Contains(%s, %v) incorrect return: got %v, want %t", tt.a, tt.slice, got, want)
		}
	}
}
