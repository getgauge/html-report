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

var whtmlPageStartTag = `<!doctype html>
<html><head>
  <meta http-equiv="X-UA-Compatible" content="IE=9; IE=8; IE=7; IE=EDGE" />
  <meta charset="utf-8" />
  <title>Gauge Test Results</title>
  <link rel="shortcut icon" type="image/x-icon" href="images/favicon.ico">
  <link rel="stylesheet" type="text/css" href="css/open-sans.css">
  <link rel="stylesheet" type="text/css" href="css/font-awesome.css">
  <link rel="stylesheet" type="text/css" href="css/normalize.css" />
  <link rel="stylesheet" type="text/css" href="css/style.css" />
</head>
<body>
<header class="top">
  <div class="header">
    <div class="container">
      <div class="logo">
        <a href=""><img src="images/logo.png" alt="Report logo"></a>
      </div>
      <h2 class="project">Project: projname</h2>
    </div>
  </div>
</header>
<main class="main-container">
<div class="container">`

var wChartDiv = `<div class="report-overview">
<div class="report_chart">
<div class="chart">
<svg id="pie-chart" data-results="2,39,0" data-total="41">
  	<path class="status failed" />
	<path class="shadow failed" data-status="failed"><title>Failed: 2/41</title></path>
	<path class="status passed" />
	<path class="shadow passed" data-status="passed"><title>Passed: 39/41</title></path>
	<path class="status skipped" />
	<path class="shadow skipped" data-status="skipped"><title>Skipped: 0/41</title></path>
</svg>
</div>
<div class="total-specs"><span class="value">41</span><span class="txt">Total specs</span></div>
</div>`

var wResCntDiv = `
  <div class="report_test-results">
    <ul>
      <li class="fail spec-filter" data-status="failed"><span class="value">2</span><span class="txt">Failed</span></li>
      <li class="pass spec-filter" data-status="passed"><span class="value">39</span><span class="txt">Passed</span></li>
      <li class="skip spec-filter" data-status="skipped"><span class="value">0</span><span class="txt">Skipped</span></li>
    </ul>
  </div>`

var wEnvLi = `<div class="report_details"><ul>
      <li>
        <label>Environment </label>
        <span>default</span>
      </li>`

var wTagsLi = `
      <li>
        <label>Tags </label>
        <span>foo</span>
      </li>`

var wSuccRateLi = `
      <li>
        <label>Success Rate </label>
        <span>34%</span>
      </li>`

var wExecTimeLi = `
     <li>
        <label>Total Time </label>
        <span>00:01:53</span>
      </li>`

var wTimestampLi = `
     <li>
        <label>Generated On </label>
        <span>Jun 3, 2016 at 12:29pm</span>
      </li>
    </ul>
  </div>
</div>`

var wSidebarAside = `<aside class="sidebar">
  <h3 class="title">Specifications</h3>
  <div class="searchbar">
    <input id="searchSpecifications" placeholder="Type specification or tag name" type="text" />
    <i class="fa fa-search"></i>
  </div>
  <div id="listOfSpecifications">
    <ul id="scenarios" class="spec-list">
		<a href="passing_spec.html">
    	<li class='passed spec-name'>
	      <span id="scenarioName" class="scenarioname">Passing Spec</span>
	      <span id="time" class="time">00:01:04</span>
    	</li>
		</a>
		<a href="failing_spec.html">
    	<li class='failed spec-name'>
	      <span id="scenarioName" class="scenarioname">Failing Spec</span>
	      <span id="time" class="time">00:00:30</span>
    	</li>
		</a>
		<a href="skipped_spec.html">
    	<li class='skipped spec-name'>
	      <span id="scenarioName" class="scenarioname">Skipped Spec</span>
	      <span id="time" class="time">00:00:00</span>
    	</li>
		</a>
    </ul>
  </div>
</aside>`

var wHookFailureWithScreenhotDiv = `<div class="error-container failed">
<div class="error-heading">BeforeSuite Failed:<span class="error-message"> SomeError</span></div>
  <div class="toggle-show">
    [Show details]
  </div>
  <div class="exception-container hidden">
      <div class="exception">
        <pre class="stacktrace">Stack trace</pre>
      </div>
      <div class="screenshot-container">
        <a href="data:image/png;base64,iVBO" rel="lightbox">
          <img src="data:image/png;base64,iVBO" class="screenshot-thumbnail" />
        </a>
      </div>
  </div>
</div>`

var wHookFailureWithoutScreenhotDiv = `<div class="error-container failed">
  <div class="error-heading">BeforeSuite Failed:<span class="error-message"> SomeError</span></div>
  <div class="toggle-show">
    [Show details]
  </div>
  <div class="exception-container hidden">
      <div class="exception">
        <pre class="stacktrace">Stack trace</pre>
      </div>
  </div>
</div>`

