/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

// Package mdgen renders Gauge SuiteResult into Markdown reports.
//
// The domain types in this file (SuiteResult, spec, scenario, step, result,
// hookFailure, concept, table, row, item, comment, buildError) are lifted
// verbatim from the original HTML generator package. They describe the shape
// of a parsed Gauge execution result independent of any output format.
//
// HTML-only types (overview, sidebar, specsMeta, specHeader, SearchIndex)
// were intentionally not ported — they were render-time helpers for the HTML
// templates and have no analogue in Markdown.
package mdgen

// summary aggregates pass/fail/skip counts for a scope (suite, spec).
// The renderer uses it for the per-scope summary tables.
type summary struct {
	Total   int
	Failed  int
	Passed  int
	Skipped int
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

// SuiteResult holds the aggregated execution information for a run.
type SuiteResult struct {
	ProjectName             string       `json:"ProjectName"`
	Timestamp               string       `json:"Timestamp"`
	SuccessRate             float32      `json:"SuccessRate"`
	Environment             string       `json:"Environment"`
	Tags                    string       `json:"Tags"`
	ExecutionTime           int64        `json:"ExecutionTime"`
	ExecutionStatus         status       `json:"ExecutionStatus"`
	SpecResults             []*spec      `json:"SpecResults"`
	BeforeSuiteHookFailure  *hookFailure `json:"BeforeSuiteHookFailure"`
	AfterSuiteHookFailure   *hookFailure `json:"AfterSuiteHookFailure"`
	PassedSpecsCount        int          `json:"PassedSpecsCount"`
	FailedSpecsCount        int          `json:"FailedSpecsCount"`
	SkippedSpecsCount       int          `json:"SkippedSpecsCount"`
	PassedScenarioCount     int          `json:"PassedScenarioCount"`
	FailedScenarioCount     int          `json:"FailedScenarioCount"`
	SkippedScenarioCount    int          `json:"SkippedScenarioCount"`
	BasePath                string       `json:"BasePath"`
	PreHookMessages         []string     `json:"PreHookMessages"`
	PostHookMessages        []string     `json:"PostHookMessages"`
	PreHookScreenshotFiles  []string     `json:"PreHookScreenshotFiles"`
	PostHookScreenshotFiles []string     `json:"PostHookScreenshotFiles"`
	PreHookScreenshots      []string     `json:"PreHookScreenshots"`
	PostHookScreenshots     []string     `json:"PostHookScreenshots"`
}

type spec struct {
	BasePath                string         `json:"BasePath"`
	CommentsBeforeDatatable string         `json:"CommentsBeforeDatatable"`
	CommentsAfterDatatable  string         `json:"CommentsAfterDatatable"`
	SpecHeading             string         `json:"SpecHeading"`
	FileName                string         `json:"FileName"`
	SpecFileName            string         `json:"SpecFileName"`
	Tags                    []string       `json:"Tags"`
	ExecutionTime           int64          `json:"ExecutionTime"`
	ExecutionStatus         status         `json:"ExecutionStatus"`
	Scenarios               []*scenario    `json:"Scenarios"`
	IsTableDriven           bool           `json:"IsTableDriven"`
	Datatable               *table         `json:"Datatable"`
	BeforeSpecHookFailures  []*hookFailure `json:"BeforeSpecHookFailures"`
	AfterSpecHookFailures   []*hookFailure `json:"AfterSpecHookFailures"`
	PassedScenarioCount     int            `json:"PassedScenarioCount"`
	FailedScenarioCount     int            `json:"FailedScenarioCount"`
	SkippedScenarioCount    int            `json:"SkippedScenarioCount"`
	Errors                  []buildError   `json:"Errors"`
	PreHookMessages         []string       `json:"PreHookMessages"`
	PostHookMessages        []string       `json:"PostHookMessages"`
	PreHookScreenshotFiles  []string       `json:"PreHookScreenshotFiles"`
	PostHookScreenshotFiles []string       `json:"PostHookScreenshotFiles"`
	PreHookScreenshots      []string       `json:"PreHookScreenshots"`
	PostHookScreenshots     []string       `json:"PostHookScreenshots"`
}

type scenario struct {
	BasePath                  string       `json:"BasePath"`
	Heading                   string       `json:"Heading"`
	Tags                      []string     `json:"Tags"`
	ExecutionTime             string       `json:"ExecutionTime"`
	ExecutionStatus           status       `json:"ExecutionStatus"`
	Contexts                  []item       `json:"Contexts"`
	Teardowns                 []item       `json:"Teardowns"`
	Items                     []item       `json:"Items"`
	BeforeScenarioHookFailure *hookFailure `json:"BeforeScenarioHookFailure"`
	AfterScenarioHookFailure  *hookFailure `json:"AfterScenarioHookFailure"`
	SkipErrors                []string     `json:"SkipErrors"`
	TableRowIndex             int          `json:"TableRowIndex"`
	ScenarioTableRowIndex     int          `json:"ScenarioTableRowIndex"`
	IsSpecTableDriven         bool         `json:"IsSpecTableDriven"`
	IsScenarioTableDriven     bool         `json:"IsScenarioTableDriven"`
	ScenarioDataTable         *table       `json:"ScenarioDataTable"`
	ScenarioTableRow          *table       `json:"ScenarioTableRow"`
	PreHookMessages           []string     `json:"PreHookMessages"`
	PostHookMessages          []string     `json:"PostHookMessages"`
	PreHookScreenshotFiles    []string     `json:"PreHookScreenshotFiles"`
	PostHookScreenshotFiles   []string     `json:"PostHookScreenshotFiles"`
	PreHookScreenshots        []string     `json:"PreHookScreenshots"`
	PostHookScreenshots       []string     `json:"PostHookScreenshots"`
	RetriesCount              int          `json:"RetriesCount"`
}

type step struct {
	BasePath                string       `json:"BasePath"`
	Fragments               []*fragment  `json:"Fragments"`
	ItemType                tokenKind    `json:"ItemType"`
	StepText                string       `json:"StepText"`
	Table                   *table       `json:"Table"`
	BeforeStepHookFailure   *hookFailure `json:"BeforeStepHookFailure"`
	AfterStepHookFailure    *hookFailure `json:"AfterStepHookFailure"`
	Result                  *result      `json:"Result"`
	PreHookMessages         []string     `json:"PreHookMessages"`
	PostHookMessages        []string     `json:"PostHookMessages"`
	PreHookScreenshotFiles  []string     `json:"PreHookScreenshotFiles"`
	PostHookScreenshotFiles []string     `json:"PostHookScreenshotFiles"`
	PreHookScreenshots      []string     `json:"PreHookScreenshots"`
	PostHookScreenshots     []string     `json:"PostHookScreenshots"`
}

func (s *step) Kind() tokenKind {
	return stepKind
}

type result struct {
	BasePath              string    `json:"BasePath"`
	Status                status    `json:"Status"`
	StackTrace            string    `json:"StackTrace"`
	FailureScreenshotFile string    `json:"ScreenshotFile"`
	FailureScreenshot     string    `json:"Screenshot"`
	ErrorMessage          string    `json:"ErrorMessage"`
	ExecutionTime         string    `json:"ExecutionTime"`
	SkippedReason         string    `json:"SkippedReason"`
	Messages              []string  `json:"Messages"`
	ErrorType             errorType `json:"ErrorType"`
	ScreenshotFiles       []string  `json:"ScreenshotFiles"`
	Screenshots           []string  `json:"Screenshots"`
}

type hookFailure struct {
	BasePath              string `json:"BasePath"`
	HookName              string `json:"HookName"`
	ErrMsg                string `json:"ErrMsg"`
	FailureScreenshotFile string `json:"ScreenshotFile"`
	FailureScreenshot     string `json:"Screenshot"`
	StackTrace            string `json:"StackTrace"`
	TableRowIndex         int32  `json:"TableRowIndex"`
}

type concept struct {
	ItemType    tokenKind `json:"ItemType"`
	ConceptStep *step     `json:"ConceptStep"`
	Items       []item    `json:"Items"`
	Result      result    `json:"Result"`
}

func (s *concept) Kind() tokenKind {
	return conceptKind
}

type table struct {
	Headers []string `json:"Headers"`
	Rows    []*row   `json:"Rows"`
}

type row struct {
	Cells  []string `json:"Cells"`
	Result status   `json:"Status"`
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

const (
	pass                status    = "pass"
	fail                status    = "fail"
	skip                status    = "skip"
	notExecuted         status    = "not executed"
	stepKind            tokenKind = "step"
	conceptKind         tokenKind = "concept"
	commentKind         tokenKind = "comment"
	parseErrorType      errorType = "parse"
	validationErrorType errorType = "validation"
)
