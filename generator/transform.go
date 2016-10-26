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

	gm "github.com/getgauge/html-report/gauge_messages"
)

const (
	execTimeFormat = "15:04:05"
	dothtml        = ".html"
)

func toOverview(res *gm.ProtoSuiteResult, specRes *gm.ProtoSpecResult) *overview {
	totalSpecs := 0
	if res.GetSpecResults() != nil {
		totalSpecs = len(res.GetSpecResults())
	}
	passed := totalSpecs - int(res.GetSpecsFailedCount()) - int(res.GetSpecsSkippedCount())
	base := ""
	if specRes != nil {
		base, _ = filepath.Rel(filepath.Dir(specRes.ProtoSpec.GetFileName()), ProjectRoot)
		base = base + "/"
	}
	return &overview{
		ProjectName: res.GetProjectName(),
		Env:         res.GetEnvironment(),
		Tags:        res.GetTags(),
		SuccRate:    res.GetSuccessRate(),
		ExecTime:    formatTime(res.GetExecutionTime()),
		Timestamp:   res.GetTimestamp(),
		Summary:     &summary{Failed: int(res.GetSpecsFailedCount()), Total: totalSpecs, Passed: passed, Skipped: int(res.GetSpecsSkippedCount())},
		BasePath:    base,
	}
}

func toHookFailure(failure *gm.ProtoHookFailure, hookName string) *hookFailure {
	if failure == nil {
		return nil
	}

	return &hookFailure{
		ErrMsg:     failure.GetErrorMessage(),
		HookName:   hookName,
		Screenshot: base64.StdEncoding.EncodeToString(failure.GetScreenShot()),
		StackTrace: failure.GetStackTrace(),
	}
}

func toHTMLFileName(specName, projectRoot string) string {
	specPath, err := filepath.Rel(projectRoot, specName)
	if err != nil {
		specPath = filepath.Join(projectRoot, filepath.Base(specName))
	}
	// specPath = strings.Replace(specPath, string(filepath.Separator), "_", -1)
	ext := filepath.Ext(specPath)
	return strings.TrimSuffix(specPath, ext) + dothtml
}

func toSidebar(res *gm.ProtoSuiteResult, currSpec *gm.ProtoSpecResult) *sidebar {
	var basePath string
	if currSpec != nil {
		basePath = filepath.Dir(currSpec.ProtoSpec.GetFileName())
	} else {
		basePath = ProjectRoot
	}
	specsMetaList := make([]*specsMeta, 0)
	for _, specRes := range res.SpecResults {
		sm := &specsMeta{
			SpecName:   specRes.ProtoSpec.GetSpecHeading(),
			ExecTime:   formatTime(specRes.GetExecutionTime()),
			Failed:     specRes.GetFailed(),
			Skipped:    specRes.GetSkipped(),
			Tags:       specRes.ProtoSpec.GetTags(),
			ReportFile: toHTMLFileName(specRes.ProtoSpec.GetFileName(), basePath),
		}
		specsMetaList = append(specsMetaList, sm)
	}

	sort.Sort(byStatus(specsMetaList))

	return &sidebar{
		IsBeforeHookFailure: res.PreHookFailure != nil,
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

func toSpecHeader(res *gm.ProtoSpecResult) *specHeader {
	return &specHeader{
		SpecName: res.ProtoSpec.GetSpecHeading(),
		ExecTime: formatTime(res.GetExecutionTime()),
		FileName: res.ProtoSpec.GetFileName(),
		Tags:     res.ProtoSpec.GetTags(),
		Summary:  toScenarioSummary(res.GetProtoSpec()),
	}
}

func toSpec(res *gm.ProtoSpecResult) *spec {
	spec := &spec{
		CommentsBeforeTable: make([]string, 0),
		CommentsAfterTable:  make([]string, 0),
		Scenarios:           make([]*scenario, 0),
		BeforeHookFailure:   toHookFailure(res.GetProtoSpec().GetPreHookFailure(), "Before Spec"),
		AfterHookFailure:    toHookFailure(res.GetProtoSpec().GetPostHookFailure(), "After Spec"),
	}
	isTableScanned := false
	for _, item := range res.GetProtoSpec().GetItems() {
		switch item.GetItemType() {
		case gm.ProtoItem_Comment:
			if isTableScanned {
				spec.CommentsAfterTable = append(spec.CommentsAfterTable, item.GetComment().GetText())
			} else {
				spec.CommentsBeforeTable = append(spec.CommentsBeforeTable, item.GetComment().GetText())
			}
		case gm.ProtoItem_Table:
			spec.Table = toTable(item.GetTable())
			isTableScanned = true
		case gm.ProtoItem_Scenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetScenario(), -1))
		case gm.ProtoItem_TableDrivenScenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetTableDrivenScenario().GetScenario(), int(item.GetTableDrivenScenario().GetTableRowIndex())))
		}
	}

	if res.GetProtoSpec().GetIsTableDriven() {
		computeTableDrivenStatuses(spec)
	}
	return spec
}

