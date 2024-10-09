/*
Copyright 2018 Google Inc. All Rights Reserved.
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

// Package system handles system specific functions.
package system

import (
	"path/filepath"

	"github.com/google/googet/v2/goolib"
	"github.com/google/googet/v2/oswrap"
	"github.com/google/logger"
)

// Regex taken from Winget uninstaller
// https://github.com/microsoft/winget-cli/blob/6ea13623e5e4b870b81efeea9142d15a98dd4208/src/AppInstallerCommonCore/NameNormalization.cpp#L262
var (
	programNameReg = []string{
		"PrefixParens",
		"EmptyParens",
		"Version",
		"TrailingSymbols",
		"LeadingSymbols",
		"FilePathParens",
		"FilePathQuotes",
		"FilePath",
		"VersionLetter",
		"VersionLetterDelimited",
		"En",
		"NonNestedBracket",
		"BracketEnclosed",
		"URIProtocol",
	}
	publisherNameReg = []string{
		"VersionDelimited",
		"Version",
		"NonNestedBracket",
		"BracketEnclosed",
		"URIProtocol",
	}
	regex = map[string]string{
		"PrefixParens":           `(^\(.*?\))`,
		"EmptyParens":            `((\(\s*\)|\[\s*\]|"\s*"))`,
		"Version":                `(?:^)|(?:P|V|R|VER|VERSI(?:O|Ó)N|VERSÃO|VERSIE|WERSJA|BUILD|RELEASE|RC|SP)(?:\P{L}|\P{L}\p{L})?(\p{Nd}|\.\p{Nd})+(?:RC|B|A|R|V|SP)?\p{Nd}?`,
		"TrailingSymbols":        `([^\p{L}\p{Nd}]+$)`,
		"LeadingSymbols":         `(^[^\p{L}\p{Nd}]+)`,
		"FilePathParens":         `(\([CDEF]:\\(.+?\\)*[^\s]*\\?\))`,
		"FilePathQuotes":         `("[CDEF]:\\(.+?\\)*[^\s]*\\?")`,
		"FilePath":               `(((INSTALLED\sAT|IN)\s)?[CDEF]:\\(.+?\\)*[^\s]*\\?)`,
		"VersionLetter":          `((?:^\p{L})(?:(?:V|VER|VERSI(?:O|Ó)N|VERSÃO|VERSIE|WERSJA|BUILD|RELEASE|RC|SP)\P{L})?\p{Lu}\p{Nd}+(?:[\p{Po}\p{Pd}\p{Pc}]\p{Nd}+)+)`,
		"VersionLetterDelimited": `((?:^\p{L})(?:(?:V|VER|VERSI(?:O|Ó)N|VERSÃO|VERSIE|WERSJA|BUILD|RELEASE|RC|SP)\P{L})?\p{Lu}\p{Nd}+(?:[\p{Po}\p{Pd}\p{Pc}]\p{Nd}+)+)`,
		"En":                     `(\sEN\s*$)`,
		"NonNestedBracket":       `(\([^\(\)]*\)|\[[^\[\]]*\])`,
		"BracketEnclosed":        `((?:\p{Ps}.*\p{Pe}|".*"))`,
		"URIProtocol":            `((?:^\p{L})(?:http[s]?|ftp):\/\/)`,
	}
)

// Verify runs a verify command given a package extraction directory and a PkgSpec struct.
func Verify(dir string, ps *goolib.PkgSpec) error {
	v := ps.Verify
	if v.Path == "" {
		return nil
	}

	logger.Infof("Running verify command: %q", v.Path)
	out, err := oswrap.Create(filepath.Join(dir, "googet_verify.log"))
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			logger.Error(err)
		}
	}()
	return goolib.Exec(filepath.Join(dir, v.Path), v.Args, v.ExitCodes, out)
}
