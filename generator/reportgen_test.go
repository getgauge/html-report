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
	"regexp"
	"testing"
)

type reportGenTest struct {
	name   string
	tmpl   string
	input  interface{}
	output string
}

var bodyHeader string = `<header class="top">
<div class="header">
  <div class="container">
     <div class="logo"><img src="images/logo.png" alt="Report logo"></div>
        <h2 class="project">Project: projname</h2>
      </div>
  </div>
</header>`

var chartDiv string = `<div class="report-overview">
  <div class="report_chart">
    <div class="chart">
      <nvd3 options="options" data="data"></nvd3>
    </div>
    <div class="total-specs"><span class="value">41</span> <span class="txt">Total specs</span></div>
  </div>`

var resCtrDiv string = `
  <div class="report_test-results">
    <ul>
      <li class="fail"><span class="value">2</span> <span class="txt">Failed</span></li>
      <li class="pass"><span class="value">39</span> <span class="txt">Passed</span></li>
      <li class="skip"><span class="value">0</span> <span class="txt">Skipped</span></li>
    </ul>
  </div>`

var env string = `<div class="report_details"><ul>
      <li>
        <label>Environment </label>
        <span>default</span>
      </li>`

var tags string = `
      <li>
        <label>Tags </label>
        <span>foo</span>
      </li>`

var succRate string = `
      <li>
        <label>Success Rate </label>
        <span>34%</span>
      </li>`

var execTime string = `
     <li>
        <label>Total Time </label>
        <span>00:01:53</span>
      </li>`

var timestamp string = `
     <li>
        <label>Generated On </label>
        <span>Jun 3, 2016 at 12:29pm</span>
      </li>
    </ul>
  </div>
</div>`

var re *regexp.Regexp = regexp.MustCompile("[ ]*\n[ ]*")

var reportGenTests = []reportGenTest{
	{"generate body header with project name", bodyHeaderTag, overview{ProjectName: "projname"}, bodyHeader},
	{"generate report overview with tags", reportOverviewTag, overview{"projname", "default", "foo", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", 41, 2, 39, 0},
		chartDiv + resCtrDiv + env + tags + succRate + execTime + timestamp},
	{"generate report overview without tags", reportOverviewTag, overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", 41, 2, 39, 0},
		chartDiv + resCtrDiv + env + succRate + execTime + timestamp},
}

func TestExecute(t *testing.T) {
	testReportGen(reportGenTests, t)
}

func testReportGen(reportGenTests []reportGenTest, t *testing.T) {
	buf := new(bytes.Buffer)
	for _, test := range reportGenTests {
		gen(test.tmpl, buf, test.input)

		got := removeNewline(buf.String())
		want := removeNewline(test.output)

		if got != want {
			t.Errorf("%s:\nwant:\n%q\ngot:\n%q\n", test.name, want, got)
		}
		buf.Reset()
	}
}

func removeNewline(s string) string {
	return re.ReplaceAllLiteralString(s, "")
}
