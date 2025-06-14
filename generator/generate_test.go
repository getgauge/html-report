/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package generator

import (
	"bytes"
	"fmt"
	"testing"

	"path/filepath"

	helper "github.com/getgauge/html-report/test_helper"
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
        <a href="."><img src="images/gaugeLogo.png" alt="Report logo"></a>
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
</div>`

var wResCntDiv = `
  <div class="report_test-results">
		<div class="report_test-result specs">
				<div class="total-specs" title="Filter all specs"><span class="txt">Total specs</span><span class="value">41</span></div>
				<div class="fail spec-filter" data-status="failed" title="Filter failed specs"><span class="value">2</span></div>
				<div class="pass spec-filter" data-status="passed" title="Filter passed specs"><span class="value">39</span></div>
				<div class="skip spec-filter" data-status="skipped" title="Filter skipped specs"><span class="value">0</span></div>
		</div>
		<div class="report_test-result scenarios">
				<div class="total-scenarios"><span class="txt">Total scenario</span><span class="value">41</span></div>
				<div class="fail scenario-stats" data-status="failed"><span class="value">2</span></div>
				<div class="pass scenario-stats" data-status="passed"><span class="value">39</span></div>
				<div class="skip scenario-stats" data-status="skipped"><span class="value">0</span></div>
		</div>
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
  <div class="specs-sorting">
			<div class="sort sort-specs-name" data-sort-by="specs-name"><span class="sort-icons"><i class="fa fa-caret-up"></i><i class="fa fa-caret-down"></i></span><span>Name</span></div>
			<div class="sort sort-execution-time" data-sort-by="execution-time"><span class="sort-icons"><i class="fa fa-caret-up"></i><i class="fa fa-caret-down"></i></span><span>Execution time</span></div>
	</div>
  <div id="listOfSpecifications">
    <ul id="scenarios" class="spec-list">
		<a href="passing_spec.html">
    	<li class="passed spec-name">
	      <span id="scenarioName" class="scenarioname">Passing Spec</span>
	      <span id="time" class="time">00:01:04</span>
    	</li>
		</a>
		<a href="failing_spec.html">
    	<li class="failed spec-name">
	      <span id="scenarioName" class="scenarioname">Failing Spec</span>
	      <span id="time" class="time">00:00:30</span>
    	</li>
		</a>
		<a href="skipped_spec.html">
    	<li class="skipped spec-name">
	      <span id="scenarioName" class="scenarioname">Skipped Spec</span>
	      <span id="time" class="time">00:00:00</span>
    	</li>
		</a>
    </ul>
  </div>
</aside>`

var wHookFailureWithScreenhotDiv = `<div class="error-container failed" data-tablerow='0'>
<div class="error-heading">BeforeSuite Failed:<span class="error-message"> SomeError</span></div>
  <div class="toggle-show">
    [Show details]
  </div>
  <div class="exception-container hidden">
		<div class="exception">
			<pre class="stacktrace">Stack trace</pre>
		</div>
		<div class="screenshot-container">
			<div class="screenshot">
				<a href="../images/iVBO" rel="lightbox">
					<img src="../images/iVBO" class="screenshot-thumbnail" />
				</a>
			</div>
		</div>
  </div>
</div>`

var wHookFailureWithoutScreenhotDiv = `<div class="error-container failed" data-tablerow='0'>
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

var wSpecCommentsWithTableTag = `
<span><p>This is an executable specification file. This file follows markdown syntax.</p>
<p>To execute this specification, run</p><pre><code>gauge specs</code></pre></span>
<table class="data-table">
  <tr>
    <th>Word</th>
    <th>Count</th>
  </tr>
  <tbody data-rowCount=3>
    <tr class="row-selector passed selected" data-rowIndex='0'>
      <td>Gauge</td>
      <td>3</td>
    </tr>
    <tr class="row-selector failed" data-rowIndex='1'>
      <td>Mingle</td>
      <td>2</td>
    </tr>
    <tr class="row-selector skipped" data-rowIndex='2'>
      <td>foobar</td>
      <td>1</td>
    </tr>
  </tbody>
</table>
<span><p>Comment 1</p>
<p>Comment 2</p>
<p>Comment 3</p></span>`

var wSpecCommentsWithoutTableTag = `
<span><p>This is an executable specification file. This file follows markdown syntax.</p>
<p>To execute this specification, run</p>
<pre><code>gauge specs</code></pre></span>
`

var wSpecCommentsWithCodeBlock = `<span><pre><code>{&#34;prop&#34;:&#34;value&#34;}</code></pre></span>`

var wScenarioContainerStartPassDiv = `<div class="scenario-container passed">`
var wScenarioContainerStartFailDiv = `<div class="scenario-container failed">`
var wScenarioContainerStartSkipDiv = `<div class="scenario-container skipped">`

var wscenarioHeaderStartDiv = `<div class="scenario-head">
  <h3 class="head borderBottom">Scenario Heading</h3>
  <span class="time">00:01:01</span>`

var wPassStepStartDiv = `<div class="step">
  <h5 class="execution-time"><span class="time">Execution Time : 00:03:31</span></h5>
  <div class="step-info passed">
    <ul>
      <li class="step">
        <div class="step-txt">`

var wFailStepStartDiv = `<div class="step">
  <h5 class="execution-time"><span class="time">Execution Time : 00:03:31</span></h5>
  <div class="step-info failed">
    <ul>
      <li class="step">
        <div class="step-txt">`

var wSkipStepStartDiv = `<div class="step">
  <div class="step-info skipped">
    <ul>
      <li class="step">
        <div class="step-txt">`

var wPassStepBodyDivWithBracketsInFragment = `
	<span>Say</span>
	<span class="parameter">"good &lt;a&gt; morning"</span>
	<span>to</span>
	<span class="parameter">"gauge"</span>
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
	  <div class="screenshot-container custom-screenshot-message">
		<div class="screenshot">
		 <p class="custom-message">To view a screenshot of this failed step, Please set up a <a href="https://docs.gauge.org/writing-specifications/#taking-custom-screenshots">custom screenshot handler.</a>
		 </p>
		</div>
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

var wBeforeSuiteMessageDiv = `<div class="suite_messages">
	<div>
		<div class="step-message"><p>Before Suite message</p></div>
	</div>
</div>`

var wAfterSuiteMessageDiv = `<div class="suite_messages">
	<div>
		<div class="step-message"><p>After Suite message</p></div>
	</div>
</div>`

var wBeforeAndAfterSuiteMessageDiv = `<div class="suite_messages">
	<div>
		<div class="step-message"><p>Before Suite message</p></div>
	</div>
	<div class="message_separator">--------</div>
	<div>
		<div class="step-message"><p>After Suite message</p></div>
	</div>
</div>`

var wBeforeSuiteScreenshotDiv = `<div class="suite_screenshots">
	<div>Before Suite Screenshots</div>
	<div class="screenshot-container">
		<div class="screenshot">
			<a href="../images/Before Suite Screenshot" rel="lightbox">
				<img src="../images/Before Suite Screenshot" class="screenshot-thumbnail" />
			</a>
		</div>
	</div>
</div>`

var wBeforeSuiteScreenshotBytesDiv = `<div class="suite_screenshots">
	<div>Before Suite Screenshots</div>
	<div class="screenshot-container">
		<div class="screenshot">
			<a href="data:image/png;base64,Before Suite Screenshot" rel="lightbox">
				<img src="data:image/png;base64,Before Suite Screenshot" class="screenshot-thumbnail" />
			</a>
		</div>
	</div>
</div>`

var wAfterSuiteScreenshotDiv = `<div class="suite_screenshots">
	<div>Before Suite Screenshots</div>
	<div class="screenshot-container">
		<div class="screenshot">
			<a href="../images/After Suite Screenshot" rel="lightbox">
				<img src="../images/After Suite Screenshot" class="screenshot-thumbnail" />
			</a>
		</div>
	</div>
</div>`

var wAfterSuiteScreenshotBytesDiv = `<div class="suite_screenshots">
	<div>After Suite Screenshots</div>
	<div class="screenshot-container">
		<div class="screenshot">
			<a href="data:image/png;base64,After Suite Screenshot" rel="lightbox">
				<img src="data:image/png;base64,After Suite Screenshot" class="screenshot-thumbnail" />
			</a>
		</div>
	</div>
</div>`

var wBeforeAndAfterSuiteScreenshotDiv = `<div class="suite_screenshots">
	<div>Before Suite Screenshots</div>
	<div class="screenshot-container">
		<div class="screenshot">
			<a href="../images/Before Suite Screenshot" rel="lightbox">
				<img src="../images/Before Suite Screenshot" class="screenshot-thumbnail" />
			</a>
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
	Result: &result{
		Status:        pass,
		ExecutionTime: "00:03:31",
	},
}

