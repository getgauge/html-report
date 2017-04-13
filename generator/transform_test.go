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
	"reflect"
	"testing"

	gm "github.com/getgauge/html-report/gauge_messages"
	"github.com/kylelemons/godebug/pretty"
)

type transformTest struct {
	name   string
	input  *gm.ProtoSuiteResult
	output interface{}
}

func checkEqual(t *testing.T, test string, want, got interface{}) {
	if diff := pretty.Compare(got, want); diff != "" {
		t.Errorf("Test:%s\n diff: (-got +want)\n%s", test, diff)
	}
}

func newCommentItem(str string) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Comment,
		Comment: &gm.ProtoComment{
			Text: str,
		},
	}
}

func newScenarioItem(scn *gm.ProtoScenario) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Scenario,
		Scenario: scn,
	}
}

func newTableItem(headers []string, rows [][]string) *gm.ProtoItem {
	r := make([]*gm.ProtoTableRow, len(rows))
	for i, row := range rows {
		r[i] = &gm.ProtoTableRow{Cells: row}
	}
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Table,
		Table:    &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: headers}, Rows: r},
	}
}

func newStepItem(failed, skipped bool, frags []*gm.Fragment) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Step,
		Step: &gm.ProtoStep{
			StepExecutionResult: &gm.ProtoStepExecutionResult{
				ExecutionResult: &gm.ProtoExecutionResult{
					Failed:        failed,
					ExecutionTime: 211316,
				},
				Skipped: skipped,
			},
			Fragments: frags,
		},
	}
}

func newDynamicParam(val string) *gm.Parameter {
	return &gm.Parameter{
		ParameterType: gm.Parameter_Dynamic,
		Value:         val,
	}
}

func newStaticParam(val string) *gm.Parameter {
	return &gm.Parameter{
		ParameterType: gm.Parameter_Static,
		Value:         val,
	}
}

func newTableParam(headers []string, rows [][]string) *gm.Parameter {
	return &gm.Parameter{
		ParameterType: gm.Parameter_Table,
		Table:         newTableItem(headers, rows).GetTable(),
	}
}

func newTextFragment(val string) *gm.Fragment {
	return &gm.Fragment{
		FragmentType: gm.Fragment_Text,
		Text:         val,
	}
}

func newParamFragment(p *gm.Parameter) *gm.Fragment {
	return &gm.Fragment{
		FragmentType: gm.Fragment_Parameter,
		Parameter:    p,
	}
}

func newConceptItem(heading string, steps []*gm.ProtoItem, cptRes *gm.ProtoStepExecutionResult) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Concept,
		Concept: &gm.ProtoConcept{
			ConceptStep: newStepItem(false, false, []*gm.Fragment{newTextFragment(heading)}).GetStep(),
			Steps:       steps,
			ConceptExecutionResult: cptRes,
		},
	}
}

func newScreenshot() string {
	return `iVBORw0KGgoAAAANSUhEUgAAAGwAAABsCAYAAACPZlfNAAAFG0lEQVR4nOyd/3HbNhTHYV0HaCeofF3AnaDyBLUmsDRBzhMknsDOBJYmiDOBmQ3s/+MzvYE3UPClAAWkRIkEH0D8eJ87VKpigoA+BAiCxNMfooV//p4u5Mv/Ms1k+tP4p1KmQqb163tZtG0fIrJOM/lyLdOVqNcJPMu0lmkl6/XhuWid2Gw24qz5oawUKnMn07RDHoVMS1nBkrRkxMg6XYhtnWYd/hyyvso6fXFZJhv2hMmKPciXRc98UEFIe6QqGCXqAES9mi3qFKjPMqTWVhNmKcsElVsNLRQlqgt8GpAFusnLUKTthKmj8BtBnsFIk3VCi3oT/VtWk2CkQdhEVeyBKM8HNVgJAZyzhsoCOP89qe9pdCbi8IhpCKNLU1/uFWGWwUiDsGsH+Y4tjfogBEFIg7CZo7zHlDZ1lO/o0iaO8x9L2n8O8x5VmmthYOzu0QWjSfMhDPiW5mMIPoo0CPN1feFT2oun/XiXBmGFr50Jf9J8TpN5lQZhax87MnAu7fW9xOxE6XIfDbxJm6hJ22fXO2rgo6XdOs6/iRdpetCxdLmTFpxKU3Oahav8W3AurRKmupDkpEnmwn/v4VTablivjsikpKkZ9kuRkLTadRhLI8WJtL0LZ5ZGCrm0gzMdLI0UUmmtU1MsjRQyaUfnElkaKSTSTk7+sjRSBkvrNFvP0kgZJK3z7RWWRoq1tF73w1gaKVbSet/AZGmk9JZmdceZpZHSS5r1IwIsjZTO0gY908HSSOkkbfBDOCyNlJPSSJ6aYmmkHJVG9pgbSyOlVRrpc4ksjZSD0sgfJGVppEDanfnB3hpnKtSXR7XurA/OFhWqox0rOi9c5H8ELCgsqgV9rvbALY2UT/qNsxam4ZZGxl8/y7cP54shuKWRUR0cXlavsDQS/AkDLG0w1fDemzDA0gZRLQvzKgywNGuqvJ2PEtvg0WM/ZJnPDgYHaxRgIRxGdGNpnUGEuWWrMIuIbjdqBUxvWFonzhEx7+BMh9wZRCHu1LRjZjOxnaRc2JSEz2knuTfDG1KE3zOZ24bh45Z2kEKW7VL/T62FqW5wMaiE26N2arNh4i1tZbH5ypSlMbvEO8tymeCI+my7carSMGAQ23qVHTbB38zVNnvoeIkLQdsdnQ8JK5ti96hRPRlCKzW7SZzvfhw7pZgBLjHIuCIsF0aN90MySFmaLeY5bEac9+DgXCl2jxRoYdSLp0nyY2n7eJ9L7AtLqxO8MMDSfhOFMMDStkQjDLC0yISB3KVFJwzkLC1KYSBXadEKAzlKi1oYyE1a9MJATtKSEAZykZaMMJCDtKSEgdSlJScMpCwtSWEgVWnJCgMpSktaGEhNWvLCQErSshAGUpGWjTCQgrSshIHYpWUnDMQsLUthIFZp2QoDMUrLWhiITVr2wkBM0liYwpD24XnXvaSxMAMlDYvogpXGwhqoxfXBStPrwzbOi8N0pXV9mtN4iYw1R1saCwuTVmksLFwOSmNhYQNptcXrLCx8auE4WFj4zMxWxsLi4Fq/YWFxwC0sMlhYZOzinrCwOCj1m0nzAyZISv1GCytGKQbTle/6jRa2HqkgTDd2IfkqYSoydjFSYZjj1GL+moOOG+H/ph1znFKmW/ODnTB1p/XGc4GYdtB45ipu8I7asH7EZxqYOqXY/irfXhj1veswdT47F3aRoJlhoKGgC/y37YcbTv2UB66wdVDhKXXpmB2Q8yLTY7MLNKmCNOM/TDzw1FRk/AoAAP//H/csAQ85/aEAAAAASUVORK5CYII=`
}

