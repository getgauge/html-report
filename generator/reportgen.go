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

	gm "github.com/getgauge/html-report/gauge_messages"
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
	StackTrace string
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
	Heading         string
	ExecTime        string
	Tags            []string
	Res             status
	Contexts        []item
	Items           []item
	TearDown        []item
	PreHookFailure  *hookFailure
	PostHookFailure *hookFailure
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
	Screenshot string
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

func generate(suiteRes *gm.ProtoSuiteResult, w io.Writer) {
	overview := toOverview(suiteRes)

	gen(htmlStartTag, w, nil)
	gen(pageHeaderTag, w, nil)
	gen(bodyStartTag, w, nil)
	gen(bodyHeaderTag, w, overview)
	gen(mainStartTag, w, nil)
	gen(containerStartDiv, w, nil)
	gen(reportOverviewTag, w, overview)

	if suiteRes.GetPreHookFailure() != nil {
		gen(hookFailureDiv, w, toHookFailure(suiteRes.GetPreHookFailure(), "Before Suite"))
	}

	if suiteRes.GetPostHookFailure() != nil {
		gen(hookFailureDiv, w, toHookFailure(suiteRes.GetPostHookFailure(), "After Suite"))
	}

	if suiteRes.GetPreHookFailure() == nil {
		gen(specsStartDiv, w, nil)
		gen(sidebarDiv, w, toSidebar(suiteRes))
		generateSpec(w, suiteRes.GetSpecResults()[0])
		gen(endDiv, w, nil)
	}

	gen(endDiv, w, nil)
	gen(mainEndTag, w, nil)
	gen(bodyFooterTag, w, nil)
	gen(bodyEndTag, w, nil)
	gen(htmlEndTag, w, nil)
}

func generateSpec(w io.Writer, res *gm.ProtoSpecResult) {
	specHeader := toSpecHeader(res)
	spec := toSpec(res)

	gen(specContainerStartDiv, w, nil)
	gen(specHeaderStartTag, w, specHeader)
	gen(tagsDiv, w, specHeader)
	gen(headerEndTag, w, nil)
	gen(specsItemsContainerDiv, w, nil)
	gen(specCommentsAndTableTag, w, spec)
	for _, scn := range spec.Scenarios {
		generateScenario(w, scn)
	}
	gen(endDiv, w, nil)
	gen(endDiv, w, nil)
	gen(endDiv, w, nil)
}

func generateScenario(w io.Writer, scn *scenario) {
	gen(scenarioContainerStartDiv, w, scn)
	gen(scenarioHeaderStartDiv, w, scn)
	gen(tagsDiv, w, scn)
	gen(endDiv, w, nil)
	generateItems(w, scn.Contexts, generateStep)
	generateItems(w, scn.Items, generateItem)
	generateItems(w, scn.TearDown, generateStep)
	if scn.PostHookFailure != nil {
		gen(hookFailureDiv, w, scn.PostHookFailure)
	}
	gen(endDiv, w, nil)
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
		gen(stepStartDiv, w, item.(*step))
		gen(stepEndDiv, w, item.(*step))
	case commentKind:
		gen(commentSpan, w, item.(*comment))
	case conceptKind:
		gen(stepStartDiv, w, item.(*concept).CptStep)
		gen(conceptSpan, w, nil)
		gen(stepEndDiv, w, item.(*concept).CptStep)
		gen(conceptStepsStartDiv, w, nil)
		generateItems(w, item.(*concept).Items, generateItem)
		gen(endDiv, w, nil)
	}
}