var stepWithCodeBlock = &spec{
	CommentsBeforeDatatable: `    {"prop":"value"}`,
}

var stepWithFileParam = &step{
	Fragments: []*fragment{
		{FragmentKind: textFragmentKind, Text: "Say "},
		{FragmentKind: specialStringFragmentKind, Text: "good morning", Name: "file:hello.txt"},
		{FragmentKind: textFragmentKind, Text: " to gauge"},
	},
	Result: &result{
		Status:        pass,
		ExecutionTime: "00:03:31",
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
						Cells:  []string{"Gauge", "3"},
						Result: pass,
					},
					{
						Cells:  []string{"Mingle", "2"},
						Result: fail,
					},
				},
			},
		},
		{FragmentKind: textFragmentKind, Text: " to gauge"},
	},
	Result: &result{
		Status:        pass,
		ExecutionTime: "00:03:31",
	},
}

var skippedStepRes = &result{
	Status:        skip,
	SkippedReason: "step impl not found",
}

var reportGenTests = []reportGenTest{
	{"generate html page start with project name", "htmlPageStartTag", &overview{ProjectName: "projname"}, whtmlPageStartTag},
	{"generate report overview with tags", "reportOverviewTag", &overview{"projname", "default", "foo", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{}, []string{}, []string{}, []string{}, []string{}, []string{}},
		wChartDiv + wResCntDiv + wEnvLi + wTagsLi + wSuccRateLi + wExecTimeLi + wTimestampLi},
	{"generate report overview without tags", "reportOverviewTag", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{}, []string{}, []string{}, []string{}, []string{}, []string{}},
		wChartDiv + wResCntDiv + wEnvLi + wSuccRateLi + wExecTimeLi + wTimestampLi},
	{"generate suite messages with before hook message", "suiteMessagesDiv", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{"Before Suite message"}, []string{}, []string{}, []string{}, []string{}, []string{}},
		wBeforeSuiteMessageDiv},
	{"generate suite messages with after hook message", "suiteMessagesDiv", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{}, []string{"After Suite message"}, []string{}, []string{}, []string{}, []string{}},
		wAfterSuiteMessageDiv},
	{"generate suite messages with before and after hook message", "suiteMessagesDiv", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{"Before Suite message"}, []string{"After Suite message"}, []string{}, []string{}, []string{}, []string{}},
		wBeforeAndAfterSuiteMessageDiv},
	{"generate suite screenshots with before hook screenshot", "suiteScreenshotsDiv", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{}, []string{}, []string{}, []string{}, []string{"Before Suite Screenshot"}, []string{}},
		wBeforeSuiteScreenshotDiv},
	{"generate suite screenshots with before hook screenshot bytes", "suiteScreenshotsDiv", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{}, []string{}, []string{"Before Suite Screenshot"}, []string{}, []string{}, []string{}},
		wBeforeSuiteScreenshotBytesDiv},
	{"generate suite screenshots with after hook screenshot", "suiteScreenshotsDiv", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{}, []string{}, []string{}, []string{}, []string{"After Suite Screenshot"}, []string{}},
		wAfterSuiteScreenshotDiv},
	{"generate suite screenshots with after hook screenshot bytes", "suiteScreenshotsDiv", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{}, []string{}, []string{}, []string{"After Suite Screenshot"}, []string{}, []string{}},
		wAfterSuiteScreenshotBytesDiv},
	{"generate suite screenshots with before and after hook screenshot", "suiteScreenshotsDiv", &overview{"projname", "default", "", 34, "00:01:53", "Jun 3, 2016 at 12:29pm", &summary{41, 2, 39, 0}, &summary{41, 2, 39, 0}, "../", []string{}, []string{}, []string{}, []string{}, []string{"Before Suite Screenshot"}, []string{}},
		wBeforeAndAfterSuiteScreenshotDiv},
	{"generate sidebar with appropriate pass/fail/skip class", "sidebarDiv", &sidebar{
		IsBeforeHookFailure: false,
		Specs: []*specsMeta{
			newSpecsMeta("Passing Spec", "00:01:04", false, false, nil, "passing_spec.html"),
			newSpecsMeta("Failing Spec", "00:00:30", true, false, nil, "failing_spec.html"),
			newSpecsMeta("Skipped Spec", "00:00:00", false, true, nil, "skipped_spec.html"),
		}}, wSidebarAside},
	{"do not generate sidebar if presuitehook failure", "sidebarDiv", &sidebar{
		IsBeforeHookFailure: true,
		Specs:               []*specsMeta{},
	}, ""},
	{"generate hook failure div with screenshot", "hookFailureDiv", newHookFailure("../", "BeforeSuite", "SomeError", "iVBO", "Stack trace"), wHookFailureWithScreenhotDiv},
	{"generate hook failure div without screenshot", "hookFailureDiv", newHookFailure("../", "BeforeSuite", "SomeError", "", "Stack trace"), wHookFailureWithoutScreenhotDiv},
	{"generate spec header with tags", "specHeaderStartTag", &specHeader{"Spec heading", "00:01:01", "/tmp/gauge/specs/foobar.spec", []string{"foo", "bar"}, &summary{0, 0, 0, 0}}, wSpecHeaderStartWithTags},
	{"generate div for tags", "tagsDiv", &specHeader{Tags: []string{"tag1", "tag2"}}, wTagsDiv},
	{"generate spec comments with data table (if present)", "specCommentsAndTableTag", newSpec(true), wSpecCommentsWithTableTag},
	{"generate spec comments without data table", "specCommentsAndTableTag", newSpec(false), wSpecCommentsWithoutTableTag},
	{"generate spec comments with code block", "specCommentsAndTableTag", stepWithCodeBlock, wSpecCommentsWithCodeBlock},
	{"generate passing scenario container", "scenarioContainerStartDiv", &scenario{ExecutionStatus: pass, TableRowIndex: -1}, wScenarioContainerStartPassDiv},
	{"generate failed scenario container", "scenarioContainerStartDiv", &scenario{ExecutionStatus: fail, TableRowIndex: -1}, wScenarioContainerStartFailDiv},
	{"generate skipped scenario container", "scenarioContainerStartDiv", &scenario{ExecutionStatus: skip, TableRowIndex: -1}, wScenarioContainerStartSkipDiv},
	{"generate scenario header", "scenarioHeaderStartDiv", &scenario{Heading: "Scenario Heading", ExecutionTime: "00:01:01"}, wscenarioHeaderStartDiv},
	{"generate pass step start div", "stepStartDiv", newStep(pass), wPassStepStartDiv},
	{"generate fail step start div", "stepStartDiv", newStep(fail), wFailStepStartDiv},
	{"generate skipped step start div", "stepStartDiv", newStep(skip), wSkipStepStartDiv},
	{"generate skipped step body div", "stepBodyDiv", stepWithBracketsInFragment, wPassStepBodyDivWithBracketsInFragment},
	{"generate skipped step skipped reason div", "skippedReasonDiv", skippedStepRes, wSkippedStepWithSkippedReason},
	{"generate step body div with file special param", "stepBodyDiv", stepWithFileParam, wStepWithFileParam},
	{"generate step body div with special table param", "stepBodyDiv", stepWithSpecialTableParam, wStepWithSpecialTableParam},
	{"generate step failure div", "stepFailureDiv", &result{ErrorMessage: "expected:<foo [foo] foo> but was:<foo [bar] foo>", StackTrace: "stacktrace"}, wStepFailDiv},
	{"generate spec error div", "specErrorDiv", &spec{Errors: []buildError{{ErrorType: parseErrorType, Message: "message"}}}, wSpecErrorDiv},
}

