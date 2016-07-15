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

var wBodyHeader string = `<header class="top">
<div class="header">
  <div class="container">
     <div class="logo"><img src="images/logo.png" alt="Report logo"></div>
        <h2 class="project">Project: projname</h2>
      </div>
  </div>
</header>`

var wChartDiv string = `<div class="report-overview">
  <div class="report_chart">
    <div class="chart">
      <nvd3 options="options" data="data"></nvd3>
    </div>
    <div class="total-specs"><span class="value">41</span> <span class="txt">Total specs</span></div>
  </div>`

var wResCntDiv string = `
  <div class="report_test-results">
    <ul>
      <li class="fail"><span class="value">2</span> <span class="txt">Failed</span></li>
      <li class="pass"><span class="value">39</span> <span class="txt">Passed</span></li>
      <li class="skip"><span class="value">0</span> <span class="txt">Skipped</span></li>
    </ul>
  </div>`

var wEnvLi string = `<div class="report_details"><ul>
      <li>
        <label>Environment </label>
        <span>default</span>
      </li>`

var wTagsLi string = `
      <li>
        <label>Tags </label>
        <span>foo</span>
      </li>`

var wSuccRateLi string = `
      <li>
        <label>Success Rate </label>
        <span>34%</span>
      </li>`

var wExecTimeLi string = `
     <li>
        <label>Total Time </label>
        <span>00:01:53</span>
      </li>`

var wTimestampLi string = `
     <li>
        <label>Generated On </label>
        <span>Jun 3, 2016 at 12:29pm</span>
      </li>
    </ul>
  </div>
</div>`

var wSidebarAside string = `<aside class="sidebar">
  <h3 class="title">Specifications</h3>
  <div class="searchbar">
    <input id="searchSpecifications" placeholder="Type specification or tag name" type="text" />
    <i class="fa fa-search"></i>
  </div>
  <div id="listOfSpecifications">
    <ul id="scenarios" class="spec-list">
    <li class='passed spec-name'>
      <span id="scenarioName" class="scenarioname">Passing Spec</span>
      <span id="time" class="time">00:01:04</span>
    </li>
    <li class='failed spec-name'>
      <span id="scenarioName" class="scenarioname">Failing Spec</span>
      <span id="time" class="time">00:00:30</span>
    </li>
    <li class='skipped spec-name'>
      <span id="scenarioName" class="scenarioname">Skipped Spec</span>
      <span id="time" class="time">00:00:00</span>
    </li>
    </ul>
  </div>
</aside>`

var wCongratsDiv string = `<div class="congratulations details">
  <p>Congratulations! You've gone all <span class="green">green</span> and saved the environment!</p>
</div>`

var wHookFailureWithScreenhotDiv string = `<div class="error-container failed">
  <div collapsable class="error-heading">BeforeSuite Failed: <span class="error-message">SomeError</span></div>
  <div class="toggleShow" data-toggle="collapse" data-target="#hookFailureDetails">
    <span>[Show details]</span>
  </div>
  <div class="exception-container" id="hookFailureDetails">
      <div class="exception">
        <pre class="stacktrace">Stack trace</pre>
      </div>
      <div class="screenshot-container">
        <a href="data:image/png;base64,iVBO" rel="lightbox">
          <img ng-src="data:image/png;base64,iVBO" class="screenshot-thumbnail"/>
        </a>
      </div>
  </div>
</div>`

var wHookFailureWithoutScreenhotDiv string = `<div class="error-container failed">
  <div collapsable class="error-heading">BeforeSuite Failed: <span class="error-message">SomeError</span></div>
  <div class="toggleShow" data-toggle="collapse" data-target="#hookFailureDetails">
    <span>[Show details]</span>
  </div>
  <div class="exception-container" id="hookFailureDetails">
      <div class="exception">
        <pre class="stacktrace">Stack trace</pre>
      </div>
  </div>
</div>`

var wSpecHeaderStartWithTags string = `<header class="curr-spec">
  <h3 class="spec-head" title="/tmp/gauge/specs/foobar.spec">Spec heading</h3>
  <span class="time">00:01:01</span>`

