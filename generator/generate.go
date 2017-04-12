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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"path"

	"github.com/getgauge/common"
	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/theme"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type summary struct {
	Total   int
	Failed  int
	Passed  int
	Skipped int
}

type overview struct {
	ProjectName   string
	Env           string
	Tags          string
	SuccessRate   float32
	ExecutionTime string
	Timestamp     string
	Summary       *summary
	BasePath      string
}

type specsMeta struct {
	SpecName      string
	ExecutionTime string
	Failed        bool
	Skipped       bool
	Tags          []string
	ReportFile    string
}

type sidebar struct {
	IsBeforeHookFailure bool
	Specs               []*specsMeta
}

type specHeader struct {
	SpecName      string
	ExecutionTime string
	FileName      string
	Tags          []string
	Summary       *summary
}

type errorType string
type tokenKind string
type status string

type buildError struct {
	ErrorType  errorType
	FileName   string
	LineNumber int
	Message    string
}

// SuiteResult holds the aggregated execution information for a run
type SuiteResult struct {
	ProjectName            string       `json:"projectName"`
	Timestamp              string       `json:"timestamp"`
	SuccessRate            float32      `json:"successRate"`
	Environment            string       `json:"environment"`
	Tags                   string       `json:"tags"`
	ExecutionTime          int64        `json:"executionTime"`
	ExecutionStatus        status       `json:"executionStatus"`
	SpecResults            []*spec      `json:"specResults"`
	BeforeSuiteHookFailure *hookFailure `json:"beforeSuiteHookFailure"`
	AfterSuiteHookFailure  *hookFailure `json:"afterSuiteHookFailure"`
	PassedSpecsCount       int          `json:"passedSpecsCount"`
	FailedSpecsCount       int          `json:"failedSpecsCount"`
	SkippedSpecsCount      int          `json:"skippedSpecsCount"`
	BasePath               string       `json:"basePath"`
}

type spec struct {
	CommentsBeforeDatatable []string     `json:"commentsBeforeDatatable"`
	CommentsAfterDatatable  []string     `json:"comentsAfterDatatable"`
	SpecHeading             string       `json:"specHeading"`
	FileName                string       `json:"fileName"`
	Tags                    []string     `json:"tags"`
	ExecutionTime           int64        `json:"executionTime"`
	ExecutionStatus         status       `json:"executionStatus"`
	Scenarios               []*scenario  `json:"scenarios"`
	IsTableDriven           bool         `json:"isTableDriven"`
	Datatable               *table       `json:"datatable"`
	BeforeSpecHookFailure   *hookFailure `json:"beforeSpecHookFailure"`
	AfterSpecHookFailure    *hookFailure `json:"afterSpecHookFailure"`
	PassedScenarioCount     int          `json:"passedScenarioCount"`
	FailedScenarioCount     int          `json:"failedScenarioCount"`
	SkippedScenarioCount    int          `json:"skippedScenarioCount"`
	Errors                  []error      `json:"errors"`
}

type scenario struct {
	Heading                   string       `json:"scenarioHeading"`
	Tags                      []string     `json:"tags"`
	ExecutionTime             string       `json:"executionTime"`
	ExecutionStatus           status       `json:"executionStatus"`
	Contexts                  []item       `json:"contexts"`
	Teardowns                 []item       `json:"teardowns"`
	Items                     []item       `json:"items"`
	BeforeScenarioHookFailure *hookFailure `json:"beforeScenarioHookFailure"`
	AfterScenarioHookFailure  *hookFailure `json:"afterScenarioHookFailure"`
	SkipErrors                []string     `json:"skipErrors"`
	TableRowIndex             int          `json:"tableRowIndex"`
}

type step struct {
	Fragments             []*fragment  `json:"fragments"`
	ItemType              tokenKind    `json:"itemType"`
	StepText              string       `json:"stepText"`
	Table                 *table       `json:"table"`
	BeforeStepHookFailure *hookFailure `json:"beforeStepHookFailure"`
	AfterStepHookFailure  *hookFailure `json:"afterStepHookFailure"`
	Result                *result      `json:"result"`
}

func (s *step) Kind() tokenKind {
	return stepKind
}