func newStackTrace() string {
	return `StepImplementation.foo(StepImplementation.java:16)
sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
java.lang.reflect.Method.invoke(Method.java:483)
com.thoughtworks.gauge.execution.MethodExecutor.execute(MethodExecutor.java:32)
com.thoughtworks.gauge.execution.HooksExecutor$TaggedHookExecutor.executeHook(HooksExecutor.java:98)
com.thoughtworks.gauge.execution.HooksExecutor$TaggedHookExecutor.execute(HooksExecutor.java:84)
com.thoughtworks.gauge.execution.HooksExecutor.execute(HooksExecutor.java:41)
com.thoughtworks.gauge.processor.MethodExecutionMessageProcessor.executeHooks(MethodExecutionMessageProcessor.java:55)
com.thoughtworks.gauge.processor.SuiteExecutionStartingProcessor.process(SuiteExecutionStartingProcessor.java:26)
com.thoughtworks.gauge.connection.MessageDispatcher.dispatchMessages(MessageDispatcher.java:72)
com.thoughtworks.gauge.GaugeRuntime.main(GaugeRuntime.java:37)`
}

var specRes1 = &gm.ProtoSpecResult{
	Failed:        false,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading:   "specRes1",
		Tags:          []string{"tag1", "tag2"},
		FileName:      "/tmp/gauge/specs/foobar.spec",
		IsTableDriven: true,
		Items: []*gm.ProtoItem{
			newCommentItem("\n"),
			newCommentItem("This is an executable specification file. This file follows markdown syntax."),
			newCommentItem("\n"),
			newCommentItem("To execute this specification, run"),
			newCommentItem("\tgauge specs"),
			newCommentItem("\n"),
			newTableItem([]string{"Word", "Count"}, [][]string{
				[]string{"Gauge", "3"},
				[]string{"Mingle", "2"},
			}),
			newCommentItem("Comment 1"),
			newCommentItem("Comment 2"),
			newCommentItem("Comment 3"),
		},
	},
}

var specRes2 = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		FileName:    "specRes2.spec",
		SpecHeading: "specRes2",
		Tags:        []string{"tag1", "tag2", "tag3"},
	},
}

var specRes3 = &gm.ProtoSpecResult{
	Failed:        false,
	Skipped:       true,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		FileName:    "specRes3.spec",
		SpecHeading: "specRes3",
		Tags:        []string{"tag1"},
	},
}

var spec1 = &spec{
	SpecHeading:   "specRes1",
	Tags:          []string{"tag1", "tag2"},
	FileName:      "/tmp/gauge/specs/foobar.spec",
	ExecutionTime: 211316,
	IsTableDriven: true,
	Datatable: &table{
		Headers: []string{"Word", "Count"},
		Rows: []*row{
			&row{Cells: []string{"Gauge", "3"}},
			&row{Cells: []string{"Mingle", "2"}}}},
	CommentsBeforeDatatable: []string{
		"\n",
		"This is an executable specification file. This file follows markdown syntax.",
		"\n",
		"To execute this specification, run",
		"\tgauge specs",
		"\n",
	},
	CommentsAfterDatatable: []string{
		"Comment 1",
		"Comment 2",
		"Comment 3",
	},
}

var spec2 = &spec{
	ExecutionStatus: fail,
	ExecutionTime:   211316,
	FileName:        "specRes2.spec",
	SpecHeading:     "specRes2",
	Tags:            []string{"tag1", "tag2", "tag3"},
}

var spec3 = &spec{
	ExecutionStatus: skip,
	ExecutionTime:   211316,
	FileName:        "specRes3.spec",
	SpecHeading:     "specRes3",
	Tags:            []string{"tag1"},
}

var datatableDrivenSpec = &gm.ProtoSpecResult{
	Failed:        false,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading:   "specRes1",
		FileName:      "/tmp/gauge/specs/foobar.spec",
		IsTableDriven: true,
		Items: []*gm.ProtoItem{
			newTableItem(
				[]string{"Word", "Count"}, [][]string{
					[]string{"Gauge", "3"},
					[]string{"Mingle", "2"},
				}),
			&gm.ProtoItem{
				ItemType: gm.ProtoItem_TableDrivenScenario,
				TableDrivenScenario: &gm.ProtoTableDrivenScenario{
					Scenario: &gm.ProtoScenario{
						ScenarioHeading: "Scenario 1",
						ExecutionStatus: gm.ExecutionStatus_FAILED,
						ScenarioItems:   []*gm.ProtoItem{newStepItem(true, false, []*gm.Fragment{newTextFragment("Step1")})},
					},
					TableRowIndex: int32(0),
				},
			},
			&gm.ProtoItem{
				ItemType: gm.ProtoItem_TableDrivenScenario,
				TableDrivenScenario: &gm.ProtoTableDrivenScenario{
					Scenario: &gm.ProtoScenario{
						ScenarioHeading: "Scenario 1",
						ExecutionStatus: gm.ExecutionStatus_PASSED,
						ScenarioItems:   []*gm.ProtoItem{newStepItem(false, false, []*gm.Fragment{newTextFragment("Step1")})},
					},
					TableRowIndex: int32(1),
				},
			},
		},
	},
}

var specResWithSpecHookFailure = &gm.ProtoSpecResult{
	Failed:        false,
	Skipped:       true,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "specRes3",
		Tags:        []string{"tag1"},
		PreHookFailures: []*gm.ProtoHookFailure{{
					ErrorMessage: "err",
					StackTrace:   "Stacktrace",
					ScreenShot:   []byte("Screenshot"),
				}},
		PostHookFailures: []*gm.ProtoHookFailure{{
			ErrorMessage: "err",
			StackTrace:   "Stacktrace",
			ScreenShot:   []byte("Screenshot"),
		}},
	},
}

var suiteRes1 = &SuiteResult{
	ProjectName:       "projName",
	Environment:       "ci-java",
	Tags:              "!unimplemented",
	SuccessRate:       80,
	ExecutionTime:     113163,
	Timestamp:         "Jun 3, 2016 at 12:29pm",
	SpecResults:       make([]*spec, 15),
	FailedSpecsCount:  2,
	SkippedSpecsCount: 5,
	PassedSpecsCount:  8,
}

var scn = &gm.ProtoScenario{
	ScenarioHeading: "Vowel counts in single word",
	ExecutionStatus: gm.ExecutionStatus_PASSED,
	Tags:            []string{"foo", "bar"},
	ExecutionTime:   113163,
	Contexts: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Context Step1")}),
		newStepItem(true, false, []*gm.Fragment{newTextFragment("Context Step2")}),
	},
	ScenarioItems: []*gm.ProtoItem{
		newCommentItem("Comment0"),
		newStepItem(true, false, []*gm.Fragment{newTextFragment("Step1")}),
		newCommentItem("Comment1"),
		newCommentItem("Comment2"),
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Step2")}),
		newCommentItem("Comment3"),
	},
	TearDownSteps: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Teardown Step1")}),
		newStepItem(true, false, []*gm.Fragment{newTextFragment("Teardown Step2")}),
	},
}

var scnWithHookFailure = &gm.ProtoScenario{
	ScenarioHeading: "Vowel counts in single word",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	ExecutionTime:   113163,
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(true, false, []*gm.Fragment{newTextFragment("Step1")}),
	},
	PreHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: "err",
		StackTrace:   "Stacktrace",
		ScreenShot:   []byte("Screenshot"),
	},
	PostHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: "err",
		StackTrace:   "Stacktrace",
		ScreenShot:   []byte("Screenshot"),
	},
}

var skippedProtoSce = &gm.ProtoScenario{
	ScenarioHeading: "Vowel counts in single word",
	ExecutionStatus: gm.ExecutionStatus_SKIPPED,
	ExecutionTime:   0,
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(true, false, []*gm.Fragment{newTextFragment("Step1")}),
	},
}