func TestExecute(t *testing.T) {
	helper.SetEnvOrFail(t, "screenshot_on_failure", "true")
	testReportGen(reportGenTests, t)
}

func testReportGen(reportGenTests []reportGenTest, t *testing.T) {
	buf := new(bytes.Buffer)
	for _, test := range reportGenTests {
		execTemplate(test.tmpl, buf, test.input)

		got := helper.RemoveNewline(buf.String())
		want := helper.RemoveNewline(test.output)

		if got != want {
			t.Errorf("%s:\nwant:\n% q\ngot:\n%q\n", test.name, want, got)
		}
		buf.Reset()
	}
}

func newHookFailure(basePath, name, errMsg, screenshot, stacktrace string) *hookFailure {
	return &hookFailure{
		BasePath:              basePath,
		HookName:              name,
		ErrMsg:                errMsg,
		FailureScreenshotFile: screenshot,
		StackTrace:            stacktrace,
	}
}

func newSpecsMeta(name, execTime string, failed, skipped bool, tags []string, fileName string) *specsMeta {
	return &specsMeta{
		SpecName:      name,
		ExecutionTime: execTime,
		Failed:        failed,
		Skipped:       skipped,
		Tags:          tags,
		ReportFile:    fileName,
	}
}

func newSpec(withTable bool) *spec {
	t := &table{
		Headers: []string{"Word", "Count"},
		Rows: []*row{
			{
				Cells:  []string{"Gauge", "3"},
				Result: pass,
			},
			{
				Cells:  []string{"Mingle", "2"},
				Result: fail,
			},
			{
				Cells:  []string{"foobar", "1"},
				Result: skip,
			},
		},
	}

	c1 := "This is an executable specification file. This file follows markdown syntax.\n\nTo execute this specification, run\n\n\tgauge specs\n"
	c2 := "\nComment 1\n\nComment 2\n\nComment 3"

	if withTable {
		return &spec{
			CommentsBeforeDatatable: c1,
			Datatable:               t,
			CommentsAfterDatatable:  c2,
		}
	}

	return &spec{
		CommentsBeforeDatatable: c1,
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
		Result: &result{
			Status:        s,
			ExecutionTime: "00:03:31",
		},
	}
}

