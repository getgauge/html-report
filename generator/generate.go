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
	SpecName   string
	ExecTime   string
	Failed     bool
	Skipped    bool
	Tags       []string
	ReportFile string
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
	PreHookFailure      *hookFailure
	PostHookFailure     *hookFailure
}

type scenario struct {
	Heading         string
	ExecTime        string
	Tags            []string
	ExecStatus      status
	Contexts        []item
	Items           []item
	Teardown        []item
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
	Fragments       []*fragment
	Res             *result
	PreHookFailure  *hookFailure
	PostHookFailure *hookFailure
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
)

func execTemplate(tmplName string, w io.Writer, data interface{}) {
	tmpl, err := template.New("Reports").Parse(tmplName)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func generateIndexPage(suiteRes *gm.ProtoSuiteResult, w io.Writer) {
	generateOverview(suiteRes, w)
	execTemplate(specsStartDiv, w, nil)
	execTemplate(sidebarDiv, w, toSidebar(suiteRes))
	execTemplate(congratsDiv, w, nil)
	execTemplate(endDiv, w, nil)
	generatePageFooter(w)
}

func generateSpecPage(suiteRes *gm.ProtoSuiteResult, w io.Writer) {
	generateOverview(suiteRes, w)

	if suiteRes.GetPreHookFailure() != nil {
		execTemplate(hookFailureDiv, w, toHookFailure(suiteRes.GetPreHookFailure(), "Before Suite"))
	}

	if suiteRes.GetPostHookFailure() != nil {
		execTemplate(hookFailureDiv, w, toHookFailure(suiteRes.GetPostHookFailure(), "After Suite"))
	}

	if suiteRes.GetPreHookFailure() == nil {
		execTemplate(specsStartDiv, w, nil)
		execTemplate(sidebarDiv, w, toSidebar(suiteRes))
		generateSpecDiv(w, suiteRes.GetSpecResults()[0])
		execTemplate(endDiv, w, nil)
	}

	generatePageFooter(w)
}

func generateOverview(suiteRes *gm.ProtoSuiteResult, w io.Writer) {
	overview := toOverview(suiteRes)

	execTemplate(htmlStartTag, w, nil)
	execTemplate(pageHeaderTag, w, nil)
	execTemplate(bodyStartTag, w, nil)
	execTemplate(bodyHeaderTag, w, overview)
	execTemplate(mainStartTag, w, nil)
	execTemplate(containerStartDiv, w, nil)
	execTemplate(reportOverviewTag, w, overview)
}

func generatePageFooter(w io.Writer) {
	execTemplate(endDiv, w, nil)
	execTemplate(mainEndTag, w, nil)
	execTemplate(bodyFooterTag, w, nil)
	execTemplate(bodyEndTag, w, nil)
	execTemplate(htmlEndTag, w, nil)
}

func generateSpecDiv(w io.Writer, res *gm.ProtoSpecResult) {
	specHeader := toSpecHeader(res)
	spec := toSpec(res)

	execTemplate(specContainerStartDiv, w, nil)
	execTemplate(specHeaderStartTag, w, specHeader)
	execTemplate(tagsDiv, w, specHeader)
	execTemplate(headerEndTag, w, nil)
	execTemplate(specsItemsContainerDiv, w, nil)

	if spec.PreHookFailure != nil {
		execTemplate(hookFailureDiv, w, spec.PreHookFailure)
	}
	if spec.PostHookFailure != nil {
		execTemplate(hookFailureDiv, w, spec.PostHookFailure)
	}

	execTemplate(specsItemsContentsDiv, w, nil)
	execTemplate(specCommentsAndTableTag, w, spec)

	if spec.PreHookFailure == nil {
		for _, scn := range spec.Scenarios {
			generateScenario(w, scn)
		}
	}

	execTemplate(endDiv, w, nil)
	execTemplate(endDiv, w, nil)
	execTemplate(endDiv, w, nil)
}

func generateScenario(w io.Writer, scn *scenario) {
	execTemplate(scenarioContainerStartDiv, w, scn)
	execTemplate(scenarioHeaderStartDiv, w, scn)
	execTemplate(tagsDiv, w, scn)
	execTemplate(endDiv, w, nil)
	if scn.PreHookFailure != nil {
		execTemplate(hookFailureDiv, w, scn.PreHookFailure)
	}

	generateItems(w, scn.Contexts, generateContextOrTeardown)
	generateItems(w, scn.Items, generateItem)
	generateItems(w, scn.Teardown, generateContextOrTeardown)

	if scn.PostHookFailure != nil {
		execTemplate(hookFailureDiv, w, scn.PostHookFailure)
	}
	execTemplate(endDiv, w, nil)
}

func generateItems(w io.Writer, items []item, predicate func(w io.Writer, item item)) {
	for _, item := range items {
		predicate(w, item)
	}
}

func generateContextOrTeardown(w io.Writer, item item) {
	execTemplate(contextOrTeardownStartDiv, w, nil)
	generateItem(w, item)
	execTemplate(endDiv, w, nil)
}

func generateItem(w io.Writer, item item) {
	switch item.kind() {
	case stepKind:
		execTemplate(stepStartDiv, w, item.(*step))
		execTemplate(stepBodyDiv, w, item.(*step))

		if item.(*step).PreHookFailure != nil {
			execTemplate(hookFailureDiv, w, item.(*step).PreHookFailure)
		}

		if item.(*step).Res.Status == fail && item.(*step).Res.Message != "" && item.(*step).Res.StackTrace != "" {
			execTemplate(stepFailureDiv, w, item.(*step).Res)
		}

		if item.(*step).PostHookFailure != nil {
			execTemplate(hookFailureDiv, w, item.(*step).PostHookFailure)
		}
		execTemplate(stepEndDiv, w, item.(*step))
	case commentKind:
		execTemplate(commentSpan, w, item.(*comment))
	case conceptKind:
		execTemplate(stepStartDiv, w, item.(*concept).CptStep)
		execTemplate(conceptSpan, w, nil)
		execTemplate(stepBodyDiv, w, item.(*concept).CptStep)
		execTemplate(stepEndDiv, w, item.(*concept).CptStep)
		execTemplate(conceptStepsStartDiv, w, nil)
		generateItems(w, item.(*concept).Items, generateItem)
		execTemplate(endDiv, w, nil)
	}
}