type result struct {
	Status        status    `json:"status"`
	StackTrace    string    `json:"stackTrace"`
	Screenshot    string    `json:"screenshot"`
	ErrorMessage  string    `json:"errorMessage"`
	ExecutionTime string    `json:"executionTime"`
	SkippedReason string    `json:"skippedReason"`
	Messages      []string  `json:"messages"`
	ErrorType     errorType `json:"errorType"`
}

type hookFailure struct {
	HookName   string `json:"hookName"`
	ErrMsg     string `json:"errorMessage"`
	Screenshot string `json:"screenshot"`
	StackTrace string `json:"stackTrace"`
}

type concept struct {
	ItemType    tokenKind `json:"itemType"`
	ConceptStep *step     `json:"conceptStep"`
	Items       []item    `json:"items"`
	Result      result    `json:"result"`
}

func (s *concept) Kind() tokenKind {
	return conceptKind
}

type table struct {
	Headers []string `json:"headers"`
	Rows    []*row   `json:"rows"`
}

type row struct {
	Cells  []string `json:"cells"`
	Result status   `json:"status"`
}

func (e buildError) Error() string {
	if e.isParseError() {
		return "[Parse Error] " + e.Message
	}
	return "[Validation Error] " + e.Message
}

func (e buildError) isParseError() bool {
	return e.ErrorType == parseErrorType
}

type item struct {
	Kind    tokenKind
	Step    *step
	Concept *concept
	Comment *comment
}

type comment struct {
	Text string
}

func (c *comment) Kind() tokenKind {
	return commentKind
}

type searchIndex struct {
	Tags  map[string][]string `json:"tags"`
	Specs map[string][]string `json:"specs"`
}

const (
	pass                  status    = "pass"
	fail                  status    = "fail"
	skip                  status    = "skip"
	notExecuted           status    = "not executed"
	stepKind              tokenKind = "step"
	conceptKind           tokenKind = "concept"
	commentKind           tokenKind = "comment"
	assertionErrorType    errorType = "assertion"
	parseErrorType        errorType = "parse"
	verificationErrorType errorType = "verification"
	validationErrorType   errorType = "validation"
)

var parsedTemplates *template.Template

