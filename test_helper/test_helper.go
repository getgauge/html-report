// Copyright 2015 ThoughtWorks, Inc.

// This file is part of getgauge/html-report.

// getgauge/html-report is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// getgauge/html-report is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with getgauge/html-report.  If not, see <http://www.gnu.org/licenses/>.

package testHelper

import (
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	htmldiff "github.com/documize/html-diff"
)

var re = regexp.MustCompile("[\\s]*[\n\t][\\s]*")

func RemoveNewline(s string) string {
	return re.ReplaceAllLiteralString(s, "")
}

func AssertEqual(expected, actual, testName string, t *testing.T) {
	if expected != actual {
		diffHTML := compare(expected, actual)
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Errorf("Unable to dump to tmp file. Raw content:\n%s\n", diffHTML)
		}
		fileName := fmt.Sprintf("%s.html", tmpFile.Name())
		ioutil.WriteFile(fileName, []byte(diffHTML), 0644)
		tmpFile.Close()
		t.Errorf("%s -  View Diff Output : %s\n", testName, fileName)
	}
}

func compare(a, b string) string {
	var cfg = &htmldiff.Config{
		InsertedSpan: []htmldiff.Attribute{{Key: "style", Val: "background-color: palegreen;"}},
		DeletedSpan:  []htmldiff.Attribute{{Key: "style", Val: "background-color: lightpink;"}},
		ReplacedSpan: []htmldiff.Attribute{{Key: "style", Val: "background-color: lightskyblue;"}},
		CleanTags:    []string{""},
	}
	
	res, _ := cfg.HTMLdiff([]string{html.EscapeString(a), html.EscapeString(b)})
	return res[0]
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}
