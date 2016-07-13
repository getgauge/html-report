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
	"text/template"
)

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
	Tags     []string
}

type sidebar struct {
	IsPreHookFailure bool
	Specs            []*specsMeta
}

type hookFailure struct {
	HookName   string
	ErrMsg     string
	Screenshot string
	Stacktrace string
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

func gen(tmplName string, w io.Writer, data interface{}) {
	tmpl, err := template.New("Reports").Parse(tmplName)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func generate(w io.Writer) {
	gen(htmlStartTag, w, nil)
	gen(headerTag, w, nil)
	gen(bodyStartTag, w, nil)
	gen(bodyEndTag, w, nil)
	gen(htmlEndTag, w, nil)
}