var suiteRes2 = &SuiteResult{
	SpecResults: []*spec{spec1, spec2, spec3},
}

var protoStep = &gm.ProtoStep{
	Fragments: []*gm.Fragment{
		newTextFragment("Say "),
		newParamFragment(newStaticParam("hi")),
		newTextFragment(" to "),
		newParamFragment(newDynamicParam("gauge")),
		newParamFragment(newTableParam([]string{"Word", "Count"}, [][]string{
			[]string{"Gauge", "3"},
			[]string{"Mingle", "2"},
		})),
	},
	StepExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{
			ExecutionTime: 211316,
		},
		SkippedReason: "Step impl not found",
		Skipped:       true,
	},
}

var protoConcept = &gm.ProtoConcept{
	ConceptStep: newStepItem(false, false, []*gm.Fragment{
		newTextFragment("Say "),
		newParamFragment(newDynamicParam("hello")),
		newTextFragment(" to "),
		newParamFragment(newTableParam([]string{"Word", "Count"}, [][]string{
			[]string{"Gauge", "3"},
			[]string{"Mingle", "2"},
		})),
	}).GetStep(),
	Steps: []*gm.ProtoItem{
		{
			ItemType: gm.ProtoItem_Concept,
			Concept: &gm.ProtoConcept{
				ConceptStep: newStepItem(false, false, []*gm.Fragment{
					newTextFragment("Tell "),
					newParamFragment(newDynamicParam("hello")),
				}).GetStep(),
				Steps: []*gm.ProtoItem{newStepItem(false, false, []*gm.Fragment{newTextFragment("Say Hi")})},
				ConceptExecutionResult: &gm.ProtoStepExecutionResult{
					ExecutionResult: &gm.ProtoExecutionResult{Failed: false, ExecutionTime: 211316},
				},
			},
		},
		newStepItem(false, false, []*gm.Fragment{
			newTextFragment("Say "),
			newParamFragment(newStaticParam("hi")),
			newTextFragment(" to "),
			newParamFragment(newDynamicParam("gauge")),
			newParamFragment(newTableParam([]string{"Word", "Count"}, [][]string{
				[]string{"Gauge", "3"},
				[]string{"Mingle", "2"},
			})),
		}),
	},
	ConceptExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{Failed: false, ExecutionTime: 211316},
	},
}

var protoStepWithSpecialParams = &gm.ProtoStep{
	Fragments: []*gm.Fragment{
		newTextFragment("Say "),
		{
			FragmentType: gm.Fragment_Parameter,
			Parameter: &gm.Parameter{
				Name:          "file:foo.txt",
				ParameterType: gm.Parameter_Special_String,
				Value:         "hi",
			},
		},
		newTextFragment(" to "),
		{
			FragmentType: gm.Fragment_Parameter,
			Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Special_Table,
				Name:          "table:myTable.csv",
				Table: newTableItem([]string{"Word", "Count"}, [][]string{
					[]string{"Gauge", "3"},
					[]string{"Mingle", "2"},
				}).GetTable(),
			},
		},
	},
	StepExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{
			Failed:        false,
			ExecutionTime: 211316,
		},
	},
}

var protoStepWithAfterHookFailure = &gm.ProtoStep{
	Fragments: []*gm.Fragment{newTextFragment("Some Step")},
	StepExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{
			Failed:        true,
			ExecutionTime: 211316,
		},
		PostHookFailure: &gm.ProtoHookFailure{
			ErrorMessage: "err",
			StackTrace:   "Stacktrace",
			ScreenShot:   []byte("Screenshot"),
		},
	},
}

var failedHookFailure = &gm.ProtoHookFailure{
	ErrorMessage: "java.lang.RuntimeException",
	StackTrace:   newStackTrace(),
	ScreenShot:   []byte(newScreenshot()),
}

func TestToOverview(t *testing.T) {
	want := &overview{
		ProjectName:   "projName",
		Env:           "ci-java",
		Tags:          "!unimplemented",
		SuccessRate:   80.00,
		ExecutionTime: "00:01:53",
		Timestamp:     "Jun 3, 2016 at 12:29pm",
		Summary:       &summary{Total: 15, Failed: 2, Passed: 8, Skipped: 5},
	}

	got := toOverview(suiteRes1, "")
	checkEqual(t, "", want, got)
}

func TestToSidebar(t *testing.T) {

	want := &sidebar{
		IsBeforeHookFailure: false,
		Specs: []*specsMeta{
			newSpecsMeta("specRes2", "00:03:31", true, false, []string{"tag1", "tag2", "tag3"}, "specRes2.html"),
			newSpecsMeta("specRes3", "00:03:31", false, true, []string{"tag1"}, "specRes3.html"),
			newSpecsMeta("specRes1", "00:03:31", false, false, []string{"tag1", "tag2"}, "foobar.html"),
		},
	}

	got := toSidebar(suiteRes2, "")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%v\ngot:\n%v\n", want, got)
	}
}

func TestToSpecHeader(t *testing.T) {
	want := &specHeader{
		SpecName:      "specRes1",
		ExecutionTime: "00:03:31",
		FileName:      "/tmp/gauge/specs/foobar.spec",
		Tags:          []string{"tag1", "tag2"},
		Summary:       &summary{},
	}

	got := toSpecHeader(spec1)
	checkEqual(t, "", want, got)
}

func TestToSpec(t *testing.T) {
	want := &spec{
		CommentsBeforeDatatable: []string{"\n", "This is an executable specification file. This file follows markdown syntax.", "\n", "To execute this specification, run", "\tgauge specs", "\n"},
		Datatable: &table{
			Headers: []string{"Word", "Count"},
			Rows:    []*row{{Cells: []string{"Gauge", "3"}, Result: skip}, {Cells: []string{"Mingle", "2"}, Result: skip}},
		},
		CommentsAfterDatatable: []string{"Comment 1", "Comment 2", "Comment 3"},
		Scenarios:              make([]*scenario, 0),
		Errors:                 make([]buildError, 0),
		SpecHeading:            "specRes1",
		Tags:                   []string{"tag1", "tag2"},
		FileName:               "/tmp/gauge/specs/foobar.spec",
		IsTableDriven:          true,
		ExecutionStatus:        pass,
		ExecutionTime:          211316,
	}

	got := toSpec(specRes1)
	checkEqual(t, "", want, got)
}

func TestToSpecWithScenariosInOrder(t *testing.T) {
	specRes := &gm.ProtoSpecResult{
		Failed:        true,
		Skipped:       false,
		ExecutionTime: 211316,
		ProtoSpec: &gm.ProtoSpec{
			SpecHeading: "specRes1",
			FileName:    "/tmp/gauge/specs/foobar.spec",
			Items: []*gm.ProtoItem{
				newScenarioItem(scn),
				newScenarioItem(scnWithHookFailure),
				newScenarioItem(scnWithHookFailure),
				newScenarioItem(skippedProtoSce),
				newScenarioItem(scn),
			},
		},
	}

	got := toSpec(specRes)
	if len(got.Scenarios) != 5 {
		t.Errorf("want:%q\ngot:%q\n", 5, len(got.Scenarios))
	}
	if got.Scenarios[0].ExecutionStatus != fail {
		t.Errorf("want:%q\ngot:%q\n", fail, got.Scenarios[0].ExecutionStatus)
	}
	if got.Scenarios[1].ExecutionStatus != fail {
		t.Errorf("want:%q\ngot:%q\n", fail, got.Scenarios[1].ExecutionStatus)
	}
	if got.Scenarios[2].ExecutionStatus != skip {
		t.Errorf("want:%q\ngot:%q\n", skip, got.Scenarios[2].ExecutionStatus)
	}
	if got.Scenarios[3].ExecutionStatus != pass {
		t.Errorf("want:%q\ngot:%q\n", pass, got.Scenarios[3].ExecutionStatus)
	}
	if got.Scenarios[4].ExecutionStatus != pass {
		t.Errorf("want:%q\ngot:%q\n", pass, got.Scenarios[4].ExecutionStatus)
	}
}

