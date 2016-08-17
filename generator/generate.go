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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/getgauge/common"
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
	BasePath    string
	EnvVars     []string
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
	IsBeforeHookFailure bool
	Specs               []*specsMeta
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
	BeforeHookFailure   *hookFailure
	AfterHookFailure    *hookFailure
}

type scenario struct {
	Heading           string
	ExecTime          string
	Tags              []string
	ExecStatus        status
	Contexts          []item
	Items             []item
	Teardown          []item
	BeforeHookFailure *hookFailure
	AfterHookFailure  *hookFailure
	TableRowIndex     int
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
	Status        status
	StackTrace    string
	Screenshot    string
	ErrorMessage  string
	ExecTime      string
	SkippedReason string
	Messages      []string
}

type searchIndex struct {
	Tags  map[string][]string `json:"tags"`
	Specs map[string][]string `json:"specs"`
}

type status int

const (
	pass status = iota
	fail
	skip
	notExecuted
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

// ProjectRoot is root dir of current project
var ProjectRoot string

// GenerateReports generates HTML report in the given report dir location
func GenerateReports(suiteRes *gm.ProtoSuiteResult, reportDir string) error {
	f, err := os.Create(filepath.Join(reportDir, "index.html"))
	if err != nil {
		return err
	}
	if suiteRes.GetPreHookFailure() != nil {
		generateOverview(suiteRes, nil, f)
		execTemplate(hookFailureDiv, f, toHookFailure(suiteRes.GetPreHookFailure(), "Before Suite"))
		if suiteRes.GetPostHookFailure() != nil {
			execTemplate(hookFailureDiv, f, toHookFailure(suiteRes.GetPostHookFailure(), "After Suite"))
		}
		generatePageFooter(f)
	} else {
		generateIndexPage(suiteRes, f)
		specRes := suiteRes.GetSpecResults()
		for _, res := range specRes {
			relPath, _ := filepath.Rel(ProjectRoot, res.GetProtoSpec().GetFileName())
			CreateDirectory(filepath.Join(reportDir, filepath.Dir(relPath)))
			sf, err := os.Create(filepath.Join(reportDir, toHTMLFileName(res.GetProtoSpec().GetFileName(), ProjectRoot)))
			if err != nil {
				return err
			}
			generateSpecPage(suiteRes, res, sf)
		}
	}
	err = generateSearchIndex(suiteRes, reportDir)
	if err != nil {
		return err
	}
	return nil
}

func newSearchIndex() *searchIndex {
	var i searchIndex
	i.Tags = make(map[string][]string)
	i.Specs = make(map[string][]string)
	return &i
}

func (i *searchIndex) hasValueForTag(tag string, spec string) bool {
	for _, s := range i.Tags[tag] {
		if s == spec {
			return true
		}
	}
	return false
}

func (i *searchIndex) hasSpec(specHeading string, specFileName string) bool {
	for _, s := range i.Specs[specHeading] {
		if s == specFileName {
			return true
		}
	}
	return false
}

func generateSearchIndex(suiteRes *gm.ProtoSuiteResult, reportDir string) error {
	CreateDirectory(filepath.Join(reportDir, "js"))
	f, err := os.Create(filepath.Join(reportDir, "js", "search_index.js"))
	if err != nil {
		return err
	}
	index := newSearchIndex()
	for _, r := range suiteRes.GetSpecResults() {
		spec := r.GetProtoSpec()
		specFileName := toHTMLFileName(spec.GetFileName(), ProjectRoot)
		for _, t := range spec.GetTags() {
			if !index.hasValueForTag(t, specFileName) {
				index.Tags[t] = append(index.Tags[t], specFileName)
			}
		}
		var addTagsFromScenario = func(s *gm.ProtoScenario) {
			for _, t := range s.GetTags() {
				if !index.hasValueForTag(t, specFileName) {
					index.Tags[t] = append(index.Tags[t], specFileName)
				}
			}
		}
		for _, i := range spec.GetItems() {
			if s := i.GetScenario(); s != nil {
				addTagsFromScenario(s)
			}
			if tds := i.GetTableDrivenScenario(); tds != nil {
				for _, s := range tds.GetScenarios() {
					addTagsFromScenario(s)
				}
			}
		}
		specHeading := spec.GetSpecHeading()
		if !index.hasSpec(specHeading, specFileName) {
			index.Specs[specHeading] = append(index.Specs[specHeading], specFileName)
		}
	}
	s, err := json.Marshal(index)
	if err != nil {
		return err
	}
	f.WriteString(fmt.Sprintf("var index = %s;", s))
	return nil
}

func generateIndexPage(suiteRes *gm.ProtoSuiteResult, w io.Writer) {
	generateOverview(suiteRes, nil, w)
	execTemplate(specsStartDiv, w, nil)
	execTemplate(sidebarDiv, w, toSidebar(suiteRes, nil))
	if !suiteRes.GetFailed() {
		execTemplate(congratsDiv, w, nil)
	}
	execTemplate(endDiv, w, nil)
	generatePageFooter(w)
}

func generateSpecPage(suiteRes *gm.ProtoSuiteResult, specRes *gm.ProtoSpecResult, w io.Writer) {
	generateOverview(suiteRes, specRes, w)

	if suiteRes.GetPreHookFailure() != nil {
		execTemplate(hookFailureDiv, w, toHookFailure(suiteRes.GetPreHookFailure(), "Before Suite"))
	}

	if suiteRes.GetPostHookFailure() != nil {
		execTemplate(hookFailureDiv, w, toHookFailure(suiteRes.GetPostHookFailure(), "After Suite"))
	}

	if suiteRes.GetPreHookFailure() == nil {
		execTemplate(specsStartDiv, w, nil)
		execTemplate(sidebarDiv, w, toSidebar(suiteRes, specRes))
		generateSpecDiv(w, specRes)
		execTemplate(endDiv, w, nil)
	}

	generatePageFooter(w)
}

func generateOverview(suiteRes *gm.ProtoSuiteResult, specRes *gm.ProtoSpecResult, w io.Writer) {
	overview := toOverview(suiteRes, specRes)

	execTemplate(htmlStartTag, w, nil)
	execTemplate(pageHeaderTag, w, overview)
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

	if spec.BeforeHookFailure != nil {
		execTemplate(hookFailureDiv, w, spec.BeforeHookFailure)
	}
	if spec.AfterHookFailure != nil {
		execTemplate(hookFailureDiv, w, spec.AfterHookFailure)
	}

	execTemplate(specsItemsContentsDiv, w, nil)
	execTemplate(specCommentsAndTableTag, w, spec)

	if spec.BeforeHookFailure == nil {
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
	if scn.BeforeHookFailure != nil {
		execTemplate(hookFailureDiv, w, scn.BeforeHookFailure)
	}

	generateItems(w, scn.Contexts, generateContextOrTeardown)
	generateItems(w, scn.Items, generateItem)
	generateItems(w, scn.Teardown, generateContextOrTeardown)

	if scn.AfterHookFailure != nil {
		execTemplate(hookFailureDiv, w, scn.AfterHookFailure)
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

		stepRes := item.(*step).Res
		if stepRes.Status == fail && stepRes.ErrorMessage != "" && stepRes.StackTrace != "" {
			execTemplate(stepFailureDiv, w, stepRes)
		}

		if item.(*step).PostHookFailure != nil {
			execTemplate(hookFailureDiv, w, item.(*step).PostHookFailure)
		}
		execTemplate(stepEndDiv, w, item.(*step))
		if stepRes.Status == skip && stepRes.SkippedReason != "" {
			execTemplate(skippedReasonDiv, w, stepRes)
		}
		execTemplate(messageDiv, w, stepRes)
	case commentKind:
		execTemplate(commentSpan, w, item.(*comment))
	case conceptKind:
		execTemplate(conceptStartDiv, w, item.(*concept).CptStep)
		execTemplate(conceptSpan, w, nil)
		execTemplate(stepBodyDiv, w, item.(*concept).CptStep)
		execTemplate(stepEndDiv, w, item.(*concept).CptStep)
		execTemplate(conceptStepsStartDiv, w, nil)
		generateItems(w, item.(*concept).Items, generateItem)
		execTemplate(endDiv, w, nil)
	}
}

// CreateDirectory creates given directory if it doesn't exist
func CreateDirectory(dir string) {
	if common.DirExists(dir) {
		return
	}
	if err := os.MkdirAll(dir, common.NewDirectoryPermissions); err != nil {
		fmt.Printf("Failed to create directory %s: %s\n", dir, err)
		os.Exit(1)
	}
}