var wSpecHeaderStartWithTags = `<div id="specificationContainer" class="details">
<header class="curr-spec">
	<div class="spec-head-wrapper">
		<h3 class="spec-head" title="/tmp/gauge/specs/foobar.spec">Spec heading</h3>
    <div class="hidden report_test-results" alt="Scenarios" title="Scenarios">
      <ul>
        <li class="fail"><span class="value">0</span><span class="txt">Failed</span></li>
        <li class="pass"><span class="value">0</span><span class="txt">Passed</span></li>
        <li class="skip"><span class="value">0</span><span class="txt">Skipped</span></li>
      </ul>
    </div>
	</div>
  <div class="spec-meta">
		<div class="spec-filename">
			<label for="specFileName">File Path</label>
			<input id="specFileName" value="/tmp/gauge/specs/foobar.spec" readonly/>
			<button class="clipboard-btn" data-clipboard-target="#specFileName" title="Copy to Clipboard">
				<i class="fa fa-clipboard" aria-hidden="true" title="Copy to Clipboard"></i>
			</button>
		</div>
		<span class="time">00:01:01</span>
	</div>`

var wTagsDiv = `<div class="tags scenario_tags contentSection">
  <strong>Tags:</strong>
  <span> tag1</span>
  <span> tag2</span>
</div>`

var wSpecCommentsWithTableTag = `<span></span>
<span><p>This is an executable specification file. This file follows markdown syntax.</p></span>
<span></span>
<span><p>To execute this specification, run</p></span><span><pre><code>gauge specs</code></pre></span>
<span></span>
<table class="data-table">
  <tr>
    <th>Word</th>
    <th>Count</th>
  </tr>
  <tbody data-rowCount=3>
    <tr class='row-selector passed selected' data-rowIndex='0'>
      <td>Gauge</td>
      <td>3</td>
    </tr>
    <tr class='row-selector failed' data-rowIndex='1'>
      <td>Mingle</td>
      <td>2</td>
    </tr>
    <tr class='row-selector skipped' data-rowIndex='2'>
      <td>foobar</td>
      <td>1</td>
    </tr>
  </tbody>
</table>
<span><p>Comment 1</p></span>
<span><p>Comment 2</p></span>
<span><p>Comment 3</p></span>`

var wSpecCommentsWithoutTableTag = `<span></span>
<span><p>This is an executable specification file. This file follows markdown syntax.</p></span><span></span>
<span><p>To execute this specification, run</p></span>
<span><pre><code>gauge specs</code></pre></span>
<span></span>`

var wScenarioContainerStartPassDiv = `<div class='scenario-container passed'>`
var wScenarioContainerStartFailDiv = `<div class='scenario-container failed'>`
var wScenarioContainerStartSkipDiv = `<div class='scenario-container skipped'>`

var wscenarioHeaderStartDiv = `<div class="scenario-head">
  <h3 class="head borderBottom">Scenario Heading</h3>
  <span class="time">00:01:01</span>`

var wPassStepStartDiv = `<div class='step'>
  <h5 class='execution-time'><span class='time'>Execution Time : 00:03:31</span></h5>
  <div class='step-info passed'>
    <ul>
      <li class='step'>
        <div class='step-txt'>`

var wFailStepStartDiv = `<div class='step'>
  <h5 class='execution-time'><span class='time'>Execution Time : 00:03:31</span></h5>
  <div class='step-info failed'>
    <ul>
      <li class='step'>
        <div class='step-txt'>`

var wSkipStepStartDiv = `<div class='step'>
  <div class='step-info skipped'>
    <ul>
      <li class='step'>
        <div class='step-txt'>`

var wStepEndDiv = `<span>Say</span><span class='parameter'>"hi"</span><span>to</span><span class='parameter'>"gauge"</span>
          <div class='inline-table'>
            <div>
              <table>
                <tr>
                  <th>Word</th>
                  <th>Count</th>
                </tr>
                <tbody>
                  <tr>
                    <td>Gauge</td>
                    <td>3</td>
                  </tr>
                  <tr>
                    <td>Mingle</td>
                    <td>2</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </li>
    </ul>
  </div>
</div>
`

var wPassStepBodyDivWithBracketsInFragment = `
	<span>Say</span>
	<span class='parameter'>"good &lt;a&gt; morning"</span>
	<span>to</span>
	<span class='parameter'>"gauge"</span>
</div>`

var wSkippedStepWithSkippedReason = `<div class="message-container">
  <h4 class="skipReason">Skipped Reason: step impl not found</h4>
</div>`