func TestToSpecWithErrors(t *testing.T) {
	specRes := &gm.ProtoSpecResult{
		Failed: true,
		Errors: []*gm.Error{
			{
				Filename:   "fileName",
				LineNumber: 2,
				Message:    "message",
				Type:       gm.Error_PARSE_ERROR,
			},
			{
				Filename:   "fileName1",
				LineNumber: 4,
				Message:    "message1",
				Type:       gm.Error_VALIDATION_ERROR,
			},
		},
	}

	want := &spec{
		Errors: []buildError{
			{FileName: "fileName", LineNumber: 2, Message: "message", ErrorType: parseErrorType},
			{FileName: "fileName1", LineNumber: 4, Message: "message1", ErrorType: validationErrorType},
		},
		Scenarios:       make([]*scenario, 0),
		ExecutionStatus: fail,
	}

	got := toSpec(specRes)

	checkEqual(t, "", want, got)
}

func TestToSpecForTableDrivenSpec(t *testing.T) {
	want := &spec{
		Datatable: &table{
			Headers: []string{"Word", "Count"},
			Rows:    []*row{{Cells: []string{"Gauge", "3"}, Result: fail}, {Cells: []string{"Mingle", "2"}, Result: pass}},
		},
		SpecHeading:     "specRes1",
		FileName:        "/tmp/gauge/specs/foobar.spec",
		IsTableDriven:   true,
		ExecutionStatus: pass,
		ExecutionTime:   211316,
		Scenarios: []*scenario{
			&scenario{
				Heading:       "Scenario 1",
				ExecutionTime: "00:00:00",
				Items: []item{
					item{
						Kind: stepKind,
						Step: &step{
							Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Step1"}},
							Result:    &result{Status: fail, ExecutionTime: "00:03:31"},
						},
					},
				},
				Contexts:                  make([]item, 0),
				Teardowns:                 make([]item, 0),
				ExecutionStatus:           fail,
				TableRowIndex:             0,
				BeforeScenarioHookFailure: nil,
				AfterScenarioHookFailure:  nil,
			},
			&scenario{
				Heading:       "Scenario 1",
				ExecutionTime: "00:00:00",
				Items: []item{
					item{
						Kind: stepKind,
						Step: &step{
							Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Step1"}},
							Result:    &result{Status: pass, ExecutionTime: "00:03:31"},
						},
					},
				},
				Contexts:                  make([]item, 0),
				Teardowns:                 make([]item, 0),
				ExecutionStatus:           pass,
				TableRowIndex:             1,
				BeforeScenarioHookFailure: nil,
				AfterScenarioHookFailure:  nil,
			},
		},
		BeforeSpecHookFailure: nil,
		AfterSpecHookFailure:  nil,
		Errors:                make([]buildError, 0),
		PassedScenarioCount:   1,
		FailedScenarioCount:   1,
		SkippedScenarioCount:  0,
	}

	got := toSpec(datatableDrivenSpec)

	checkEqual(t, "", want, got)
}

func TestToSpecWithHookFailure(t *testing.T) {
	encodedScreenShot := base64.StdEncoding.EncodeToString([]byte("Screenshot"))
	want := &spec{
		Scenarios:             make([]*scenario, 0),
		BeforeSpecHookFailure: []*hookFailure{newHookFailure("Before Spec", "err", encodedScreenShot, "Stacktrace")},
		AfterSpecHookFailure:  []*hookFailure{newHookFailure("After Spec", "err", encodedScreenShot, "Stacktrace")},
		Errors:                make([]error, 0),
		Tags:                  []string{"tag1"},
		SpecHeading:           "specRes3",
		ExecutionStatus:       skip,
		ExecutionTime:         211316,
	}

	got := toSpec(specResWithSpecHookFailure)
	checkEqual(t, "", want, got)
}

func TestToSpecWithFileName(t *testing.T) {
	want := specRes1.GetProtoSpec().GetFileName()
	got := toSpec(specRes1).FileName

	if got != want {
		t.Errorf("Expecting spec.FileName=%s, got %s\n", want, got)
	}
}

func TestToSpecWithSpecHeading(t *testing.T) {
	want := specRes1.GetProtoSpec().GetSpecHeading()
	got := toSpec(specRes1).SpecHeading

	if got != want {
		t.Errorf("Expecting spec.SpecHeading=%s, got %s\n", want, got)
	}
}

func TestToSpecWithTags(t *testing.T) {
	want := specRes1.GetProtoSpec().GetTags()
	got := toSpec(specRes1).Tags

	checkEqual(t, "", want, got)
}

func TestToSpecWithDataTableMapsCommentsBeforeDatatable(t *testing.T) {
	want := []string{
		"\n",
		"This is an executable specification file. This file follows markdown syntax.",
		"\n",
		"To execute this specification, run",
		"\tgauge specs",
		"\n",
	}
	got := toSpec(specRes1).CommentsBeforeDatatable

	checkEqual(t, "", want, got)
}

func TestToSpecWithDataTableMapsCommentsAfterDatatable(t *testing.T) {
	want := []string{
		"Comment 1",
		"Comment 2",
		"Comment 3",
	}
	got := toSpec(specRes1).CommentsAfterDatatable

	checkEqual(t, "", want, got)
}

func TestToSpecWithDataTableIsTableDriven(t *testing.T) {
	got := toSpec(specRes1).IsTableDriven

	if !got {
		t.Errorf("Expecting spec.IsTableDriven=true\n")
	}
}

func TestToSpecWithDataTableHasDatatable(t *testing.T) {
	want := &table{
		Headers: []string{"Word", "Count"},
		Rows: []*row{
			&row{Cells: []string{"Gauge", "3"}, Result: skip},
			&row{Cells: []string{"Mingle", "2"}, Result: skip},
		},
	}
	got := toSpec(specRes1).Datatable

	checkEqual(t, "", want, got)
}

func TestToSpecWithDataTableExecutionTime(t *testing.T) {
	want := 211316
	got := toSpec(specRes1).ExecutionTime

	checkEqual(t, "", want, got)
}

func TestToSpecWithDataTableExecutionStatusPass(t *testing.T) {
	want := pass
	got := toSpec(specRes1).ExecutionStatus

	checkEqual(t, "", want, got)
}

func TestToSpecWithDataTableExecutionStatusSkip(t *testing.T) {
	want := skip
	got := toSpec(&gm.ProtoSpecResult{Skipped: true, Failed: false}).ExecutionStatus

	checkEqual(t, "", want, got)
}

func TestToSpecWithDataTableExecutionStatusFail(t *testing.T) {
	want := fail
	got := toSpec(&gm.ProtoSpecResult{Skipped: false, Failed: true}).ExecutionStatus

	checkEqual(t, "", want, got)
}

func TestToSpecWithBeforeHookFailure(t *testing.T) {
	want := []*hookFailure{{ErrMsg: "err", HookName: "Before Spec", Screenshot: "U2NyZWVuc2hvdA==", StackTrace: "Stacktrace"}}
	got := toSpec(specResWithSpecHookFailure).BeforeSpecHookFailure

	checkEqual(t, "", want, got)
}

