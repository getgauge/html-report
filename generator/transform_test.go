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

	gm "github.com/getgauge/html-report/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type transformTest struct {
	name   string
	input  *gm.ProtoSuiteResult
	output interface{}
}

func newCommentItem(str string) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Comment.Enum(),
		Comment: &gm.ProtoComment{
			Text: proto.String(str),
		},
	}
}

func newScenarioItem(scn *gm.ProtoScenario) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Scenario.Enum(),
		Scenario: scn,
	}
}

func newTableItem(headers []string, rows [][]string) *gm.ProtoItem {
	r := make([]*gm.ProtoTableRow, len(rows))
	for i, row := range rows {
		r[i] = &gm.ProtoTableRow{
			Cells: row,
		}
	}
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Table.Enum(),
		Table: &gm.ProtoTable{
			Headers: &gm.ProtoTableRow{
				Cells: headers,
			},
			Rows: r,
		},
	}
}

func newStepItem(failed bool, frags []*gm.Fragment) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Step.Enum(),
		Step: &gm.ProtoStep{
			StepExecutionResult: &gm.ProtoStepExecutionResult{
				ExecutionResult: &gm.ProtoExecutionResult{
					Failed:        proto.Bool(failed),
					ExecutionTime: proto.Int64(211316),
				},
			},
			Fragments: frags,
		},
	}
}

func newDynamicParam(val string) *gm.Parameter {
	return &gm.Parameter{
		ParameterType: gm.Parameter_Dynamic.Enum(),
		Value:         proto.String(val),
	}
}

func newStaticParam(val string) *gm.Parameter {
	return &gm.Parameter{
		ParameterType: gm.Parameter_Static.Enum(),
		Value:         proto.String(val),
	}
}

func newTextFragment(val string) *gm.Fragment {
	return &gm.Fragment{
		FragmentType: gm.Fragment_Text.Enum(),
		Text:         proto.String(val),
	}
}

func newParamFragment(p *gm.Parameter) *gm.Fragment {
	return &gm.Fragment{
		FragmentType: gm.Fragment_Parameter.Enum(),
		Parameter:    p,
	}
}

func newConceptItem(heading string, steps []*gm.ProtoItem) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Concept.Enum(),
		Concept: &gm.ProtoConcept{
			ConceptStep: newStepItem(false, []*gm.Fragment{newTextFragment(heading)}).GetStep(),
			Steps:       steps,
		},
	}
}

var specRes1 = &gm.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading:   proto.String("specRes1"),
		Tags:          []string{"tag1", "tag2"},
		FileName:      proto.String("/tmp/gauge/specs/foobar.spec"),
		IsTableDriven: proto.Bool(false),
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
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("specRes2"),
		Tags:        []string{"tag1", "tag2", "tag3"},
	},
}

var specRes3 = &gm.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(true),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("specRes3"),
		Tags:        []string{"tag1"},
	},
}

var suiteRes1 = &gm.ProtoSuiteResult{
	ProjectName:       proto.String("projName"),
	Environment:       proto.String("ci-java"),
	Tags:              proto.String("!unimplemented"),
	SuccessRate:       proto.Float32(80.00),
	ExecutionTime:     proto.Int64(113163),
	Timestamp:         proto.String("Jun 3, 2016 at 12:29pm"),
	SpecResults:       make([]*gm.ProtoSpecResult, 15),
	SpecsFailedCount:  proto.Int32(2),
	SpecsSkippedCount: proto.Int32(5),
}

