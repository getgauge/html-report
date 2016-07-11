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
	ProjectName string
	Env         string
	Tags        string
	SuccRate    float32
	ExecTime    string
	Timestamp   string
	TotalSpecs  int
	Failed      int
	Passed      int
	Skipped     int
}

type specsMeta struct {
	SpecName string
	ExecTime string
	Failed   bool
	Skipped  bool
}

type sidebar struct {
	IsPreHookFailure bool
	Specs            []*specsMeta
}

func generate() {
	f, err := os.Create("report-template/index2.html")
	if err != nil {
		log.Fatalf(err.Error())
	}
	o := newOverview()
	gen(htmlStartTag, f, nil)
	gen(headerTag, f, nil)
	gen(bodyStartTag, f, nil)
	gen(bodyHeaderTag, f, o)
	gen(mainStartTag, f, nil)
	gen(containerStartTag, f, nil)
	gen(reportOverviewTag, f, o)
	gen(containerEndTag, f, nil)
	gen(mainEndTag, f, nil)
	gen(bodyEndTag, f, nil)
	gen(htmlEndTag, f, nil)
}
