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
	"encoding/base64"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"path"

	gm "github.com/getgauge/html-report/gauge_messages"
)

const (
	execTimeFormat = "15:04:05"
	dothtml        = ".html"
)

// ToSuiteResult Converts the ProtoSuiteResult to SuiteResult type.
func ToSuiteResult(pRoot string, psr *gm.ProtoSuiteResult) *SuiteResult {
	projectRoot = pRoot
	suiteResult := SuiteResult{
		ProjectName:            psr.GetProjectName(),
		Environment:            psr.GetEnvironment(),
		Tags:                   psr.GetTags(),
		ExecutionTime:          psr.GetExecutionTime(),
		PassedSpecsCount:       len(psr.GetSpecResults()) - int(psr.GetSpecsFailedCount()) - int(psr.GetSpecsSkippedCount()),
		FailedSpecsCount:       int(psr.GetSpecsFailedCount()),
		SkippedSpecsCount:      int(psr.GetSpecsSkippedCount()),
		BeforeSuiteHookFailure: toHookFailure(psr.GetPreHookFailure(), "Before Suite"),
		AfterSuiteHookFailure:  toHookFailure(psr.GetPostHookFailure(), "After Suite"),
		SuccessRate:            psr.GetSuccessRate(),
		Timestamp:              psr.GetTimestamp(),
		ExecutionStatus:        pass,
	}
	if psr.GetFailed() {
		suiteResult.ExecutionStatus = fail
	}
	suiteResult.SpecResults = make([]*spec, 0)
	for _, protoSpecRes := range psr.GetSpecResults() {
		suiteResult.SpecResults = append(suiteResult.SpecResults, toSpec(protoSpecRes))
	}
	return &suiteResult
}

func toNestedSuiteResult(basePath string, result *SuiteResult) *SuiteResult {
	sr := &SuiteResult{
		ProjectName: result.ProjectName,
		Timestamp:   result.Timestamp,
		Environment: result.Environment,
		Tags:        result.Tags,
		BeforeSuiteHookFailure: result.BeforeSuiteHookFailure,
		AfterSuiteHookFailure:  result.AfterSuiteHookFailure,
		ExecutionStatus:        pass,
		SpecResults:            getNestedSpecResults(result.SpecResults, basePath),
		BasePath:               filepath.Clean(basePath),
	}

	for _, spec := range sr.SpecResults {
		if spec.ExecutionStatus == fail {
			sr.ExecutionStatus = fail
			sr.FailedSpecsCount++
		}
		if spec.ExecutionStatus == skip {
			sr.SkippedSpecsCount++
		}
		if spec.ExecutionStatus == pass {
			sr.PassedSpecsCount++
		}
		sr.ExecutionTime += spec.ExecutionTime
	}
	sr.SuccessRate = getSuccessRate(len(sr.SpecResults), sr.FailedSpecsCount+sr.SkippedSpecsCount)
	return sr
}

func getSuccessRate(totalSpecs int, failedSpecs int) float32 {
	if totalSpecs == 0 {
		return 0
	}
	return (float32)(100.0 * (totalSpecs - failedSpecs) / totalSpecs)
}

func getNestedSpecResults(specResults []*spec, basePath string) []*spec {
	nestedSpecResults := make([]*spec, 0)
	for _, specResult := range specResults {
		rel, _ := filepath.Rel(projectRoot, specResult.FileName)
		if strings.HasPrefix(rel, basePath) {
			nestedSpecResults = append(nestedSpecResults, specResult)
		}
	}
	return nestedSpecResults
}

func toOverview(res *SuiteResult, filePath string) *overview {
	totalSpecs := 0
	if res.SpecResults != nil {
		totalSpecs = len(res.SpecResults)
	}
	base := ""
	if filePath != "" {
		base, _ = filepath.Rel(filepath.Dir(filePath), projectRoot)
		base = path.Join(base, "/")
	} else if res.BasePath != "" {
		base, _ = filepath.Rel(filepath.Join(projectRoot, res.BasePath), projectRoot)
		base = path.Join(base, "/")
	}
	return &overview{
		ProjectName:   res.ProjectName,
		Env:           res.Environment,
		Tags:          res.Tags,
		SuccessRate:   res.SuccessRate,
		ExecutionTime: formatTime(res.ExecutionTime),
		Timestamp:     res.Timestamp,
		Summary:       &summary{Failed: res.FailedSpecsCount, Total: totalSpecs, Passed: res.PassedSpecsCount, Skipped: res.SkippedSpecsCount},
		BasePath:      base,
	}
}

