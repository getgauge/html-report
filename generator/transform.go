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
	"time"

	"github.com/getgauge/html-report/gauge_messages"
)

const (
	execTimeFormat = "15:04:05"
)

func toOverview(res *gauge_messages.ProtoSuiteResult) *overview {
	passed := len(res.GetSpecResults()) - int(res.GetSpecsFailedCount()) - int(res.GetSpecsSkippedCount())
	return &overview{
		ProjectName: res.GetProjectName(),
		Env:         res.GetEnvironment(),
		Tags:        res.GetTags(),
		SuccRate:    res.GetSuccessRate(),
		ExecTime:    formatTime(res.GetExecutionTime()),
		Timestamp:   res.GetTimestamp(),
		TotalSpecs:  len(res.GetSpecResults()),
		Failed:      int(res.GetSpecsFailedCount()),
		Passed:      passed,
		Skipped:     int(res.GetSpecsSkippedCount()),
	}
}

func toSidebar(res *gauge_messages.ProtoSuiteResult) *sidebar {
	specsMetaList := make([]*specsMeta, 0)
	for _, specRes := range res.SpecResults {
		sm := &specsMeta{
			SpecName: specRes.ProtoSpec.GetSpecHeading(),
			ExecTime: formatTime(specRes.GetExecutionTime()),
			Failed:   specRes.GetFailed(),
			Skipped:  specRes.GetSkipped(),
			Tags:     specRes.ProtoSpec.GetTags(),
		}
		specsMetaList = append(specsMetaList, sm)
	}

	return &sidebar{
		IsPreHookFailure: res.PreHookFailure != nil,
		Specs:            specsMetaList,
	}
}

func toSpecHeader(res *gauge_messages.ProtoSpecResult) *specHeader {
	return &specHeader{
		SpecName: res.ProtoSpec.GetSpecHeading(),
		ExecTime: formatTime(res.GetExecutionTime()),
		FileName: res.ProtoSpec.GetFileName(),
		Tags:     res.ProtoSpec.GetTags(),
	}
}

func toSpec(res *gauge_messages.ProtoSpecResult) *spec {
	spec := &spec{
		CommentsBeforeTable: make([]string, 0),
		CommentsAfterTable:  make([]string, 0),
		Scenarios:           make([]*scenario, 0),
	}
	isTableScanned := false
	for _, item := range res.GetProtoSpec().GetItems() {
		switch item.GetItemType() {
		case gauge_messages.ProtoItem_Comment:
			if isTableScanned {
				spec.CommentsAfterTable = append(spec.CommentsAfterTable, item.GetComment().GetText())
			} else {
				spec.CommentsBeforeTable = append(spec.CommentsBeforeTable, item.GetComment().GetText())
			}
		case gauge_messages.ProtoItem_Table:
			spec.Table = toTable(item.GetTable())
			isTableScanned = true
		case gauge_messages.ProtoItem_Scenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetScenario()))
		}
	}
	return spec
}

func toScenario(scn *gauge_messages.ProtoScenario) *scenario {
	return &scenario{
		Heading:  scn.GetScenarioHeading(),
		ExecTime: formatTime(scn.GetExecutionTime()),
		Tags:     scn.GetTags(),
		Res:      getStatus(scn.GetFailed(), scn.GetSkipped()),
	}
}

func toStep(protoStep *gauge_messages.ProtoStep) *step {
	res := protoStep.GetStepExecutionResult().GetExecutionResult()
	result := &result{
		Status:     getStatus(res.GetFailed(), protoStep.GetStepExecutionResult().GetSkipped()),
		ScreenShot: string(res.GetScreenShot()),
		StackTrace: res.GetStackTrace(),
		Message:    res.GetErrorMessage(),
	}
	if protoStep.GetStepExecutionResult().GetSkipped() {
		result.Message = protoStep.GetStepExecutionResult().GetSkippedReason()
	}
	return &step{
		Fragments: toFragments(protoStep.GetFragments()),
		Res:       result,
	}
}

func toFragments(protoFragments []*gauge_messages.Fragment) []fragment {
	fragments := make([]fragment, 0)
	for _, f := range protoFragments {
		switch f.GetFragmentType() {
		case gauge_messages.Fragment_Text:
			fragments = append(fragments, &textFragment{Text: f.GetText()})
		case gauge_messages.Fragment_Parameter:
			switch f.GetParameter().GetParameterType() {
			case gauge_messages.Parameter_Static:
				fragments = append(fragments, &staticFragment{Text: f.GetParameter().GetValue()})
			case gauge_messages.Parameter_Dynamic:
				fragments = append(fragments, &dynamicFragment{Text: f.GetParameter().GetValue()})
			case gauge_messages.Parameter_Table:
				fragments = append(fragments, &tableFragment{Table: toTable(f.GetParameter().GetTable())})
			case gauge_messages.Parameter_Special_Table:
				fragments = append(fragments, &specialTableFragment{Name: f.GetParameter().GetName(), Table: toTable(f.GetParameter().GetTable())})
			case gauge_messages.Parameter_Special_String:
				fragments = append(fragments, &specialStringFragment{Name: f.GetParameter().GetName(), Text: f.GetParameter().GetValue()})
			}
		}
	}
	return fragments
}

func toTable(protoTable *gauge_messages.ProtoTable) *table {
	rows := make([]*row, len(protoTable.GetRows()))
	for i, r := range protoTable.GetRows() {
		rows[i] = &row{
			Cells: r.GetCells(),
			Res:   pass,
		}
	}
	return &table{Headers: protoTable.GetHeaders().GetCells(), Rows: rows}
}

func getStatus(failed, skipped bool) status {
	if failed {
		return fail
	} else if skipped {
		return skip
	}
	return pass
}

func formatTime(ms int64) string {
	return time.Unix(0, ms*int64(time.Millisecond)).UTC().Format(execTimeFormat)
}