func TestToSpecWithAfterHookFailure(t *testing.T) {
	want := []*hookFailure{{ErrMsg: "err", HookName: "After Spec", Screenshot: "U2NyZWVuc2hvdA==", StackTrace: "Stacktrace", TableRowIndex: 0}}
	got := toSpec(specResWithSpecHookFailure).AfterSpecHookFailure

	checkEqual(t, "", want, got)
}

func TestToSpecWithScenarios(t *testing.T) {
	got := toSpec(&gm.ProtoSpecResult{
		ProtoSpec: &gm.ProtoSpec{
			Items: []*gm.ProtoItem{
				&gm.ProtoItem{
					ItemType: gm.ProtoItem_TableDrivenScenario,
					TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario: &gm.ProtoScenario{
							ScenarioHeading: "Scenario 1",
							ExecutionStatus: gm.ExecutionStatus_FAILED,
							ScenarioItems:   []*gm.ProtoItem{newStepItem(true, false, []*gm.Fragment{newTextFragment("Step1")})},
						},
						TableRowIndex: int32(0),
					},
				},
				&gm.ProtoItem{
					ItemType: gm.ProtoItem_TableDrivenScenario,
					TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario: &gm.ProtoScenario{
							ScenarioHeading: "Scenario 1",
							ExecutionStatus: gm.ExecutionStatus_PASSED,
							ScenarioItems:   []*gm.ProtoItem{newStepItem(false, false, []*gm.Fragment{newTextFragment("Step1")})},
						},
						TableRowIndex: int32(1),
					},
				},
			},
		},
	}).Scenarios

	if len(got) != 2 {
		t.Errorf("Expected 2 scenarios, got %d\n", len(got))
	}
}

func TestToSpecWithScenariosTableDriven(t *testing.T) {
	got := toSpec(datatableDrivenSpec).Scenarios

	if len(got) != 2 {
		t.Errorf("Expected 2 scenarios, got %d\n", len(got))
	}
}

func TestToSpecWithScenarioStatusCounts(t *testing.T) {
	got := toSpec(&gm.ProtoSpecResult{
		ProtoSpec: &gm.ProtoSpec{
			Items: []*gm.ProtoItem{
				&gm.ProtoItem{
					ItemType: gm.ProtoItem_TableDrivenScenario,
					TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario: &gm.ProtoScenario{
							ScenarioHeading: "Scenario 1",
							ExecutionStatus: gm.ExecutionStatus_FAILED,
							ScenarioItems:   []*gm.ProtoItem{newStepItem(true, false, []*gm.Fragment{newTextFragment("Step1")})},
						},
						TableRowIndex: int32(0),
					},
				},
				&gm.ProtoItem{
					ItemType: gm.ProtoItem_TableDrivenScenario,
					TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario: &gm.ProtoScenario{
							ScenarioHeading: "Scenario 1",
							ExecutionStatus: gm.ExecutionStatus_SKIPPED,
							ScenarioItems:   []*gm.ProtoItem{newStepItem(true, false, []*gm.Fragment{newTextFragment("Step1")})},
						},
						TableRowIndex: int32(0),
					},
				},
				&gm.ProtoItem{
					ItemType: gm.ProtoItem_TableDrivenScenario,
					TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario: &gm.ProtoScenario{
							ScenarioHeading: "Scenario 1",
							ExecutionStatus: gm.ExecutionStatus_PASSED,
							ScenarioItems:   []*gm.ProtoItem{newStepItem(false, false, []*gm.Fragment{newTextFragment("Step1")})},
						},
						TableRowIndex: int32(1),
					},
				},
			},
		},
	})

	if got.PassedScenarioCount != 1 {
		t.Errorf("Expected spec.PassedScenarioCount=1, got %d\n", got.PassedScenarioCount)
	}
	if got.SkippedScenarioCount != 1 {
		t.Errorf("Expected spec.SkippedScenarioCount=1, got %d\n", got.SkippedScenarioCount)
	}
	if got.FailedScenarioCount != 1 {
		t.Errorf("Expected spec.FailedScenarioCount=1, got %d\n", got.FailedScenarioCount)
	}
}

type summaryTest struct {
	name     string
	result   *spec
	expected summary
}

var summaryTests = []*summaryTest{
	{"All Passed",
		&spec{
			PassedScenarioCount:  2,
			SkippedScenarioCount: 0,
			FailedScenarioCount:  0,
		},
		summary{Failed: 0, Passed: 2, Skipped: 0, Total: 2},
	},
	{"With Skipped",
		&spec{
			PassedScenarioCount:  1,
			SkippedScenarioCount: 1,
			FailedScenarioCount:  0,
		},
		summary{Failed: 0, Passed: 1, Skipped: 1, Total: 2},
	},
	{"With failed",
		&spec{
			PassedScenarioCount:  1,
			SkippedScenarioCount: 0,
			FailedScenarioCount:  1,
		},
		summary{Failed: 1, Passed: 1, Skipped: 0, Total: 2},
	},
	{"With failed and skipped",
		&spec{
			PassedScenarioCount:  0,
			SkippedScenarioCount: 1,
			FailedScenarioCount:  1,
		},
		summary{Failed: 1, Passed: 0, Skipped: 1, Total: 2},
	},
}

func TestToScenarioSummary(t *testing.T) {
	for _, test := range summaryTests {
		want := test.expected
		got := *toScenarioSummary(test.result)
		checkEqual(t, test.name, want, got)
	}
}

