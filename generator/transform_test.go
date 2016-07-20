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
	"reflect"
	"testing"

	"github.com/getgauge/html-report/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type transformTest struct {
	name   string
	input  *gauge_messages.ProtoSuiteResult
	output interface{}
}

func newCommentItem(str string) *gauge_messages.ProtoItem {
	return &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Comment.Enum(),
		Comment: &gauge_messages.ProtoComment{
			Text: proto.String(str),
		},
	}
}

func newScenarioItem(scn *gauge_messages.ProtoScenario) *gauge_messages.ProtoItem {
	return &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Scenario.Enum(),
		Scenario: scn,
	}
}

func newTableItem(headers []string, rows [][]string) *gauge_messages.ProtoItem {
	r := make([]*gauge_messages.ProtoTableRow, len(rows))
	for i, row := range rows {
		r[i] = &gauge_messages.ProtoTableRow{
			Cells: row,
		}
	}
	return &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Table.Enum(),
		Table: &gauge_messages.ProtoTable{
			Headers: &gauge_messages.ProtoTableRow{
				Cells: headers,
			},
			Rows: r,
		},
	}
}

func newStepItem(failed bool, frags []*gauge_messages.Fragment) *gauge_messages.ProtoItem {
	return &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Step.Enum(),
		Step: &gauge_messages.ProtoStep{
			StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{
				ExecutionResult: &gauge_messages.ProtoExecutionResult{
					Failed:        proto.Bool(failed),
					ExecutionTime: proto.Int64(211316),
				},
			},
			Fragments: frags,
		},
	}
}

var specRes1 = &gauge_messages.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gauge_messages.ProtoSpec{
		SpecHeading:   proto.String("specRes1"),
		Tags:          []string{"tag1", "tag2"},
		FileName:      proto.String("/tmp/gauge/specs/foobar.spec"),
		IsTableDriven: proto.Bool(false),
		Items: []*gauge_messages.ProtoItem{
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

var specRes2 = &gauge_messages.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gauge_messages.ProtoSpec{
		SpecHeading: proto.String("specRes2"),
		Tags:        []string{"tag1", "tag2", "tag3"},
	},
}

var specRes3 = &gauge_messages.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(true),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gauge_messages.ProtoSpec{
		SpecHeading: proto.String("specRes3"),
		Tags:        []string{"tag1"},
	},
}

var suiteRes1 = &gauge_messages.ProtoSuiteResult{
	ProjectName:       proto.String("projName"),
	Environment:       proto.String("ci-java"),
	Tags:              proto.String("!unimplemented"),
	SuccessRate:       proto.Float32(80.00),
	ExecutionTime:     proto.Int64(113163),
	Timestamp:         proto.String("Jun 3, 2016 at 12:29pm"),
	SpecResults:       make([]*gauge_messages.ProtoSpecResult, 15),
	SpecsFailedCount:  proto.Int32(2),
	SpecsSkippedCount: proto.Int32(5),
}