func toHookFailure(failure *gm.ProtoHookFailure, hookName string) *hookFailure {
	if failure == nil {
		return nil
	}

	return &hookFailure{
		ErrMsg:        failure.GetErrorMessage(),
		HookName:      hookName,
		Screenshot:    base64.StdEncoding.EncodeToString(failure.GetScreenShot()),
		StackTrace:    failure.GetStackTrace(),
		TableRowIndex: failure.TableRowIndex,
	}
}

func toHTMLFileName(specName, basePath string) string {
	specPath, err := filepath.Rel(basePath, specName)
	if err != nil {
		specPath = filepath.Join(basePath, filepath.Base(specName))
	}
	// specPath = strings.Replace(specPath, string(filepath.Separator), "_", -1)
	ext := filepath.Ext(specPath)
	return strings.TrimSuffix(specPath, ext) + dothtml
}

func getFilePathBasedOnSpecLocation(specFilePath, path string) string {
	if specFilePath != "" {
		return filepath.Dir(specFilePath)
	}
	if path != "" {
		return filepath.Join(projectRoot, path)
	}
	return projectRoot
}

func toSidebar(res *SuiteResult, specFilePath string) *sidebar {
	basePath := getFilePathBasedOnSpecLocation(specFilePath, res.BasePath)
	specsMetaList := make([]*specsMeta, 0)
	for _, specRes := range res.SpecResults {
		sm := &specsMeta{
			SpecName:      specRes.SpecHeading,
			ExecutionTime: formatTime(specRes.ExecutionTime),
			Failed:        specRes.ExecutionStatus == fail,
			Skipped:       specRes.ExecutionStatus == skip,
			Tags:          specRes.Tags,
			ReportFile:    toHTMLFileName(specRes.FileName, basePath),
		}
		specsMetaList = append(specsMetaList, sm)
	}
	sort.Sort(byStatus(specsMetaList))

	return &sidebar{
		IsBeforeHookFailure: res.BeforeSuiteHookFailure != nil,
		Specs:               specsMetaList,
	}
}

type byStatus []*specsMeta

func (s byStatus) Len() int {
	return len(s)
}
func (s byStatus) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byStatus) Less(i, j int) bool {
	return getState(s[i]) < getState(s[j])
}

func getState(r *specsMeta) int {
	if r.Failed {
		return -1
	}
	if r.Skipped {
		return 0
	}
	return 1
}

type bySceStatus []*scenario

func (s bySceStatus) Len() int {
	return len(s)
}
func (s bySceStatus) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s bySceStatus) Less(i, j int) bool {
	return getSceState(s[i]) < getSceState(s[j])
}

func getSceState(s *scenario) int {
	if s.ExecutionStatus == fail {
		return -1
	}
	if s.ExecutionStatus == skip {
		return 0
	}
	return 1
}

func getSpecName(s *gm.ProtoSpec) string {
	specName := s.GetSpecHeading()
	if strings.TrimSpace(specName) == "" {
		specName = filepath.Base(s.GetFileName())
	}
	return specName
}

func toSpecHeader(res *spec) *specHeader {
	return &specHeader{
		SpecName:      res.SpecHeading,
		ExecutionTime: formatTime(res.ExecutionTime),
		FileName:      res.FileName,
		Tags:          res.Tags,
		Summary:       toScenarioSummary(res),
	}
}

func toSpec(res *gm.ProtoSpecResult) *spec {
	spec := &spec{
		Scenarios:             make([]*scenario, 0),
		BeforeSpecHookFailure: make([]*hookFailure, 0),
		AfterSpecHookFailure:  make([]*hookFailure, 0),
		Errors:                make([]buildError, 0),
		FileName:              res.GetProtoSpec().GetFileName(),
		SpecHeading:           res.GetProtoSpec().GetSpecHeading(),
		IsTableDriven:         res.GetProtoSpec().GetIsTableDriven(),
		ExecutionTime:         res.GetExecutionTime(),
		ExecutionStatus:       pass,
	}
	if res.GetFailed() {
		spec.ExecutionStatus = fail
	}
	if res.GetSkipped() {
		spec.ExecutionStatus = skip
	}
	sourceTags := res.GetProtoSpec().GetTags()
	if sourceTags != nil {
		spec.Tags = make([]string, 0)
		for _, t := range sourceTags {
			spec.Tags = append(spec.Tags, t)
		}
	}
	if hasParseErrors(res.Errors) {
		spec.Errors = toErrors(res.Errors)
		return spec
	}
	isTableScanned := false
	for _, item := range res.GetProtoSpec().GetItems() {
		switch item.GetItemType() {
		case gm.ProtoItem_Comment:
			if isTableScanned {
				spec.CommentsAfterDatatable = append(spec.CommentsAfterDatatable, item.GetComment().GetText())
			} else {
				spec.CommentsBeforeDatatable = append(spec.CommentsBeforeDatatable, item.GetComment().GetText())
			}
		case gm.ProtoItem_Table:
			spec.Datatable = toTable(item.GetTable())
			isTableScanned = true
		case gm.ProtoItem_Scenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetScenario(), -1))
		case gm.ProtoItem_TableDrivenScenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetTableDrivenScenario().GetScenario(), int(item.GetTableDrivenScenario().GetTableRowIndex())))
		}
	}
	for _, preHookFailure := range res.GetProtoSpec().GetPreHookFailures() {
		spec.BeforeSpecHookFailure = append(spec.BeforeSpecHookFailure, toHookFailure(preHookFailure, "Before Spec"))
	}
	for _, postHookFailure := range res.GetProtoSpec().GetPostHookFailures() {
		spec.AfterSpecHookFailure = append(spec.AfterSpecHookFailure, toHookFailure(postHookFailure, "After Spec"))
	}

	if res.GetProtoSpec().GetIsTableDriven() {
		computeTableDrivenStatuses(spec)
	}
	p, f, s := computeScenarioStatistics(spec)
	spec.PassedScenarioCount = p
	spec.FailedScenarioCount = f
	spec.SkippedScenarioCount = s
	sort.Sort(bySceStatus(spec.Scenarios))
	return spec
}