func computeTableDrivenStatuses(spec *spec) {
	for _, r := range spec.Table.Rows {
		r.Res = skip
	}
	for _, s := range spec.Scenarios {
		var row = spec.Table.Rows[s.TableRowIndex]
		if s.ExecStatus == fail {
			row.Res = fail
		} else if row.Res != fail && s.ExecStatus == pass {
			row.Res = pass
		}
	}
}

func toScenarioSummary(s *gm.ProtoSpec) *summary {
	var sum summary
	for _, item := range s.GetItems() {
		if item.GetItemType() == gm.ProtoItem_Scenario {
			switch item.GetScenario().GetExecutionStatus() {
			case gm.ExecutionStatus_FAILED:
				sum.Failed++
			case gm.ExecutionStatus_PASSED:
				sum.Passed++
			case gm.ExecutionStatus_SKIPPED:
				sum.Skipped++
			}
		}
	}
	sum.Total = sum.Failed + sum.Passed + sum.Skipped
	return &sum
}

func toScenario(scn *gm.ProtoScenario, tableRowIndex int) *scenario {
	return &scenario{
		Heading:           scn.GetScenarioHeading(),
		ExecTime:          formatTime(scn.GetExecutionTime()),
		Tags:              scn.GetTags(),
		ExecStatus:        getScenarioStatus(scn),
		Contexts:          getItems(scn.GetContexts()),
		Items:             getItems(scn.GetScenarioItems()),
		Teardown:          getItems(scn.GetTearDownSteps()),
		BeforeHookFailure: toHookFailure(scn.GetPreHookFailure(), "Before Scenario"),
		AfterHookFailure:  toHookFailure(scn.GetPostHookFailure(), "After Scenario"),
		TableRowIndex:     tableRowIndex,
	}
}

func toComment(protoComment *gm.ProtoComment) *comment {
	return &comment{Text: protoComment.GetText()}
}

func toStep(protoStep *gm.ProtoStep) *step {
	res := protoStep.GetStepExecutionResult().GetExecutionResult()
	result := &result{
		Status:       getStepStatus(protoStep.GetStepExecutionResult()),
		Screenshot:   base64.StdEncoding.EncodeToString(res.GetScreenShot()),
		StackTrace:   res.GetStackTrace(),
		ErrorMessage: res.GetErrorMessage(),
		ExecTime:     formatTime(res.GetExecutionTime()),
		Messages:     res.GetMessage(),
	}
	if protoStep.GetStepExecutionResult().GetSkipped() {
		result.SkippedReason = protoStep.GetStepExecutionResult().GetSkippedReason()
	}
	return &step{
		Fragments:       toFragments(protoStep.GetFragments()),
		Res:             result,
		PreHookFailure:  toHookFailure(protoStep.GetStepExecutionResult().GetPreHookFailure(), "Before Step"),
		PostHookFailure: toHookFailure(protoStep.GetStepExecutionResult().GetPostHookFailure(), "After Step"),
	}
}

func toConcept(protoConcept *gm.ProtoConcept) *concept {
	protoConcept.ConceptStep.StepExecutionResult = protoConcept.GetConceptExecutionResult()
	return &concept{
		CptStep: toStep(protoConcept.GetConceptStep()),
		Items:   getItems(protoConcept.GetSteps()),
	}
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
				fragments = append(fragments, &fragment{FragmentKind: specialTableFragmentKind, Name: f.GetParameter().GetName(), Table: toTable(f.GetParameter().GetTable())})
			case gm.Parameter_Special_String:
				fragments = append(fragments, &fragment{FragmentKind: specialStringFragmentKind, Name: f.GetParameter().GetName(), Text: f.GetParameter().GetValue()})
			}
		}
	}
	return fragments
}

func toTable(protoTable *gm.ProtoTable) *table {
	rows := make([]*row, len(protoTable.GetRows()))
	for i, r := range protoTable.GetRows() {
		rows[i] = &row{
			Cells: r.GetCells(),
			Res:   pass,
		}
	}
	return &table{Headers: protoTable.GetHeaders().GetCells(), Rows: rows}
}

func getItems(protoItems []*gm.ProtoItem) []item {
	items := make([]item, 0)
	for _, i := range protoItems {
		switch i.GetItemType() {
		case gm.ProtoItem_Step:
			items = append(items, toStep(i.GetStep()))
		case gm.ProtoItem_Comment:
			items = append(items, toComment(i.GetComment()))
		case gm.ProtoItem_Concept:
			items = append(items, toConcept(i.GetConcept()))
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
