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

import (
	"io"
	"log"
	"os"
	"text/template"
)

func gen(tmplName string, f io.Writer, data interface{}) {
	tmpl, err := template.New("Reports").Parse(tmplName)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = tmpl.Execute(f, data)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

type overview struct {
	Env        string
	Tags       string
	SuccRate   string
	ExecTime   string
	Timestamp  string
	TotalSpecs int
	Failed     int
	Passed     int
	Skipped    int
}

func genOverview(f io.Writer) {
	o := &overview{
		Env:       "default",
		SuccRate:  "95",
		ExecTime:  "00:01:53",
		Timestamp: "Jun 3, 2016 at 12:29pm",
	}
	gen(reportOverviewTag, f, o)
}

func generate() {
	f, err := os.Create("report-template/index2.html")
	if err != nil {
		log.Fatalf(err.Error())
	}
	gen(htmlStartTag, f, nil)
	gen(header, f, nil)
	gen(bodyStartTag, f, nil)
	gen(bodyHeader, f, nil)
	gen(mainStartTag, f, nil)
	gen(containerStartTag, f, nil)
	genOverview(f)
	gen(containerEndTag, f, nil)
	gen(mainEndTag, f, nil)
	gen(bodyEndTag, f, nil)
	gen(htmlEndTag, f, nil)
}