func readTemplates(themePath string) {
	var encodeNewLine = func(s string) string {
		return strings.Replace(s, "\n", "<br/>", -1)
	}
	var parseMarkdown = func(args ...interface{}) string {
		s := blackfriday.MarkdownCommon([]byte(fmt.Sprintf("%s", args...)))
		return string(s)
	}
	var sanitizeHTML = func(s string) string {
		var b bytes.Buffer
		var html = bluemonday.UGCPolicy().SanitizeBytes([]byte(s))
		b.Write(html)
		return b.String()
	}

	var funcs = template.FuncMap{
		"parseMarkdown":       parseMarkdown,
		"sanitize":            sanitizeHTML,
		"escapeHTML":          template.HTMLEscapeString,
		"encodeNewLine":       encodeNewLine,
		"containsParseErrors": containsParseErrors,
		"toSpecHeader":        toSpecHeader,
		"toSidebar":           toSidebar,
		"toOverview":          toOverview,
		"toPath":              path.Join,
	}
	f, err := ioutil.ReadFile(filepath.Join(getAbsThemePath(themePath), "views", "partials.tmpl"))
	if err != nil {
		log.Fatalf(err.Error())
	}
	parsedTemplates, err = template.New("Reports").Funcs(funcs).Parse(string(f))
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func getAbsThemePath(themePath string) string {
	if path.IsAbs(themePath) {
		return themePath
	}
	return filepath.Join(projectRoot, themePath)
}

func execTemplate(tmplName string, w io.Writer, data interface{}) {
	err := parsedTemplates.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// ProjectRoot is root dir of current project
var projectRoot string

// GenerateReports generates HTML report in the given report dir location
func GenerateReports(res *SuiteResult, reportsDir, themePath string) error {
	readTemplates(themePath)
	f, err := os.Create(filepath.Join(reportsDir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	if res.BeforeSuiteHookFailure != nil {
		execTemplate("indexPageFailure", f, res)
	} else {
		var wg sync.WaitGroup
		wg.Add(1)
		res.BasePath = ""
		go generateIndexPage(res, f, &wg)
		go generateIndexPages(res, reportsDir, &wg)
		specRes := res.SpecResults
		for _, r := range specRes {
			relPath, _ := filepath.Rel(projectRoot, r.FileName)
			env.CreateDirectory(filepath.Join(reportsDir, filepath.Dir(relPath)))
			sf, err := os.Create(filepath.Join(reportsDir, toHTMLFileName(r.FileName, projectRoot)))
			if err != nil {
				return err
			}
			defer sf.Close()
			wg.Add(1)
			go generateSpecPage(res, r, sf, &wg)
		}
		wg.Wait()
	}
	err = generateSearchIndex(res, reportsDir)
	if err != nil {
		return err
	}
	return nil
}

func RegenerateReport(inputFile, reportsDir, themePath string) {
	b, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	res := &SuiteResult{}
	if err = json.Unmarshal(b, res); err != nil {
		log.Fatal(err.Error())
	}
	env.CreateDirectory(reportsDir)
	if themePath == "" {
		themePath = theme.GetDefaultThemePath()
	}
	GenerateReport(res, reportsDir, themePath)
}

func GenerateReport(res *SuiteResult, reportDir, themePath string) {
	err := GenerateReports(res, reportDir, themePath)
	if err != nil {
		log.Fatalf("Failed to generate reports: %s\n", err.Error())
	}
	err = theme.CopyReportTemplateFiles(themePath, reportDir)
	if err != nil {
		log.Fatalf("Error copying template directory :%s\n", err.Error())
	}
	fmt.Printf("Successfully generated html-report to => %s\n", reportDir)
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

func containsParseErrors(errors []error) bool {
	for _, e := range errors {
		if e.(buildError).isParseError() {
			return true
		}
	}
	return false
}

func generateSearchIndex(suiteRes *SuiteResult, reportsDir string) error {
	env.CreateDirectory(filepath.Join(reportsDir, "js"))
	f, err := os.Create(filepath.Join(reportsDir, "js", "search_index.js"))
	if err != nil {
		return err
	}
	defer f.Close()
	index := newSearchIndex()
	for _, r := range suiteRes.SpecResults {
		specFileName := toHTMLFileName(r.FileName, projectRoot)
		for _, t := range r.Tags {
			if !index.hasValueForTag(t, specFileName) {
				index.Tags[t] = append(index.Tags[t], specFileName)
			}
		}
		for _, s := range r.Scenarios {
			for _, t := range s.Tags {
				if !index.hasValueForTag(t, specFileName) {
					index.Tags[t] = append(index.Tags[t], specFileName)
				}
			}
		}
		specHeading := r.SpecHeading
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

func generateIndexPage(suiteRes *SuiteResult, w io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()
	execTemplate("indexPage", w, suiteRes)
}

func generateIndexPages(suiteRes *SuiteResult, reportsDir string, wg *sync.WaitGroup) {
	dirs := make(map[string]int)
	for _, s := range suiteRes.SpecResults {
		p, err := filepath.Rel(projectRoot, filepath.Dir(s.FileName))
		if err != nil {
			log.Fatal(err)
		}
		childDirs := filepath.SplitList(p)
		basePath := ""
		for _, d := range childDirs {
			if _, ok := dirs[d]; !ok {
				dirs[d] = 1
			}
			basePath = filepath.Join(basePath, d)
		}

	}
	delete(dirs, ".")
	for d := range dirs {
		dirPath := filepath.Join(reportsDir, d)
		os.MkdirAll(dirPath, common.NewDirectoryPermissions)
		p := filepath.Join(dirPath, "index.html")
		f, err := os.Create(p)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		wg.Add(1)
		res := toNestedSuiteResult(d, suiteRes)
		generateIndexPage(res, f, wg)
	}
}

func generateSpecPage(suiteRes *SuiteResult, specRes *spec, w io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()
	execTemplate("specPage", w, struct {
		SuiteRes *SuiteResult
		SpecRes  *spec
	}{suiteRes, specRes})
}