func TestGetAbsThemePathForRelPath(t *testing.T) {
	oldProjectRoot := projectRoot
	projectRoot, _ = filepath.Abs(filepath.Join("Dummy", "Project", "Root"))
	themePath := filepath.Join("some", "path")
	want := filepath.Join(projectRoot, themePath)

	got := getAbsThemePath(themePath)

	if want != got {
		t.Errorf("Expected theme path = %s, got %s", want, got)
	}
	projectRoot = oldProjectRoot
}

func BenchmarkGenerateReport(b *testing.B) {
	ps := &SuiteResult{
		ProjectName: "Foo",
		SpecResults: []*spec{},
	}

	for i := 0; i < b.N; i++ {
		s := newSpec(false)
		s.FileName = fmt.Sprintf("example%d.spec", i)
		ps.SpecResults = append(ps.SpecResults, s)
	}
	GenerateReport(ps, filepath.Join("_testdata", "benchmark"), filepath.Join("_testdata", "dummyReportTheme"), false)
}

func newStepWithHookFailures() *step {
	s := newStep(fail)
	s.AfterStepHookFailure = &hookFailure{}
	s.BeforeStepHookFailure = &hookFailure{}
	return s
}

var stepItem = item{Kind: stepKind, Step: newStepWithHookFailures()}
var basePathSeedSpec = func() *spec {
	return &spec{
		FileName:               filepath.Join("some", "base", "path", "example.spec"),
		BeforeSpecHookFailures: []*hookFailure{{}},
		AfterSpecHookFailures:  []*hookFailure{{}},
		Scenarios: []*scenario{
			{
				AfterScenarioHookFailure:  &hookFailure{},
				BeforeScenarioHookFailure: &hookFailure{},
				Contexts: []item{
					stepItem,
					{Kind: conceptKind, Concept: &concept{
						Items:       []item{stepItem},
						ConceptStep: newStepWithHookFailures()},
					},
				},
				Teardowns: []item{
					stepItem,
					{Kind: conceptKind, Concept: &concept{
						Items:       []item{stepItem},
						ConceptStep: newStepWithHookFailures()},
					},
				},
				Items: []item{
					stepItem,
					{Kind: conceptKind, Concept: &concept{
						Items:       []item{stepItem},
						ConceptStep: newStepWithHookFailures()},
					},
				},
			},
		},
	}
}