func TestToScenario(t *testing.T) {
	want := &scenario{
		Heading:         "Vowel counts in single word",
		ExecutionTime:   "00:01:53",
		ExecutionStatus: pass,
		Tags:            []string{"foo", "bar"},
		Contexts: []item{
			item{
				Kind: stepKind,
				Step: &step{
					Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Context Step1"}},
					Result:    &result{Status: pass, ExecutionTime: "00:03:31"},
				},
			},
			item{
				Kind: stepKind,
				Step: &step{
					Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Context Step2"}},
					Result:    &result{Status: fail, ExecutionTime: "00:03:31"},
				},
			},
		},
		Items: []item{
			item{
				Kind:    commentKind,
				Comment: &comment{Text: "Comment0"},
			},
			item{
				Kind: stepKind,
				Step: &step{
					Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Step1"}},
					Result:    &result{Status: fail, ExecutionTime: "00:03:31"},
				},
			},
			item{
				Kind:    commentKind,
				Comment: &comment{Text: "Comment1"},
			},
			item{
				Kind:    commentKind,
				Comment: &comment{Text: "Comment2"},
			},
			item{
				Kind: stepKind,
				Step: &step{
					Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Step2"}},
					Result:    &result{Status: pass, ExecutionTime: "00:03:31"},
				},
			},
			item{
				Kind:    commentKind,
				Comment: &comment{Text: "Comment3"},
			},
		},
		Teardowns: []item{
			item{
				Kind: stepKind,
				Step: &step{
					Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Teardown Step1"}},
					Result:    &result{Status: pass, ExecutionTime: "00:03:31"},
				},
			},
			item{
				Kind: stepKind,
				Step: &step{
					Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Teardown Step2"}},
					Result:    &result{Status: fail, ExecutionTime: "00:03:31"},
				},
			},
		},
		TableRowIndex: -1,
	}

	got := toScenario(scn, -1)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToScenarioWithHookFailures(t *testing.T) {
	encodedScreenShot := base64.StdEncoding.EncodeToString([]byte("Screenshot"))
	want := &scenario{
		Heading:         "Vowel counts in single word",
		ExecutionTime:   "00:01:53",
		ExecutionStatus: fail,
		Contexts:        []item{},
		Items: []item{
			item{
				Kind: stepKind,
				Step: &step{
					Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Step1"}},
					Result:    &result{Status: fail, ExecutionTime: "00:03:31"},
				},
			},
		},
		Teardowns:                 []item{},
		BeforeScenarioHookFailure: newHookFailure("Before Scenario", "err", encodedScreenShot, "Stacktrace"),
		AfterScenarioHookFailure:  newHookFailure("After Scenario", "err", encodedScreenShot, "Stacktrace"),
		TableRowIndex:             -1,
	}

	got := toScenario(scnWithHookFailure, -1)
	checkEqual(t, "", want, got)
}

func TestToConcept(t *testing.T) {
	want := &concept{
		ConceptStep: &step{
			Fragments: []*fragment{
				{FragmentKind: textFragmentKind, Text: "Say "},
				{FragmentKind: dynamicFragmentKind, Text: "hello"},
				{FragmentKind: textFragmentKind, Text: " to "},
				{FragmentKind: tableFragmentKind,
					Table: &table{
						Headers: []string{"Word", "Count"},
						Rows:    []*row{{Cells: []string{"Gauge", "3"}, Result: pass}, {Cells: []string{"Mingle", "2"}, Result: pass}},
					},
				},
			},
			Result: &result{Status: pass, ExecutionTime: "00:03:31"},
		},
		Items: []item{
			item{
				Kind: conceptKind,
				Concept: &concept{
					ConceptStep: &step{
						Fragments: []*fragment{
							{FragmentKind: textFragmentKind, Text: "Tell "},
							{FragmentKind: dynamicFragmentKind, Text: "hello"},
						},
						Result: &result{Status: pass, ExecutionTime: "00:03:31"},
					},
					Items: []item{
						item{
							Kind: stepKind,
							Step: &step{
								Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Say Hi"}},
								Result:    &result{Status: pass, ExecutionTime: "00:03:31"},
							},
						},
					},
				},
			},
			item{
				Kind: stepKind,
				Step: &step{
					Fragments: []*fragment{
						{FragmentKind: textFragmentKind, Text: "Say "},
						{FragmentKind: staticFragmentKind, Text: "hi"},
						{FragmentKind: textFragmentKind, Text: " to "},
						{FragmentKind: dynamicFragmentKind, Text: "gauge"},
						{FragmentKind: tableFragmentKind,
							Table: &table{
								Headers: []string{"Word", "Count"},
								Rows:    []*row{{Cells: []string{"Gauge", "3"}, Result: pass}, {Cells: []string{"Mingle", "2"}, Result: pass}},
							},
						},
					},
					Result: &result{Status: pass, ExecutionTime: "00:03:31"},
				},
			},
		},
	}

	got := toConcept(protoConcept)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToStep(t *testing.T) {
	want := &step{
		Fragments: []*fragment{
			{FragmentKind: textFragmentKind, Text: "Say "},
			{FragmentKind: staticFragmentKind, Text: "hi"},
			{FragmentKind: textFragmentKind, Text: " to "},
			{FragmentKind: dynamicFragmentKind, Text: "gauge"},
			{FragmentKind: tableFragmentKind,
				Table: &table{
					Headers: []string{"Word", "Count"},
					Rows:    []*row{{Cells: []string{"Gauge", "3"}, Result: pass}, {Cells: []string{"Mingle", "2"}, Result: pass}},
				},
			},
		},
		Result: &result{Status: skip, ExecutionTime: "00:03:31", SkippedReason: "Step impl not found"},
	}

	got := toStep(protoStep)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToCSV(t *testing.T) {
	table := newTableItem([]string{"Word", "Count"}, [][]string{
		[]string{"Gauge", "3"},
		[]string{"Mingle", "2"},
	}).GetTable()

	want := "Word,Count\n" +
		"Gauge,3\n" +
		"Mingle,2"

	got := toCsv(table)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToStepWithSpecialParams(t *testing.T) {
	want := &step{
		Fragments: []*fragment{
			{FragmentKind: textFragmentKind, Text: "Say "},
			{FragmentKind: specialStringFragmentKind, Name: "file:foo.txt", Text: "hi", FileName: "foo.txt"},
			{FragmentKind: textFragmentKind, Text: " to "},
			{FragmentKind: specialTableFragmentKind,
				Name: "table:myTable.csv",
				Text: `Word,Count
Gauge,3
Mingle,2`,
				FileName: "myTable.csv",
			},
		},
		Result: &result{
			Status:        pass,
			ExecutionTime: "00:03:31",
		},
	}

	got := toStep(protoStepWithSpecialParams)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToStepWithAfterHookFailure(t *testing.T) {
	encodedScreenShot := base64.StdEncoding.EncodeToString([]byte("Screenshot"))
	want := &step{
		Fragments: []*fragment{
			{FragmentKind: textFragmentKind, Text: "Some Step"},
		},
		Result: &result{
			Status:        fail,
			ExecutionTime: "00:03:31",
		},
		AfterStepHookFailure: newHookFailure("After Step", "err", encodedScreenShot, "Stacktrace"),
	}

	got := toStep(protoStepWithAfterHookFailure)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToComment(t *testing.T) {
	want := &comment{Text: "Whatever"}

	got := toComment(newCommentItem("Whatever").GetComment())
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToHookFailure(t *testing.T) {
	encodedScreenShot := base64.StdEncoding.EncodeToString([]byte(newScreenshot()))
	want := newHookFailure("Before Suite", "java.lang.RuntimeException", encodedScreenShot, newStackTrace())

	got := toHookFailure(failedHookFailure, "Before Suite")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToHookFailureWithNilInput(t *testing.T) {
	var want *hookFailure = nil
	got := toHookFailure(nil, "foobar")

	if got != want {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

type specNameGenerationTest struct {
	specName     string
	projectRoot  string
	HTMLFilename string
}

var specNameGenerationTests = []*specNameGenerationTest{
	{filepath.Join("Users", "gauge", "foo", "simple_specification.spec"), filepath.Join("Users", "gauge", "foo"), "simple_specification.html"},
	{filepath.Join("Users", "gauge", "foo", "simple_specification.spec"), filepath.Join("Users", "gauge"), filepath.Join("foo", "simple_specification.html")},
	{"simple_specification.spec", "", "simple_specification.html"},
	{filepath.Join("Users", "gauge", "foo", "abcd1234.spec"), filepath.Join("Users", "gauge", "foo"), "abcd1234.html"},
	{filepath.Join("Users", "gauge", "foo", "bar", "simple_specification.spec"), filepath.Join("Users", "gauge", "foo"), filepath.Join("bar", "simple_specification.html")},
	{filepath.Join("Users", "gauge", "foo", "bar", "simple_specification.spec"), "Users", filepath.Join("gauge", "foo", "bar", "simple_specification.html")},
	{filepath.Join("Users", "gauge12", "fo_o", "b###$ar", "simple_specification.spec"), "Users", filepath.Join("gauge12", "fo_o", "b###$ar", "simple_specification.html")},
}

func TestToHTMLFileName(t *testing.T) {
	for _, test := range specNameGenerationTests {
		got := toHTMLFileName(test.specName, test.projectRoot)
		want := test.HTMLFilename
		if got != want {
			t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
		}
	}
}

func TestGetSpecNameWhenHeadingIsPresent(t *testing.T) {
	want := "heading"

	got := getSpecName(&gm.ProtoSpec{SpecHeading: "heading"})

	if got != want {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestGetSpecNameWhenHeadingIsNotPresent(t *testing.T) {
	want := "example.spec"

	got := getSpecName(&gm.ProtoSpec{FileName: filepath.Join("specs", "specs1", "example.spec")})

	if got != want {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

type tableDrivenStatusComputeTest struct {
	name   string
	spec   *spec
	status status
}

var tableDrivenStatusComputeTests = []*tableDrivenStatusComputeTest{
	{"all passed",
		&spec{Datatable: &table{Headers: []string{"foo"}, Rows: []*row{{Cells: []string{"foo1"}}}},
			Scenarios: []*scenario{
				{ExecutionStatus: pass, TableRowIndex: 0},
				{ExecutionStatus: pass, TableRowIndex: 0},
			}},
		pass},
	{"pass and fail",
		&spec{Datatable: &table{Headers: []string{"foo"}, Rows: []*row{{Cells: []string{"foo1"}}}},
			Scenarios: []*scenario{
				{ExecutionStatus: pass, TableRowIndex: 0},
				{ExecutionStatus: fail, TableRowIndex: 0},
			}},
		fail},
	{"pass and skip",
		&spec{Datatable: &table{Headers: []string{"foo"}, Rows: []*row{{Cells: []string{"foo1"}}}},
			Scenarios: []*scenario{
				{ExecutionStatus: pass, TableRowIndex: 0},
				{ExecutionStatus: skip, TableRowIndex: 0},
			}},
		pass},
	{"skip and fail",
		&spec{Datatable: &table{Headers: []string{"foo"}, Rows: []*row{{Cells: []string{"foo1"}}}},
			Scenarios: []*scenario{
				{ExecutionStatus: skip, TableRowIndex: 0},
				{ExecutionStatus: fail, TableRowIndex: 0},
			}},
		fail},
	{"all fail",
		&spec{Datatable: &table{Headers: []string{"foo"}, Rows: []*row{{Cells: []string{"foo1"}}}},
			Scenarios: []*scenario{
				{ExecutionStatus: fail, TableRowIndex: 0},
				{ExecutionStatus: fail, TableRowIndex: 0},
			}},
		fail},
	{"all skip",
		&spec{Datatable: &table{Headers: []string{"foo"}, Rows: []*row{{Cells: []string{"foo1"}}}},
			Scenarios: []*scenario{
				{ExecutionStatus: skip, TableRowIndex: 0},
				{ExecutionStatus: skip, TableRowIndex: 0},
			}},
		skip},
}

func TestTableDrivenStatusCompute(t *testing.T) {
	for _, test := range tableDrivenStatusComputeTests {
		want := test.status
		computeTableDrivenStatuses(test.spec)
		got := test.spec.Datatable.Rows[0].Result
		if want != got {
			t.Errorf("test: %s want:\n%q\ngot:\n%q\n", test.name, want, got)
		}
	}
}

func TestMapProjectNametoSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{ProjectName: "foo"}
	res := ToSuiteResult("", psr)

	if res.ProjectName != "foo" {
		t.Errorf("Expected ProjectName=foo, got %s", res.ProjectName)
	}
}

func TestMapEnvironmenttoSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{Environment: "foo"}
	res := ToSuiteResult("", psr)

	if res.Environment != "foo" {
		t.Errorf("Expected Environment=foo, got %s", res.Environment)
	}
}

func TestMapTagstoSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{Tags: "foo, bar"}
	res := ToSuiteResult("", psr)

	if res.Tags != "foo, bar" {
		t.Errorf("Expected Tags=foo, bar; got %s", res.Tags)
	}
}

func TestMapExecutionTimeToSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{ExecutionTime: 113163}
	res := ToSuiteResult("", psr)

	if res.ExecutionTime != 113163 {
		t.Errorf("Expected ExecutionTime=113163; got %s", res.ExecutionTime)
	}
}

func TestSpecsCountToSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{SpecsFailedCount: 2, SpecsSkippedCount: 1, SpecResults: []*gm.ProtoSpecResult{
		{Skipped: false, Failed: false},
		{Skipped: false, Failed: true},
		{Skipped: true, Failed: false},
		{Skipped: false, Failed: true},
		{Skipped: false, Failed: false},
		{Skipped: false, Failed: false},
	}}
	res := ToSuiteResult("", psr)

	if res.PassedSpecsCount != 3 {
		t.Errorf("Expected PassedSpecsCount=3; got %s\n", res.PassedSpecsCount)
	}
	if res.SkippedSpecsCount != 1 {
		t.Errorf("Expected SkippedSpecsCount=3; got %s\n", res.SkippedSpecsCount)
	}
	if res.FailedSpecsCount != 2 {
		t.Errorf("Expected FailedSpecsCount=3; got %s\n", res.FailedSpecsCount)
	}
}

func TestMapPreHookFailureToSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{PreHookFailure: &gm.ProtoHookFailure{ErrorMessage: "foo failure"}}
	res := ToSuiteResult("", psr)

	if res.BeforeSuiteHookFailure == nil {
		t.Errorf("Expected BeforeSuiteHookFailure not nil\n")
	}

	if res.BeforeSuiteHookFailure.ErrMsg != "foo failure" {
		t.Errorf("Expected BeforeSuiteHookFailure.Message= 'foo failure', got %s\n", res.BeforeSuiteHookFailure.ErrMsg)
	}
}

func TestMapPostHookFailureToSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{PostHookFailure: &gm.ProtoHookFailure{ErrorMessage: "foo failure"}}
	res := ToSuiteResult("", psr)

	if res.AfterSuiteHookFailure == nil {
		t.Errorf("Expected AfterSuiteHookFailure not nil\n")
	}

	if res.AfterSuiteHookFailure.ErrMsg != "foo failure" {
		t.Errorf("Expected AfterSuiteHookFailure.Message= 'foo failure', got %s\n", res.AfterSuiteHookFailure.ErrMsg)
	}
}

func TestMapTimestampToSuiteResult(t *testing.T) {
	timestamp := "Jun 3, 2016 at 12:29pm"
	psr := &gm.ProtoSuiteResult{Timestamp: timestamp}
	res := ToSuiteResult("", psr)

	if res.Timestamp != timestamp {
		t.Errorf("Expected Timestamp=%s; got %s\n", timestamp, res.Timestamp)
	}
}

func TestMapExecutionStatusPassByDefaultToSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{}
	res := ToSuiteResult("", psr)

	if res.ExecutionStatus != pass {
		t.Errorf("Expected ExecutionStatus=pass, got %s\n", res.ExecutionStatus)
	}
}

func TestMapExecutionStatusOnFailureToSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{Failed: true}
	res := ToSuiteResult("", psr)

	if res.ExecutionStatus != fail {
		t.Errorf("Expected ExecutionStatus=fail, got %s\n", res.ExecutionStatus)
	}
}

func TestMapSpecResultsToSuiteResult(t *testing.T) {
	psr := &gm.ProtoSuiteResult{
		SpecResults: []*gm.ProtoSpecResult{
			{ProtoSpec: &gm.ProtoSpec{FileName: "foo.spec", SpecHeading: "Foo Spec"}},
			{ProtoSpec: &gm.ProtoSpec{FileName: "bar.spec", SpecHeading: "Boo Spec"}},
			{ProtoSpec: &gm.ProtoSpec{FileName: "baz.spec", SpecHeading: "Baz Spec"}},
		},
	}

	res := ToSuiteResult("", psr)

	if len(res.SpecResults) != 3 {
		t.Errorf("Expected 3 spec results, got %d\n", len(res.SpecResults))
	}

	for i, s := range res.SpecResults {
		wantFileName := psr.GetSpecResults()[i].GetProtoSpec().GetFileName()
		gotFileName := s.FileName
		if wantFileName != gotFileName {
			t.Errorf("Spec Filename Mismatch, want '%s', got '%s'", wantFileName, gotFileName)
		}
	}
}

func TestToNestedSuiteResultMapsProjectName(t *testing.T) {
	want := "foo"
	sr := &SuiteResult{ProjectName: want}

	got := toNestedSuiteResult("some/path", sr)

	if got.ProjectName != want {
		t.Fatalf("Expected ProjectName=%s, got %s", want, got)
	}
}

func TestToNestedSuiteResultMapsTimeStamp(t *testing.T) {
	want := "Jul 13, 2016 at 11:49am"
	sr := &SuiteResult{Timestamp: want}

	got := toNestedSuiteResult("some/path", sr)

	if got.Timestamp != want {
		t.Fatalf("Expected TimeStamp=%s, got %s", want, got)
	}
}

func TestToNestedSuiteResultMapsEnvironment(t *testing.T) {
	want := "foo"
	sr := &SuiteResult{Environment: want}

	got := toNestedSuiteResult("some/path", sr)

	if got.Environment != want {
		t.Fatalf("Expected Environment=%s, got %s", want, got)
	}
}

func TestToNestedSuiteResultMapsTags(t *testing.T) {
	want := "tag1, tag2"
	sr := &SuiteResult{Tags: want}

	got := toNestedSuiteResult("some/path", sr)

	if got.Tags != want {
		t.Fatalf("Expected Tags=%s, got %s", want, got.Tags)
	}
}

func TestToNestedSuiteResultMapsBeforeSuiteHookFailure(t *testing.T) {
	want := "fooHookFailure"
	sr := &SuiteResult{BeforeSuiteHookFailure: &hookFailure{HookName: want}}

	got := toNestedSuiteResult("some/path", sr)

	if got.BeforeSuiteHookFailure == nil {
		t.Fatal("Expected BeforeSuiteHookFailure to be not nil")
	}

	if got.BeforeSuiteHookFailure.HookName != want {
		t.Fatalf("expected BeforeSuiteHookFailure to have HookName = %s, got %s", want, got)
	}
}

func TestToNestedSuiteResultMapsAfterSuiteHookFailure(t *testing.T) {
	want := "fooHookFailure"
	sr := &SuiteResult{AfterSuiteHookFailure: &hookFailure{HookName: want}}

	got := toNestedSuiteResult("some/path", sr)

	if got.AfterSuiteHookFailure == nil {
		t.Fatal("Expected AfterSuiteHookFailure to be not nil")
	}

	if got.AfterSuiteHookFailure.HookName != want {
		t.Fatalf("expected AfterSuiteHookFailure to have HookName = %s, got %s", want, got)
	}
}

func TestToNestedSuiteResultMapsExecutionTime(t *testing.T) {
	want := int64(2000)
	sr := &SuiteResult{SpecResults: []*spec{
		{FileName: "nested1/foo.spec", SpecHeading: "Foo Spec", ExecutionTime: 1000},
		{FileName: "nested1/nested2/boo.spec", SpecHeading: "Boo Spec", ExecutionTime: 1000},
		{FileName: "bar.spec", SpecHeading: "Bar Spec", ExecutionTime: 1000},
	}}

	got := toNestedSuiteResult("nested1", sr)

	if got.ExecutionTime != want {
		t.Fatalf("expected ExecutionTime = %d, got %d", want, got.ExecutionTime)
	}
}

func TestToNestedSuiteResultMapsPassedCount(t *testing.T) {
	want := 1
	sr := &SuiteResult{SpecResults: []*spec{
		{FileName: "nested1/foo.spec", SpecHeading: "Foo Spec", ExecutionStatus: pass},
		{FileName: "nested1/nested2/boo.spec", SpecHeading: "Boo Spec", ExecutionStatus: skip},
		{FileName: "bar.spec", SpecHeading: "Bar Spec", ExecutionStatus: pass},
	}}

	got := toNestedSuiteResult("nested1", sr)

	if got.PassedSpecsCount != want {
		t.Fatalf("Expected FailedCount=%d, got %d", want, got.PassedSpecsCount)
	}
}

func TestToNestedSuiteResultMapsFailedCount(t *testing.T) {
	want := 1
	sr := &SuiteResult{SpecResults: []*spec{
		{FileName: "nested1/foo.spec", SpecHeading: "Foo Spec", ExecutionStatus: pass},
		{FileName: "nested1/nested2/boo.spec", SpecHeading: "Boo Spec", ExecutionStatus: fail},
		{FileName: "nested1/bar.spec", SpecHeading: "Bar Spec", ExecutionStatus: pass},
	}}

	got := toNestedSuiteResult("nested1", sr)

	if got.FailedSpecsCount != want {
		t.Fatalf("Expected FailedCount=%d, got %d", want, got.FailedSpecsCount)
	}
}

func TestToNestedSuiteResultMapsSkippedCount(t *testing.T) {
	want := 2
	sr := &SuiteResult{SpecResults: []*spec{
		{FileName: "nested1/foo.spec", SpecHeading: "Foo Spec", ExecutionStatus: skip},
		{FileName: "nested1/nested2/boo.spec", SpecHeading: "Boo Spec", ExecutionStatus: skip},
		{FileName: "nested1/bar.spec", SpecHeading: "Bar Spec", ExecutionStatus: fail},
	}}

	got := toNestedSuiteResult("nested1", sr)

	if got.SkippedSpecsCount != want {
		t.Fatalf("Expected SkippedCount=%d, got %d", want, got.SkippedSpecsCount)
	}
}

func TestToNestedSuiteResultMapsSuccessRate(t *testing.T) {
	want := float32(50.0)
	sr := &SuiteResult{SpecResults: []*spec{
		{FileName: "nested1/foo.spec", SpecHeading: "Foo Spec", ExecutionStatus: fail},
		{FileName: "nested1/nested2/boo.spec", SpecHeading: "Boo Spec", ExecutionStatus: skip},
		{FileName: "nested1/nested2/bar.spec", SpecHeading: "Bar Spec", ExecutionStatus: pass},
		{FileName: "nested1/bar.spec", SpecHeading: "Baz Spec", ExecutionStatus: pass},
	}}

	got := toNestedSuiteResult("nested1", sr)

	if got.SuccessRate != want {
		t.Fatalf("Expected SuccessRate=%f, got %f", want, got.SuccessRate)
	}
}

func TestToNestedSuiteResultMapsSpecResults(t *testing.T) {
	fooSpec := spec{FileName: "nested1/foo.spec", SpecHeading: "Foo Spec"}
	booSpec := spec{FileName: "nested1/nested2/boo.spec", SpecHeading: "Boo Spec"}
	barSpec := spec{FileName: "nested1/bar.spec", SpecHeading: "Bar Spec"}
	bazSpec := spec{FileName: "nested2/nested1/baz.spec", SpecHeading: "Baz Spec"}
	specResults := []*spec{
		&fooSpec,
		&booSpec,
		&bazSpec,
		&barSpec,
	}
	got := toNestedSuiteResult("nested1", &SuiteResult{SpecResults: specResults})
	want := []*spec{
		&fooSpec,
		&booSpec,
		&barSpec,
	}

	checkEqual(t, "", want, got.SpecResults)
}