func computeScenarioStatistics(s *spec) (passed, failed, skipped int) {
	for _, scn := range s.Scenarios {
		switch scn.ExecutionStatus {
		case pass:
			passed++
		case fail:
			failed++
		case skip:
			skipped++
		}
	}
	return passed, failed, skipped
}

func toErrors(errors []*gm.Error) []buildError {
	var buildErrors []buildError
	for _, e := range errors {
		err := buildError{FileName: e.Filename, LineNumber: int(e.LineNumber), Message: e.Message}
		if e.Type == gm.Error_PARSE_ERROR {
			err.ErrorType = parseErrorType
		} else if e.Type == gm.Error_VALIDATION_ERROR {
			err.ErrorType = validationErrorType
		}
		buildErrors = append(buildErrors, err)
	}
	return buildErrors
}

func hasParseErrors(errors []*gm.Error) bool {
	for _, e := range errors {
		if e.Type == gm.Error_PARSE_ERROR {
			return true
		}
	}
	return false
}

func computeTableDrivenStatuses(spec *spec) {
	for _, r := range spec.Datatable.Rows {
		r.Result = skip
	}
	for _, s := range spec.Scenarios {
		if s.TableRowIndex >= 0 {
			var row = spec.Datatable.Rows[s.TableRowIndex]
			if s.ExecutionStatus == fail {
				row.Result = fail
			} else if row.Result != fail && s.ExecutionStatus == pass {
				row.Result = pass
			}
		}
	}
	for _, s := range spec.BeforeSpecHookFailure {
		if s.TableRowIndex >= 0 {
			var row = spec.Datatable.Rows[s.TableRowIndex]
			row.Result = fail
		}
	}
	for _, s := range spec.AfterSpecHookFailure {
		if s.TableRowIndex >= 0 {
			var row = spec.Datatable.Rows[s.TableRowIndex]
			row.Result = fail
		}
	}
}

func toScenarioSummary(s *spec) *summary {
	var sum = summary{Failed: s.FailedScenarioCount, Passed: s.PassedScenarioCount, Skipped: s.SkippedScenarioCount}
	sum.Total = sum.Failed + sum.Passed + sum.Skipped
	return &sum
}

func toScenario(scn *gm.ProtoScenario, tableRowIndex int) *scenario {
	return &scenario{
		Heading:                   scn.GetScenarioHeading(),
		ExecutionTime:             formatTime(scn.GetExecutionTime()),
		Tags:                      scn.GetTags(),
		ExecutionStatus:           getScenarioStatus(scn),
		Contexts:                  getItems(scn.GetContexts()),
		Items:                     getItems(scn.GetScenarioItems()),
		Teardowns:                 getItems(scn.GetTearDownSteps()),
		BeforeScenarioHookFailure: toHookFailure(scn.GetPreHookFailure(), "Before Scenario"),
		AfterScenarioHookFailure:  toHookFailure(scn.GetPostHookFailure(), "After Scenario"),
		TableRowIndex:             tableRowIndex,
	}
}

func toComment(protoComment *gm.ProtoComment) *comment {
	return &comment{Text: protoComment.GetText()}
}

