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

package generator

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/getgauge/html-report/gauge_messages"
)

var suiteRes = &gauge_messages.ProtoSuiteResult{}

func TestHTMLGeneration(t *testing.T) {
	cont, err := ioutil.ReadFile("_testdata/expected.html")
	if err != nil {
		t.Errorf("Error reading expected HTML file: %s", err.Error())
	}

	buf := new(bytes.Buffer)
	generate(buf)

	want := removeNewline(string(cont))
	got := removeNewline(buf.String())

	if got != want {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}