var wTagsDiv string = `<div class="tags scenario_tags contentSection">
  <strong>Tags:</strong>
  <span> tag1</span>
  <span> tag2</span>
</div>`

var wTableTag string = `<table class="data-table">
  <tr>
    <th>Word</th>
    <th>Count</th>
  </tr>
  <tbody>
    <tr class='passed'>
      <td>Gauge</td>
      <td>3</td>
    </tr>
    <tr class='failed'>
      <td>Mingle</td>
      <td>2</td>
    </tr>
    <tr class='skipped'>
      <td>foobar</td>
      <td>1</td>
    </tr>
  </tbody>
</table>`

var re *regexp.Regexp = regexp.MustCompile("[ ]*[\n\t][ ]*")

var reportGenTests = []reportGenTest{
	{"generate body header with project name", bodyHeaderTag, &overview{ProjectName: "projname"}, wBodyHeader},
	{"generate report overview with tags", reportOverviewTag, &overview{"projname", "default", "foo", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", 41, 2, 39, 0},
		wChartDiv + wResCntDiv + wEnvLi + wTagsLi + wSuccRateLi + wExecTimeLi + wTimestampLi},
	{"generate report overview without tags", reportOverviewTag, &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", 41, 2, 39, 0},
		wChartDiv + wResCntDiv + wEnvLi + wSuccRateLi + wExecTimeLi + wTimestampLi},
	{"generate sidebar with appropriate pass/fail/skip class", sidebarDiv, &sidebar{
		IsPreHookFailure: false,
		Specs: []*specsMeta{
			newSpecsMeta("Passing Spec", "00:01:04", false, false, nil),
			newSpecsMeta("Failing Spec", "00:00:30", true, false, nil),
			newSpecsMeta("Skipped Spec", "00:00:00", false, true, nil),
		}}, wSidebarAside},
	{"do not generate sidebar if presuitehook failure", sidebarDiv, &sidebar{
		IsPreHookFailure: true,
		Specs:            []*specsMeta{},
	}, ""},
	{"generate congratulations bar if all specs are passed", congratsDiv, &overview{}, wCongratsDiv},
	{"don't generate congratulations bar if some spec failed", congratsDiv, &overview{Failed: 1}, ""},
	{"generate hook failure div with screenshot", hookFailureDiv, newHookFailure("BeforeSuite", "SomeError", "iVBO", "Stack trace"), wHookFailureWithScreenhotDiv},
	{"generate hook failure div without screenshot", hookFailureDiv, newHookFailure("BeforeSuite", "SomeError", "", "Stack trace"), wHookFailureWithoutScreenhotDiv},
	{"generate spec header with tags", specHeaderStartTag, &specHeader{"Spec heading", "00:01:01", "/tmp/gauge/specs/foobar.spec", []string{"foo", "bar"}}, wSpecHeaderStartWithTags},
	{"generate div for tags", tagsDiv, &specHeader{Tags: []string{"tag1", "tag2"}}, wTagsDiv},
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

func newHookFailure(name, errMsg, screenshot, stacktrace string) *hookFailure {
	return &hookFailure{
		HookName:   name,
		ErrMsg:     errMsg,
		Screenshot: screenshot,
		Stacktrace: stacktrace,
	}
}

func newOverview() *overview {
	return &overview{
		ProjectName: "gauge-testsss",
		Env:         "default",
		SuccRate:    95,
		ExecTime:    "00:01:53",
		Timestamp:   "Jun 3, 2016 at 12:29pm",
	}
}

func newSpecsMeta(name, execTime string, failed, skipped bool, tags []string) *specsMeta {
	return &specsMeta{
		SpecName: name,
		ExecTime: execTime,
		Failed:   failed,
		Skipped:  skipped,
		Tags:     tags,
	}
}

func newTable() *table {
	return &table{
		Headers: []string{"Word", "Count"},
		Rows: []*row{
			&row{
				Cells: []string{"Gauge", "3"},
				Res:   PASS,
			},
			&row{
				Cells: []string{"Mingle", "2"},
				Res:   FAIL,
			},
			&row{
				Cells: []string{"foobar", "1"},
				Res:   SKIP,
			},
		},
	}
}
