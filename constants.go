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

package main

const htmlStartTag = `<!doctype html>
<html>`

const htmlEndTag = `</html>`

const header = `<head>
<meta http-equiv="X-UA-Compatible" content="IE=9; IE=8; IE=7; IE=EDGE"/>
<title>Gauge Test Results</title>
<link rel="shortcut icon" type="image/x-icon" href="images/favicon.ico">
<link rel="stylesheet" type="text/css" href="css/open-sans.css">
<link rel="stylesheet" type="text/css" href="css/font-awesome.css">
<link rel="stylesheet" type="text/css" href="css/normalize.css"/>
<link rel="stylesheet" type="text/css" href="css/angular-hovercard.css"/>
<link rel="stylesheet" type="text/css" href="css/style.css"/>
</head>`

const bodyStartTag = `<body>
`

const bodyEndTag = `</body>`

const bodyHeader = `
<header class="top">
  <div class="header">
    <div class="container">
      <div class="logo"><img src="images/logo.png" alt="Report logo"></div>
      <h2 class="project">Project: gauge-tests</h2>
    </div>
  </div>
</header>`

const mainStartTag = `<main class="main-container">`

const mainEndTag = `</main>`

const containerStartTag = `<div class="container">`

const containerEndTag = `</div>`

const reportOverviewTag = `<div class="report-overview">
  <div class="report_chart">
    <div class="chart">
      <nvd3 options="options" data="data"></nvd3>
    </div>
    <div class="total-specs"><span class="value">{{.TotalSpecs}}</span> <span class="txt">Total specs</span></div>
  </div>
  <div class="report_test-results">
    <ul>
      <li class="fail"><span class="value">{{.Failed}}</span> <span class="txt">Failed</span></li>
      <li class="pass"><span class="value">{{.Passed}}</span> <span class="txt">Passed</span></li>
      <li class="skip"><span class="value">{{.Skipped}}</span> <span class="txt">Skipped</span></li>
    </ul>
  </div>
  <div class="report_details">
    <ul>
      <li>
        <label>Environment </label>
        <span>{{.Env}}</span>
      </li>
      {{if .Tags}}
      <li>
        <label>Tags </label>
        <span>{{.Tags}}</span>
      </li>
      {{end}}
      <li>
        <label>Success Rate </label>
        <span>{{.SuccRate}}%</span>
      </li>
      <li>
        <label>Total Time </label>
        <span>{{.ExecTime}}</span>
      </li>
      <li>
        <label>Generated On </label>
        <span>{{.Timestamp}}</span>
      </li>
    </ul>
  </div>
</div>
`