var scn = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Vowel counts in single word"),
	Failed:          proto.Bool(false),
	Skipped:         proto.Bool(false),
	Tags:            []string{"foo", "bar"},
	ExecutionTime:   proto.Int64(113163),
	Contexts: []*gm.ProtoItem{
		newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Context Step1")}}),
		newStepItem(true, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Context Step2")}}),
	},
	ScenarioItems: []*gm.ProtoItem{
		newCommentItem("Comment0"),
		newStepItem(true, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Step1")}}),
		newCommentItem("Comment1"),
		newCommentItem("Comment2"),
		newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Step2")}}),
		newCommentItem("Comment3"),
	},
	TearDownSteps: []*gm.ProtoItem{
		newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Teardown Step1")}}),
		newStepItem(true, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Teardown Step2")}}),
	},
}

var suiteRes2 = &gm.ProtoSuiteResult{
	SpecResults: []*gm.ProtoSpecResult{specRes1, specRes2, specRes3},
}

var protoStep = &gm.ProtoStep{
	Fragments: []*gm.Fragment{
		{
			FragmentType: gm.Fragment_Text.Enum(),
			Text:         proto.String("Say "),
		},
		{
			FragmentType: gm.Fragment_Parameter.Enum(),
			Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Static.Enum(),
				Value:         proto.String("hi"),
			},
		},
		{
			FragmentType: gm.Fragment_Text.Enum(),
			Text:         proto.String(" to "),
		},
		{
			FragmentType: gm.Fragment_Parameter.Enum(),
			Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Dynamic.Enum(),
				Value:         proto.String("gauge"),
			},
		},
		{
			FragmentType: gm.Fragment_Parameter.Enum(),
			Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Table.Enum(),
				Table: newTableItem([]string{"Word", "Count"}, [][]string{
					[]string{"Gauge", "3"},
					[]string{"Mingle", "2"},
				}).GetTable(),
			},
		},
	},
	StepExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{
			Failed:        proto.Bool(false),
			ExecutionTime: proto.Int64(211316),
		},
	},
}

var protoConcept = &gm.ProtoConcept{
	ConceptStep: newStepItem(false, []*gm.Fragment{
		{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Say ")},
		{
			FragmentType: gm.Fragment_Parameter.Enum(),
			Parameter:    &gm.Parameter{ParameterType: gm.Parameter_Dynamic.Enum(), Value: proto.String("hello")},
		},
		{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String(" to ")},
		{
			FragmentType: gm.Fragment_Parameter.Enum(),
			Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Table.Enum(),
				Table: newTableItem([]string{"Word", "Count"}, [][]string{
					[]string{"Gauge", "3"},
					[]string{"Mingle", "2"},
				}).GetTable(),
			},
		},
	}).GetStep(),
	Steps: []*gm.ProtoItem{
		{
			ItemType: gm.ProtoItem_Concept.Enum(),
			Concept: &gm.ProtoConcept{
				ConceptStep: newStepItem(false, []*gm.Fragment{
					{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Tell ")},
					{
						FragmentType: gm.Fragment_Parameter.Enum(),
						Parameter:    &gm.Parameter{ParameterType: gm.Parameter_Dynamic.Enum(), Value: proto.String("hello")},
					},
				}).GetStep(),
				Steps: []*gm.ProtoItem{
					newStepItem(false, []*gm.Fragment{
						{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Say Hi")},
					}),
				},
			},
		},
		newStepItem(false, []*gm.Fragment{
			{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Say ")},
			{
				FragmentType: gm.Fragment_Parameter.Enum(),
				Parameter:    &gm.Parameter{ParameterType: gm.Parameter_Static.Enum(), Value: proto.String("hi")},
			},
			{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String(" to ")},
			{
				FragmentType: gm.Fragment_Parameter.Enum(),
				Parameter:    &gm.Parameter{ParameterType: gm.Parameter_Dynamic.Enum(), Value: proto.String("gauge")},
			},
			{
				FragmentType: gm.Fragment_Parameter.Enum(),
				Parameter: &gm.Parameter{
					ParameterType: gm.Parameter_Table.Enum(),
					Table: newTableItem([]string{"Word", "Count"}, [][]string{
						[]string{"Gauge", "3"},
						[]string{"Mingle", "2"},
					}).GetTable(),
				},
			},
		}),
	},
	ConceptExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{Failed: proto.Bool(false), ExecutionTime: proto.Int64(211316)},
	},
}

var protoStepWithSpecialParams = &gm.ProtoStep{
	Fragments: []*gm.Fragment{
		{
			FragmentType: gm.Fragment_Text.Enum(),
			Text:         proto.String("Say "),
		},
		{
			FragmentType: gm.Fragment_Parameter.Enum(),
			Parameter: &gm.Parameter{
				Name:          proto.String("foo.txt"),
				ParameterType: gm.Parameter_Special_String.Enum(),
				Value:         proto.String("hi"),
			},
		},
		{
			FragmentType: gm.Fragment_Text.Enum(),
			Text:         proto.String(" to "),
		},
		{
			FragmentType: gm.Fragment_Parameter.Enum(),
			Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Special_Table.Enum(),
				Name:          proto.String("myTable.csv"),
				Table: newTableItem([]string{"Word", "Count"}, [][]string{
					[]string{"Gauge", "3"},
					[]string{"Mingle", "2"},
				}).GetTable(),
			},
		},
	},
	StepExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{
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