type basePathPropogationTest struct {
	name     string
	expected string
	spec     *spec
	actual   func(s *spec) string
}

func (b basePathPropogationTest) getActual() string {
	return b.actual(b.spec)
}

func TestSpecBasepathPropogation(t *testing.T) {
	bp := filepath.Join("..", "..", "..")
	var basePathPropogationTests = []basePathPropogationTest{
		{name: "spec.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.BasePath }},
		{name: "spec.beforehookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.BeforeSpecHookFailures[0].BasePath }},
		{name: "spec.afterhookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.AfterSpecHookFailures[0].BasePath }},
		{name: "spec.scenario.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].BasePath }},
		{name: "spec.scenario.beforehookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].BeforeScenarioHookFailure.BasePath }},
		{name: "spec.scenario.afterhookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].AfterScenarioHookFailure.BasePath }},
		{name: "spec.scenario.context.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Contexts[0].Step.BasePath }},
		{name: "spec.scenario.context.beforehookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Contexts[0].Step.BeforeStepHookFailure.BasePath }},
		{name: "spec.scenario.context.afterhookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Contexts[0].Step.AfterStepHookFailure.BasePath }},
		{name: "spec.scenario.context.concept.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Contexts[1].Concept.ConceptStep.BasePath }},
		{name: "spec.scenario.context.concept.beforehookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string {
			return s.Scenarios[0].Contexts[1].Concept.ConceptStep.BeforeStepHookFailure.BasePath
		}},
		{name: "spec.scenario.context.concept.afterhookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string {
			return s.Scenarios[0].Contexts[1].Concept.ConceptStep.AfterStepHookFailure.BasePath
		}},
		{name: "spec.scenario.teardown.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Teardowns[0].Step.BasePath }},
		{name: "spec.scenario.teardown.beforehookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Teardowns[0].Step.BeforeStepHookFailure.BasePath }},
		{name: "spec.scenario.teardown.afterhookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Teardowns[0].Step.BeforeStepHookFailure.BasePath }},
		{name: "spec.scenario.teardown.concept.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Teardowns[1].Concept.ConceptStep.BasePath }},
		{name: "spec.scenario.teardown.concept.beforehookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string {
			return s.Scenarios[0].Teardowns[1].Concept.ConceptStep.BeforeStepHookFailure.BasePath
		}},
		{name: "spec.scenario.teardown.concept.afterhookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string {
			return s.Scenarios[0].Teardowns[1].Concept.ConceptStep.BeforeStepHookFailure.BasePath
		}},
		{name: "spec.scenario.step.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Items[0].Step.BasePath }},
		{name: "spec.scenario.step.beforehookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].AfterScenarioHookFailure.BasePath }},
		{name: "spec.scenario.step.afterhookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].AfterScenarioHookFailure.BasePath }},
		{name: "spec.scenario.concept.step.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Items[1].Concept.ConceptStep.BasePath }},
		{name: "spec.scenario.concept.step.beforehookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string {
			return s.Scenarios[0].Items[1].Concept.ConceptStep.BeforeStepHookFailure.BasePath
		}},
		{name: "spec.scenario.concept.step.afterhookfailure.basepath", expected: bp, spec: basePathSeedSpec(), actual: func(s *spec) string { return s.Scenarios[0].Items[1].Concept.ConceptStep.AfterStepHookFailure.BasePath }},
	}

	for _, tt := range basePathPropogationTests {
		t.Run(tt.name, func(t *testing.T) {
			propogateBasePath(tt.spec)
			a := tt.getActual()
			if filepath.Clean(a) != filepath.Clean(tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected, a)
			}
		})
	}
}