var scn = &gauge_messages.ProtoScenario{
	ScenarioHeading: proto.String("Vowel counts in single word"),
	Failed:          proto.Bool(false),
	Skipped:         proto.Bool(false),
	Tags:            []string{"foo", "bar"},
	ExecutionTime:   proto.Int64(113163),
	Contexts: []*gauge_messages.ProtoItem{
		newStepItem(false, []*gauge_messages.Fragment{{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Context Step1")}}),
		newStepItem(true, []*gauge_messages.Fragment{{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Context Step2")}}),
	},
	ScenarioItems: []*gauge_messages.ProtoItem{
		newCommentItem("Comment0"),
		newStepItem(true, []*gauge_messages.Fragment{{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Step1")}}),
		newCommentItem("Comment1"),
		newCommentItem("Comment2"),
		newStepItem(false, []*gauge_messages.Fragment{{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Step2")}}),
		newCommentItem("Comment3"),
	},
	TearDownSteps: []*gauge_messages.ProtoItem{
		newStepItem(false, []*gauge_messages.Fragment{{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Teardown Step1")}}),
		newStepItem(true, []*gauge_messages.Fragment{{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Teardown Step2")}}),
	},
}

var suiteRes2 = &gauge_messages.ProtoSuiteResult{
	SpecResults: []*gauge_messages.ProtoSpecResult{specRes1, specRes2, specRes3},
}

var protoStep = &gauge_messages.ProtoStep{
	Fragments: []*gauge_messages.Fragment{
		{
			FragmentType: gauge_messages.Fragment_Text.Enum(),
			Text:         proto.String("Say "),
		},
		{
			FragmentType: gauge_messages.Fragment_Parameter.Enum(),
			Parameter: &gauge_messages.Parameter{
				ParameterType: gauge_messages.Parameter_Static.Enum(),
				Value:         proto.String("hi"),
			},
		},
		{
			FragmentType: gauge_messages.Fragment_Text.Enum(),
			Text:         proto.String(" to "),
		},
		{
			FragmentType: gauge_messages.Fragment_Parameter.Enum(),
			Parameter: &gauge_messages.Parameter{
				ParameterType: gauge_messages.Parameter_Dynamic.Enum(),
				Value:         proto.String("gauge"),
			},
		},
		{
			FragmentType: gauge_messages.Fragment_Parameter.Enum(),
			Parameter: &gauge_messages.Parameter{
				ParameterType: gauge_messages.Parameter_Table.Enum(),
				Table: newTableItem([]string{"Word", "Count"}, [][]string{
					[]string{"Gauge", "3"},
					[]string{"Mingle", "2"},
				}).GetTable(),
			},
		},
	},
	StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{
		ExecutionResult: &gauge_messages.ProtoExecutionResult{
			Failed:        proto.Bool(false),
			ExecutionTime: proto.Int64(211316),
		},
	},
}

var protoConcept = &gauge_messages.ProtoConcept{
	ConceptStep: newStepItem(false, []*gauge_messages.Fragment{
		{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Say ")},
		{
			FragmentType: gauge_messages.Fragment_Parameter.Enum(),
			Parameter:    &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Dynamic.Enum(), Value: proto.String("hello")},
		},
		{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String(" to ")},
		{
			FragmentType: gauge_messages.Fragment_Parameter.Enum(),
			Parameter: &gauge_messages.Parameter{
				ParameterType: gauge_messages.Parameter_Table.Enum(),
				Table: newTableItem([]string{"Word", "Count"}, [][]string{
					[]string{"Gauge", "3"},
					[]string{"Mingle", "2"},
				}).GetTable(),
			},
		},
	}).GetStep(),
	Steps: []*gauge_messages.ProtoItem{
		{
			ItemType: gauge_messages.ProtoItem_Concept.Enum(),
			Concept: &gauge_messages.ProtoConcept{
				ConceptStep: newStepItem(false, []*gauge_messages.Fragment{
					{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Tell ")},
					{
						FragmentType: gauge_messages.Fragment_Parameter.Enum(),
						Parameter:    &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Dynamic.Enum(), Value: proto.String("hello")},
					},
				}).GetStep(),
				Steps: []*gauge_messages.ProtoItem{
					newStepItem(false, []*gauge_messages.Fragment{
						{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Say Hi")},
					}),
				},
			},
		},
		newStepItem(false, []*gauge_messages.Fragment{
			{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("Say ")},
			{
				FragmentType: gauge_messages.Fragment_Parameter.Enum(),
				Parameter:    &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Static.Enum(), Value: proto.String("hi")},
			},
			{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String(" to ")},
			{
				FragmentType: gauge_messages.Fragment_Parameter.Enum(),
				Parameter:    &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Dynamic.Enum(), Value: proto.String("gauge")},
			},
			{
				FragmentType: gauge_messages.Fragment_Parameter.Enum(),
				Parameter: &gauge_messages.Parameter{
					ParameterType: gauge_messages.Parameter_Table.Enum(),
					Table: newTableItem([]string{"Word", "Count"}, [][]string{
						[]string{"Gauge", "3"},
						[]string{"Mingle", "2"},
					}).GetTable(),
				},
			},
		}),
	},
	ConceptExecutionResult: &gauge_messages.ProtoStepExecutionResult{
		ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(false), ExecutionTime: proto.Int64(211316)},
	},
}

var protoStepWithSpecialParams = &gauge_messages.ProtoStep{
	Fragments: []*gauge_messages.Fragment{
		{
			FragmentType: gauge_messages.Fragment_Text.Enum(),
			Text:         proto.String("Say "),
		},
		{
			FragmentType: gauge_messages.Fragment_Parameter.Enum(),
			Parameter: &gauge_messages.Parameter{
				Name:          proto.String("foo.txt"),
				ParameterType: gauge_messages.Parameter_Special_String.Enum(),
				Value:         proto.String("hi"),
			},
		},
		{
			FragmentType: gauge_messages.Fragment_Text.Enum(),
			Text:         proto.String(" to "),
		},
		{
			FragmentType: gauge_messages.Fragment_Parameter.Enum(),
			Parameter: &gauge_messages.Parameter{
				ParameterType: gauge_messages.Parameter_Special_Table.Enum(),
				Name:          proto.String("myTable.csv"),
				Table: newTableItem([]string{"Word", "Count"}, [][]string{
					[]string{"Gauge", "3"},
					[]string{"Mingle", "2"},
				}).GetTable(),
			},
		},
	},
	StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{
		ExecutionResult: &gauge_messages.ProtoExecutionResult{
			Failed:        proto.Bool(false),
			ExecutionTime: proto.Int64(211316),
		},
	},
}

func TestTransformOverview(t *testing.T) {
	want := &overview{
		ProjectName: "projName",
		Env:         "ci-java",
		Tags:        "!unimplemented",
		SuccRate:    80.00,
		ExecTime:    "00:01:53",
		Timestamp:   "Jun 3, 2016 at 12:29pm",
		TotalSpecs:  15,
		Failed:      2,
		Passed:      8,
		Skipped:     5,
	}

	got := toOverview(suiteRes1)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestTransformSidebar(t *testing.T) {
	want := &sidebar{
		IsPreHookFailure: false,
		Specs: []*specsMeta{
			{
				SpecName: "specRes1",
				ExecTime: "00:03:31",
				Failed:   false,
				Skipped:  false,
				Tags:     []string{"tag1", "tag2"},
			},
			{
				SpecName: "specRes2",
				ExecTime: "00:03:31",
				Failed:   true,
				Skipped:  false,
				Tags:     []string{"tag1", "tag2", "tag3"},
			},
			{
				SpecName: "specRes3",
				ExecTime: "00:03:31",
				Failed:   false,
				Skipped:  true,
				Tags:     []string{"tag1"},
			},
		},
	}

	got := toSidebar(suiteRes2)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestTransformSpecHeader(t *testing.T) {
	want := &specHeader{
		SpecName: "specRes1",
		ExecTime: "00:03:31",
		FileName: "/tmp/gauge/specs/foobar.spec",
		Tags:     []string{"tag1", "tag2"},
	}

	got := toSpecHeader(specRes1)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToSpec(t *testing.T) {
	want := &spec{
		CommentsBeforeTable: []string{"\n", "This is an executable specification file. This file follows markdown syntax.", "\n", "To execute this specification, run", "\tgauge specs", "\n"},
		Table: &table{
			Headers: []string{"Word", "Count"},
			Rows: []*row{
				{
					Cells: []string{"Gauge", "3"},
					Res:   pass,
				},
				{
					Cells: []string{"Mingle", "2"},
					Res:   pass,
				},
			},
		},
		CommentsAfterTable: []string{"Comment 1", "Comment 2", "Comment 3"},
		Scenarios:          make([]*scenario, 0),
	}

	got := toSpec(specRes1)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToScenario(t *testing.T) {
	want := &scenario{
		Heading:  "Vowel counts in single word",
		ExecTime: "00:01:53",
		Res:      pass,
		Tags:     []string{"foo", "bar"},
		Contexts: []item{
			&step{
				Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Context Step1"}},
				Res:       &result{Status: pass, ExecTime: "00:03:31"},
			},
			&step{
				Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Context Step2"}},
				Res:       &result{Status: fail, ExecTime: "00:03:31"},
			},
		},
		Items: []item{
			&comment{Text: "Comment0"},
			&step{
				Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Step1"}},
				Res:       &result{Status: fail, ExecTime: "00:03:31"},
			},
			&comment{Text: "Comment1"},
			&comment{Text: "Comment2"},
			&step{
				Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Step2"}},
				Res:       &result{Status: pass, ExecTime: "00:03:31"},
			},
			&comment{Text: "Comment3"},
		},
		TearDown: []item{
			&step{
				Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Teardown Step1"}},
				Res:       &result{Status: pass, ExecTime: "00:03:31"},
			},
			&step{
				Fragments: []*fragment{{FragmentKind: textFragmentKind, Text: "Teardown Step2"}},
				Res:       &result{Status: fail, ExecTime: "00:03:31"},
			},
		},
	}

	got := toScenario(scn)
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
					Rows: []*row{
						{Cells: []string{"Gauge", "3"}},
						{Cells: []string{"Mingle", "2"}},
					},
				},
			},
		},
		Res: &result{
			Status:   pass,
			ExecTime: "00:03:31",
		},
	}

	got := toStep(protoStep)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToConcept(t *testing.T) {
	want := &concept{
		CptStep: &step{
			Fragments: []*fragment{
				{FragmentKind: textFragmentKind, Text: "Say "},
				{FragmentKind: dynamicFragmentKind, Text: "hello"},
				{FragmentKind: textFragmentKind, Text: " to "},
				{FragmentKind: tableFragmentKind,
					Table: &table{
						Headers: []string{"Word", "Count"},
						Rows: []*row{
							{Cells: []string{"Gauge", "3"}},
							{Cells: []string{"Mingle", "2"}},
						},
					},
				},
			},
			Res: &result{
				Status:   pass,
				ExecTime: "00:03:31",
			},
		},
		Items: []item{
			&concept{
				CptStep: &step{
					Fragments: []*fragment{
						{FragmentKind: textFragmentKind, Text: "Tell "},
						{FragmentKind: dynamicFragmentKind, Text: "hello"},
					},
					Res: &result{
						Status:   pass,
						ExecTime: "00:03:31",
					},
				},
				Items: []item{
					&step{
						Fragments: []*fragment{
							{FragmentKind: textFragmentKind, Text: "Say Hi"},
						},
						Res: &result{
							Status:   pass,
							ExecTime: "00:03:31",
						},
					},
				},
			},
			&step{
				Fragments: []*fragment{
					{FragmentKind: textFragmentKind, Text: "Say "},
					{FragmentKind: staticFragmentKind, Text: "hi"},
					{FragmentKind: textFragmentKind, Text: " to "},
					{FragmentKind: dynamicFragmentKind, Text: "gauge"},
					{FragmentKind: tableFragmentKind,
						Table: &table{
							Headers: []string{"Word", "Count"},
							Rows: []*row{
								{Cells: []string{"Gauge", "3"}},
								{Cells: []string{"Mingle", "2"}},
							},
						},
					},
				},
				Res: &result{
					Status:   pass,
					ExecTime: "00:03:31",
				},
			},
		},
	}

	got := toConcept(protoConcept)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestToStepWithSpecialParams(t *testing.T) {
	want := &step{
		Fragments: []*fragment{
			{FragmentKind: textFragmentKind, Text: "Say "},
			{FragmentKind: specialStringFragmentKind, Name: "foo.txt", Text: "hi"},
			{FragmentKind: textFragmentKind, Text: " to "},
			{FragmentKind: specialTableFragmentKind,
				Name: "myTable.csv",
				Table: &table{
					Headers: []string{"Word", "Count"},
					Rows: []*row{
						{Cells: []string{"Gauge", "3"}},
						{Cells: []string{"Mingle", "2"}},
					},
				},
			},
		},
		Res: &result{
			Status:   pass,
			ExecTime: "00:03:31",
		},
	}

	got := toStep(protoStepWithSpecialParams)
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
