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

	"github.com/getgauge/html-report/gauge_messages"
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

type specHeader struct {
	SpecName string
	ExecTime string
	FileName string
	Tags     []string
}

type row struct {
	Cells []string
	Res   status
}

type table struct {
	Headers []string
	Rows    []*row
}

type spec struct {
	CommentsBeforeTable []string
	Table               *table
	CommentsAfterTable  []string
	Scenarios           []*scenario
}

type scenario struct {
	Heading  string
	ExecTime string
	Tags     []string
	Res      status
	Contexts []item
	Items    []item
	TearDown []item
}

const (
	stepKind kind = iota
	commentKind
	conceptKind
)

type kind int

type item interface {
	kind() kind
}

type step struct {
	Fragments []*fragment
	Res       *result
}

func (s *step) kind() kind {
	return stepKind
}

type concept struct {
	CptStep *step
	Items   []item
}

func (c *concept) kind() kind {
	return conceptKind
}

type comment struct {
	Text string
}

func (c *comment) kind() kind {
	return commentKind
}

type result struct {
	Status     status
	StackTrace string
	ScreenShot string
	Message    string
	ExecTime   string
}

type status int

const (
	pass status = iota
	fail
	skip
	notExecuted
)

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

func generate(suiteRes *gauge_messages.ProtoSuiteResult, w io.Writer) {
	overview := toOverview(suiteRes)
	sidebar := toSidebar(suiteRes)
	specHeader := toSpecHeader(suiteRes.GetSpecResults()[0])
	spec := toSpec(suiteRes.GetSpecResults()[0])

	gen(htmlStartTag, w, nil)
	gen(pageHeaderTag, w, nil)
	gen(bodyStartTag, w, nil)
	gen(bodyHeaderTag, w, overview)
	gen(mainStartTag, w, nil)
	gen(containerStartDiv, w, nil)
	gen(reportOverviewTag, w, overview)
	gen(specsStartDiv, w, nil)
	gen(sidebarDiv, w, sidebar)
	gen(specContainerStartDiv, w, nil)
	gen(specHeaderStartTag, w, specHeader)
	gen(tagsDiv, w, specHeader)
	gen(headerEndTag, w, nil)
	gen(specsItemsContainerDiv, w, nil)
	gen(specCommentsAndTableTag, w, spec)
	gen(scenarioContainerStartDiv, w, spec.Scenarios[0])
	gen(scenarioHeaderStartDiv, w, spec.Scenarios[0])
	gen(tagsDiv, w, spec.Scenarios[0])
	gen(endDiv, w, nil)
	generateItems(w, spec.Scenarios[0].Contexts, generateStep)
	generateItems(w, spec.Scenarios[0].Items, generateItem)
	generateItems(w, spec.Scenarios[0].TearDown, generateStep)
	gen(endDiv, w, nil)
	gen(endDiv, w, nil)
	gen(endDiv, w, nil)
	gen(endDiv, w, nil)
	gen(endDiv, w, nil)
	gen(endDiv, w, nil)
	gen(mainEndTag, w, nil)
	gen(bodyFooterTag, w, nil)
	gen(bodyEndTag, w, nil)
	gen(htmlEndTag, w, nil)
}

func generateItems(w io.Writer, items []item, predicate func(w io.Writer, item item)) {
	for _, item := range items {
		predicate(w, item)
	}
}

func generateStep(w io.Writer, item item) {
	gen(contextStepStartDiv, w, nil)
	generateItem(w, item)
	gen(endDiv, w, nil)
}

func generateItem(w io.Writer, item item) {
	switch item.kind() {
	case stepKind:
		gen(stepDiv, w, item.(*step))
	case commentKind:
		gen(commentSpan, w, item.(*comment))
	}
}
