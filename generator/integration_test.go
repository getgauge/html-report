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
	"io/ioutil"
	"testing"

	gm "github.com/getgauge/html-report/gauge_messages"
	"github.com/golang/protobuf/proto"
)

var scenario1 = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Vowel counts in single word"),
	Failed:          proto.Bool(false),
	Skipped:         proto.Bool(false),
	Tags:            []string{"foo", "bar"},
	ExecutionTime:   proto.Int64(113163),
	Contexts: []*gm.ProtoItem{
		newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Context Step1")}}),
		newStepItem(false, []*gm.Fragment{
			{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Say ")},
			{FragmentType: gm.Fragment_Parameter.Enum(), Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Static.Enum(),
				Value:         proto.String("hi"),
			}},
			{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String(" to ")},
			{FragmentType: gm.Fragment_Parameter.Enum(), Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Dynamic.Enum(),
				Value:         proto.String("gauge"),
			}},
		}),
	},
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Step1")}}),
		newCommentItem("Comment1"),
		newStepItem(false, []*gm.Fragment{
			{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Say ")},
			{FragmentType: gm.Fragment_Parameter.Enum(), Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Static.Enum(),
				Value:         proto.String("hi"),
			}},
			{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String(" to ")},
			{FragmentType: gm.Fragment_Parameter.Enum(), Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Dynamic.Enum(),
				Value:         proto.String("gauge"),
			}},
		}),
		newCommentItem("Comment2"),
		&gm.ProtoItem{
			ItemType: gm.ProtoItem_Concept.Enum(),
			Concept: &gm.ProtoConcept{
				ConceptStep: newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Concept Heading")}}).GetStep(),
				Steps: []*gm.ProtoItem{
					newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Concept Step1")}}),
					newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Concept Step2")}}),
				},
			},
		},
	},
	TearDownSteps: []*gm.ProtoItem{
		newStepItem(false, []*gm.Fragment{{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Teardown Step1")}}),
		newStepItem(false, []*gm.Fragment{
			{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String("Say ")},
			{FragmentType: gm.Fragment_Parameter.Enum(), Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Static.Enum(),
				Value:         proto.String("hi"),
			}},
			{FragmentType: gm.Fragment_Text.Enum(), Text: proto.String(" to ")},
			{FragmentType: gm.Fragment_Parameter.Enum(), Parameter: &gm.Parameter{
				ParameterType: gm.Parameter_Dynamic.Enum(),
				Value:         proto.String("gauge"),
			}},
		}),
	},
}

var passSpecRes1 = &gm.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Passing Specification 1"),
		Tags:        []string{"tag1", "tag2"},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
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
			newScenarioItem(scenario1),
		},
	},
}

var passSpecRes2 = &gm.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Passing Specification 2"),
		Tags:        []string{},
	},
}

var passSpecRes3 = &gm.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Passing Specification 3"),
		Tags:        []string{"foo"},
	},
}

var failSpecRes1 = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(0),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
	},
}

var skipSpecRes1 = &gm.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(true),
	ExecutionTime: proto.Int64(0),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Skipped Specification 1"),
		Tags:        []string{"bar"},
	},
}

var suiteRes = &gm.ProtoSuiteResult{
	SpecResults:       []*gm.ProtoSpecResult{passSpecRes1, passSpecRes2, passSpecRes3, failSpecRes1, skipSpecRes1},
	Failed:            proto.Bool(false),
	SpecsFailedCount:  proto.Int32(1),
	ExecutionTime:     proto.Int64(122609),
	SuccessRate:       proto.Float32(60),
	Environment:       proto.String("default"),
	Tags:              proto.String(""),
	ProjectName:       proto.String("Gauge Project"),
	Timestamp:         proto.String("Jul 13, 2016 at 11:49am"),
	SpecsSkippedCount: proto.Int32(1),
}

func TestHTMLGeneration(t *testing.T) {
	content, err := ioutil.ReadFile("_testdata/expected.html")
	if err != nil {
		t.Errorf("Error reading expected HTML file: %s", err.Error())
	}

	buf := new(bytes.Buffer)
	generate(suiteRes, buf)

	want := removeNewline(string(content))
	got := removeNewline(buf.String())

	if got != want {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}