func toStep(protoStep *gm.ProtoStep) *step {
	res := protoStep.GetStepExecutionResult().GetExecutionResult()
	result := &result{
		Status:        getStepStatus(protoStep.GetStepExecutionResult()),
		Screenshot:    base64.StdEncoding.EncodeToString(res.GetScreenShot()),
		StackTrace:    res.GetStackTrace(),
		ErrorMessage:  res.GetErrorMessage(),
		ExecutionTime: formatTime(res.GetExecutionTime()),
		Messages:      res.GetMessage(),
	}
	if protoStep.GetStepExecutionResult().GetSkipped() {
		result.SkippedReason = protoStep.GetStepExecutionResult().GetSkippedReason()
	}
	return &step{
		Fragments:             toFragments(protoStep.GetFragments()),
		Result:                result,
		BeforeStepHookFailure: toHookFailure(protoStep.GetStepExecutionResult().GetPreHookFailure(), "Before Step"),
		AfterStepHookFailure:  toHookFailure(protoStep.GetStepExecutionResult().GetPostHookFailure(), "After Step"),
	}
}

func toConcept(protoConcept *gm.ProtoConcept) *concept {
	protoConcept.ConceptStep.StepExecutionResult = protoConcept.GetConceptExecutionResult()
	return &concept{
		ConceptStep: toStep(protoConcept.GetConceptStep()),
		Items:       getItems(protoConcept.GetSteps()),
	}
}

func toFileName(name string) string {
	if strings.Contains(name, ":") {
		return strings.Split(name, ":")[1]
	}
	return name
}

func toFragments(protoFragments []*gm.Fragment) []*fragment {
	fragments := make([]*fragment, 0)
	for _, f := range protoFragments {
		switch f.GetFragmentType() {
		case gm.Fragment_Text:
			fragments = append(fragments, &fragment{FragmentKind: textFragmentKind, Text: f.GetText()})
		case gm.Fragment_Parameter:
			switch f.GetParameter().GetParameterType() {
			case gm.Parameter_Static:
				fragments = append(fragments, &fragment{FragmentKind: staticFragmentKind, Text: f.GetParameter().GetValue()})
			case gm.Parameter_Dynamic:
				fragments = append(fragments, &fragment{FragmentKind: dynamicFragmentKind, Text: f.GetParameter().GetValue()})
			case gm.Parameter_Table:
				fragments = append(fragments, &fragment{FragmentKind: tableFragmentKind, Table: toTable(f.GetParameter().GetTable())})
			case gm.Parameter_Special_Table:
				fragments = append(fragments, &fragment{FragmentKind: specialTableFragmentKind, Name: f.GetParameter().GetName(), Text: toCsv(f.GetParameter().GetTable()), FileName: toFileName(f.GetParameter().GetName())})
			case gm.Parameter_Special_String:
				fragments = append(fragments, &fragment{FragmentKind: specialStringFragmentKind, Name: f.GetParameter().GetName(), Text: f.GetParameter().GetValue(), FileName: toFileName(f.GetParameter().GetName())})
			}
		}
	}
	return fragments
}

func toTable(protoTable *gm.ProtoTable) *table {
	rows := make([]*row, len(protoTable.GetRows()))
	for i, r := range protoTable.GetRows() {
		rows[i] = &row{
			Cells:  r.GetCells(),
			Result: pass,
		}
	}
	return &table{Headers: protoTable.GetHeaders().GetCells(), Rows: rows}
}

func toCsv(protoTable *gm.ProtoTable) string {
	csv := []string{strings.Join(protoTable.GetHeaders().GetCells(), ",")}
	for _, row := range protoTable.GetRows() {
		csv = append(csv, strings.Join(row.GetCells(), ","))
	}
	return strings.Join(csv, "\n")
}

func getItems(protoItems []*gm.ProtoItem) []item {
	items := make([]item, 0)
	for _, i := range protoItems {
		switch i.GetItemType() {
		case gm.ProtoItem_Step:
			items = append(items, item{Kind: stepKind, Step: toStep(i.GetStep())})
		case gm.ProtoItem_Comment:
			items = append(items, item{Kind: commentKind, Comment: toComment(i.GetComment())})
		case gm.ProtoItem_Concept:
			items = append(items, item{Kind: conceptKind, Concept: toConcept(i.GetConcept())})
		}
	}
	return items
}

func getStepStatus(res *gm.ProtoStepExecutionResult) status {
	if res.GetSkipped() {
		return skip
	}
	if res.GetExecutionResult() == nil {
		return notExecuted
	}
	if res.GetExecutionResult().GetFailed() {
		return fail
	}
	return pass
}

func getScenarioStatus(scn *gm.ProtoScenario) status {
	switch scn.GetExecutionStatus() {
	case gm.ExecutionStatus_FAILED:
		return fail
	case gm.ExecutionStatus_PASSED:
		return pass
	case gm.ExecutionStatus_SKIPPED:
		return skip
	default:
		return notExecuted
	}
}

func formatTime(ms int64) string {
	return time.Unix(0, ms*int64(time.Millisecond)).UTC().Format(execTimeFormat)
}