var wStepWithFileParam = `
    <span>Say</span>
		<span class="modal-link">&lt;file:hello.txt&gt;</span>
		<div class="modal">
			<h2 class="modal-title"></h2>
			<span class="close">&times;</span>
			<div class="modal-content">
					<pre>good morning</pre>
			</div>
			</div>
			<span>to gauge</span>
		</div>`

var wStepWithSpecialTableParam = `
    <span>Say</span>
		<span class="modal-link">&lt;table:hello.csv&gt;</span>
		<div class="modal">
			<h2 class="modal-title"></h2>
			<span class="close">&times;</span>
			<div class="modal-content">
				<pre></pre>
			</div>
		</div>
		<span>to gauge</span>
		</div>`

var wStepFailDiv = `<div class="error-container failed">
  <div class="exception-container">
      <div class="exception">
        <h4 class="error-message">
          <pre>expected:&lt;foo [foo] foo&gt; but was:&lt;foo [bar] foo&gt;</pre>
        </h4>
        <pre class="stacktrace">stacktrace</pre>
      </div>
  </div>
</div>`

var wSpecErrorDiv = `<div class="error-container failed">
  <div class="error-heading">Errors:</div>
  <div class="exception-container">
      <div class="exception">
        <pre class="error">[Parse Error] message</pre>
      </div>
   </div>
</div>`

var stepWithBracketsInFragment = &step{
	Fragments: []*fragment{
		{FragmentKind: textFragmentKind, Text: "Say "},
		{FragmentKind: staticFragmentKind, Text: "good <a> morning"},
		{FragmentKind: textFragmentKind, Text: " to "},
		{FragmentKind: dynamicFragmentKind, Text: "gauge"},
	},
	Res: &result{
		Status:   pass,
		ExecTime: "00:03:31",
	},
}

var stepWithFileParam = &step{
	Fragments: []*fragment{
		{FragmentKind: textFragmentKind, Text: "Say "},
		{FragmentKind: specialStringFragmentKind, Text: "good morning", Name: "file:hello.txt"},
		{FragmentKind: textFragmentKind, Text: " to gauge"},
	},
	Res: &result{
		Status:   pass,
		ExecTime: "00:03:31",
	},
}

var stepWithSpecialTableParam = &step{
	Fragments: []*fragment{
		{FragmentKind: textFragmentKind, Text: "Say "},
		{FragmentKind: specialTableFragmentKind, Name: "table:hello.csv",
			Table: &table{
				Headers: []string{"Word", "Count"},
				Rows: []*row{
					{
						Cells: []string{"Gauge", "3"},
						Res:   pass,
					},
					{
						Cells: []string{"Mingle", "2"},
						Res:   fail,
					},
				},
			},
		},
		{FragmentKind: textFragmentKind, Text: " to gauge"},
	},
	Res: &result{
		Status:   pass,
		ExecTime: "00:03:31",
	},
}

var skippedStepRes = &result{
	Status:        skip,
	SkippedReason: "step impl not found",
}

var re = regexp.MustCompile("[\\s]*[\n\t][\\s]*")

var reportGenTests = []reportGenTest{
	{"generate html page start with project name", htmlPageStartTag, &overview{ProjectName: "projname"}, whtmlPageStartTag},
	{"generate report overview with tags", reportOverviewTag, &overview{"projname", "default", "foo", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, "/"},
		wChartDiv + wResCntDiv + wEnvLi + wTagsLi + wSuccRateLi + wExecTimeLi + wTimestampLi},
	{"generate report overview without tags", reportOverviewTag, &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, "/"},
		wChartDiv + wResCntDiv + wEnvLi + wSuccRateLi + wExecTimeLi + wTimestampLi},
	{"generate sidebar with appropriate pass/fail/skip class", sidebarDiv, &sidebar{
		IsBeforeHookFailure: false,
		Specs: []*specsMeta{
			newSpecsMeta("Passing Spec", "00:01:04", false, false, nil, "passing_spec.html"),
			newSpecsMeta("Failing Spec", "00:00:30", true, false, nil, "failing_spec.html"),
			newSpecsMeta("Skipped Spec", "00:00:00", false, true, nil, "skipped_spec.html"),
		}}, wSidebarAside},
	{"do not generate sidebar if presuitehook failure", sidebarDiv, &sidebar{
		IsBeforeHookFailure: true,
		Specs:               []*specsMeta{},
	}, ""},
	{"generate hook failure div with screenshot", hookFailureDiv, newHookFailure("BeforeSuite", "SomeError", "iVBO", "Stack trace"), wHookFailureWithScreenhotDiv},
	{"generate hook failure div without screenshot", hookFailureDiv, newHookFailure("BeforeSuite", "SomeError", "", "Stack trace"), wHookFailureWithoutScreenhotDiv},
	{"generate spec header with tags", specHeaderStartTag, &specHeader{"Spec heading", "00:01:01", "/tmp/gauge/specs/foobar.spec", []string{"foo", "bar"}, &summary{0, 0, 0, 0}}, wSpecHeaderStartWithTags},
	{"generate div for tags", tagsDiv, &specHeader{Tags: []string{"tag1", "tag2"}}, wTagsDiv},
	{"generate spec comments with data table (if present)", specCommentsAndTableTag, newSpec(true), wSpecCommentsWithTableTag},
	{"generate spec comments without data table", specCommentsAndTableTag, newSpec(false), wSpecCommentsWithoutTableTag},
	{"generate passing scenario container", scenarioContainerStartDiv, &scenario{ExecStatus: pass, TableRowIndex: -1}, wScenarioContainerStartPassDiv},
	{"generate failed scenario container", scenarioContainerStartDiv, &scenario{ExecStatus: fail, TableRowIndex: -1}, wScenarioContainerStartFailDiv},
	{"generate skipped scenario container", scenarioContainerStartDiv, &scenario{ExecStatus: skip, TableRowIndex: -1}, wScenarioContainerStartSkipDiv},
	{"generate scenario header", scenarioHeaderStartDiv, &scenario{Heading: "Scenario Heading", ExecTime: "00:01:01"}, wscenarioHeaderStartDiv},
	{"generate pass step start div", stepStartDiv, newStep(pass), wPassStepStartDiv},
	{"generate fail step start div", stepStartDiv, newStep(fail), wFailStepStartDiv},
	{"generate skipped step start div", stepStartDiv, newStep(skip), wSkipStepStartDiv},
	{"generate skipped step body div", stepBodyDiv, stepWithBracketsInFragment, wPassStepBodyDivWithBracketsInFragment},
	{"generate skipped step skipped reason div", skippedReasonDiv, skippedStepRes, wSkippedStepWithSkippedReason},
	{"generate step body div with file special param", stepBodyDiv, stepWithFileParam, wStepWithFileParam},
	{"generate step body div with special table param", stepBodyDiv, stepWithSpecialTableParam, wStepWithSpecialTableParam},
	{"generate step failure div", stepFailureDiv, &result{ErrorMessage: "expected:<foo [foo] foo> but was:<foo [bar] foo>", StackTrace: "stacktrace"}, wStepFailDiv},
	{"generate spec error div", specErrorDiv, &spec{Errors: []error{buildError{ErrorType: parseError, Message: "message"}}}, wSpecErrorDiv},
}

func TestExecute(t *testing.T) {
	testReportGen(reportGenTests, t)
}

func testReportGen(reportGenTests []reportGenTest, t *testing.T) {
	buf := new(bytes.Buffer)
	for _, test := range reportGenTests {
		execTemplate(test.tmpl, buf, test.input)

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
		StackTrace: stacktrace,
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

func newSpecsMeta(name, execTime string, failed, skipped bool, tags []string, fileName string) *specsMeta {
	return &specsMeta{
		SpecName:   name,
		ExecTime:   execTime,
		Failed:     failed,
		Skipped:    skipped,
		Tags:       tags,
		ReportFile: fileName,
	}
}

func newSpec(withTable bool) *spec {
	t := &table{
		Headers: []string{"Word", "Count"},
		Rows: []*row{
			{
				Cells: []string{"Gauge", "3"},
				Res:   pass,
			},
			{
				Cells: []string{"Mingle", "2"},
				Res:   fail,
			},
			{
				Cells: []string{"foobar", "1"},
				Res:   skip,
			},
		},
	}

	c1 := []string{"\n", "This is an executable specification file. This file follows markdown syntax.", "\n", "To execute this specification, run", "\tgauge specs", "\n"}
	c2 := []string{"Comment 1", "Comment 2", "Comment 3"}

	if withTable {
		return &spec{
			CommentsBeforeTable: c1,
			Table:               t,
			CommentsAfterTable:  c2,
		}
	}

	return &spec{
		CommentsBeforeTable: c1,
	}
}

func newStep(s status) *step {
	return &step{
		Fragments: []*fragment{
			{FragmentKind: textFragmentKind, Text: "Say "},
			{FragmentKind: staticFragmentKind, Text: "hi"},
			{FragmentKind: textFragmentKind, Text: " to "},
			{FragmentKind: dynamicFragmentKind, Text: "gauge"},
			{FragmentKind: tableFragmentKind,
				Table: &table{
					Headers: []string{"Word", "Count"},
					Rows: []*row{
						{Cells: []string{"Gauge", "3"}},
						{Cells: []string{"Mingle", "2"}},
					},
				},
			},
		},
		Res: &result{
			Status:   s,
			ExecTime: "00:03:31",
		},
	}
}
