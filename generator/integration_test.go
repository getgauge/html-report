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
	"path/filepath"
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
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Context Step1")}),
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Context Step2")}),
	},
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Step1")}),
		newCommentItem("Comment1"),
		newStepItem(false, false, []*gm.Fragment{
			newTextFragment("Say "),
			newParamFragment(newStaticParam("hi")),
			newTextFragment(" to "),
			newParamFragment(newDynamicParam("gauge")),
		}),
		newCommentItem("Comment2"),
		newConceptItem("Concept Heading", []*gm.ProtoItem{
			newStepItem(false, false, []*gm.Fragment{newTextFragment("Concept Step1")}),
			newStepItem(false, false, []*gm.Fragment{newTextFragment("Concept Step2")}),
		}, &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{Failed: proto.Bool(false), ExecutionTime: proto.Int64(211316)},
		}),
		newConceptItem("Outer Concept", []*gm.ProtoItem{
			newStepItem(false, false, []*gm.Fragment{newTextFragment("Outer Concept Step 1")}),
			newConceptItem("Inner Concept", []*gm.ProtoItem{
				newStepItem(false, false, []*gm.Fragment{newTextFragment("Inner Concept Step 1")}),
				newStepItem(false, false, []*gm.Fragment{newTextFragment("Inner Concept Step 2")}),
			}, &gm.ProtoStepExecutionResult{
				ExecutionResult: &gm.ProtoExecutionResult{Failed: proto.Bool(false), ExecutionTime: proto.Int64(211316)},
			}),
			newStepItem(false, false, []*gm.Fragment{newTextFragment("Outer Concept Step 2")}),
		}, &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{Failed: proto.Bool(false), ExecutionTime: proto.Int64(211316)},
		}),
	},
	TearDownSteps: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Teardown Step1")}),
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Teardown Step2")}),
	},
}

var scenarioWithConceptFailure = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Vowel counts in single word"),
	Failed:          proto.Bool(true),
	Skipped:         proto.Bool(false),
	Tags:            []string{"foo", "bar"},
	ExecutionTime:   proto.Int64(113163),
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Step1")}),
		newConceptItem("Outer Concept", []*gm.ProtoItem{
			newStepItem(false, false, []*gm.Fragment{newTextFragment("Outer Concept Step 1")}),
			newConceptItem("Inner Concept", []*gm.ProtoItem{
				newStepItem(false, false, []*gm.Fragment{newTextFragment("Inner Concept Step 1")}),
				failedStep,
				newStepItem(false, true, []*gm.Fragment{newTextFragment("Inner Concept Step 3")}),
			}, &gm.ProtoStepExecutionResult{ExecutionResult: &gm.ProtoExecutionResult{Failed: proto.Bool(true), ExecutionTime: proto.Int64(113163)}}),
			newStepItem(false, true, []*gm.Fragment{newTextFragment("Outer Concept Step 2")}),
		}, &gm.ProtoStepExecutionResult{ExecutionResult: &gm.ProtoExecutionResult{Failed: proto.Bool(true), ExecutionTime: proto.Int64(113163)}}),
	},
}

var scenario2 = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Vowel counts in multiple words"),
	Failed:          proto.Bool(false),
	Skipped:         proto.Bool(false),
	ExecutionTime:   proto.Int64(113163),
	Contexts: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Context Step1")}),
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Context Step2")}),
	},
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{
			newTextFragment("Almost all words have vowels"),
			newParamFragment(newTableParam([]string{"Word", "Count"}, [][]string{
				[]string{"Gauge", "3"},
				[]string{"Mingle", "2"},
			})),
		}),
	},
	TearDownSteps: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Teardown Step1")}),
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Teardown Step2")}),
	},
}

var skippedScenario = &gm.ProtoScenario{
	ScenarioHeading: proto.String("skipped scenario"),
	Failed:          proto.Bool(false),
	Skipped:         proto.Bool(true),
	ExecutionTime:   proto.Int64(0),
	Contexts: []*gm.ProtoItem{
		newStepItem(false, true, []*gm.Fragment{newTextFragment("Context Step")}),
	},
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, true, []*gm.Fragment{
			newTextFragment("skipped step"),
		}),
	},
}

var scenarioWithAfterHookFail = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Scenario Heading"),
	Failed:          proto.Bool(true),
	Skipped:         proto.Bool(false),
	ExecutionTime:   proto.Int64(113163),
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Some step")}),
	},
	PostHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: proto.String("java.lang.RuntimeException"),
		StackTrace:   proto.String(newStackTrace()),
		ScreenShot:   []byte(newScreenshot()),
	},
}

var scenarioWithBeforeHookFail = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Scenario Heading"),
	Failed:          proto.Bool(true),
	Skipped:         proto.Bool(false),
	ExecutionTime:   proto.Int64(113163),
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, true, []*gm.Fragment{newTextFragment("Some step")}),
	},
	PreHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: proto.String("java.lang.RuntimeException"),
		StackTrace:   proto.String(newStackTrace()),
		ScreenShot:   []byte(newScreenshot()),
	},
}

var scenarioWithBeforeAndAfterHookFail = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Scenario Heading"),
	Failed:          proto.Bool(true),
	Skipped:         proto.Bool(false),
	ExecutionTime:   proto.Int64(113163),
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, true, []*gm.Fragment{newTextFragment("Some step")}),
	},
	PreHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: proto.String("java.lang.RuntimeException"),
		StackTrace:   proto.String(newStackTrace()),
		ScreenShot:   []byte(newScreenshot()),
	},
	PostHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: proto.String("java.lang.RuntimeException"),
		StackTrace:   proto.String(newStackTrace()),
		ScreenShot:   []byte(newScreenshot()),
	},
}

var stepWithBeforeHookFail = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step.Enum(),
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{
				Failed:        proto.Bool(true),
				ExecutionTime: proto.Int64(211316),
			},
			PreHookFailure: &gm.ProtoHookFailure{
				ErrorMessage: proto.String("java.lang.RuntimeException"),
				StackTrace:   proto.String(newStackTrace()),
				ScreenShot:   []byte(newScreenshot()),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This is a failing step")},
	},
}

var stepWithAfterHookFail = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step.Enum(),
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{
				Failed:        proto.Bool(true),
				ExecutionTime: proto.Int64(211316),
			},
			PostHookFailure: &gm.ProtoHookFailure{
				ErrorMessage: proto.String("java.lang.RuntimeException"),
				StackTrace:   proto.String(newStackTrace()),
				ScreenShot:   []byte(newScreenshot()),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This is a failing step")},
	},
}

var stepWithBeforeAndAfterHookFail = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step.Enum(),
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{
				Failed:        proto.Bool(true),
				ExecutionTime: proto.Int64(211316),
			},
			PreHookFailure: &gm.ProtoHookFailure{
				ErrorMessage: proto.String("java.lang.RuntimeException"),
				StackTrace:   proto.String(newStackTrace()),
				ScreenShot:   []byte(newScreenshot()),
			},
			PostHookFailure: &gm.ProtoHookFailure{
				ErrorMessage: proto.String("java.lang.RuntimeException"),
				StackTrace:   proto.String(newStackTrace()),
				ScreenShot:   []byte(newScreenshot()),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This is a failing step")},
	},
}

var failedStep = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step.Enum(),
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{
				Failed:        proto.Bool(true),
				ExecutionTime: proto.Int64(211316),
				ErrorMessage:  proto.String("java.lang.RuntimeException"),
				StackTrace:    proto.String(newStackTrace()),
				ScreenShot:    []byte(newScreenshot()),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This is a failing step")},
	},
}

var stepNotExecuted = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step.Enum(),
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			Skipped: proto.Bool(true),
			ExecutionResult: &gm.ProtoExecutionResult{
				ExecutionTime: proto.Int64(0),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This step is skipped because previous one failed")},
	},
}

var scenarioWithBeforeStepFail = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Scenario Heading"),
	Failed:          proto.Bool(true),
	Skipped:         proto.Bool(false),
	ExecutionTime:   proto.Int64(113163),
	ScenarioItems:   []*gm.ProtoItem{stepWithBeforeHookFail, stepNotExecuted},
}

var scenarioWithAfterStepFail = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Scenario Heading"),
	Failed:          proto.Bool(true),
	Skipped:         proto.Bool(false),
	ExecutionTime:   proto.Int64(113163),
	ScenarioItems:   []*gm.ProtoItem{stepWithAfterHookFail, stepNotExecuted},
}

var scenarioWithBeforeAndAfterStepFail = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Scenario Heading"),
	Failed:          proto.Bool(true),
	Skipped:         proto.Bool(false),
	ExecutionTime:   proto.Int64(113163),
	ScenarioItems:   []*gm.ProtoItem{stepWithBeforeAndAfterHookFail, stepNotExecuted},
}

var scenarioWithStepFail = &gm.ProtoScenario{
	ScenarioHeading: proto.String("Scenario Heading"),
	Failed:          proto.Bool(true),
	Skipped:         proto.Bool(false),
	ExecutionTime:   proto.Int64(113163),
	ScenarioItems:   []*gm.ProtoItem{newStepItem(false, false, []*gm.Fragment{newTextFragment("passing step")}), failedStep, stepNotExecuted},
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
			newScenarioItem(scenario2),
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

var failSpecResWithAfterScenarioFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithAfterHookFail),
		},
	},
}

var failSpecResWithBeforeScenarioFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithBeforeHookFail),
		},
	},
}

var failSpecResWithBeforeAndAfterScenarioFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithBeforeAndAfterHookFail),
		},
	},
}

var specResWithMultipleScenarios = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithStepFail),
			newScenarioItem(scenario2),
		},
	},
}

var failSpecResWithBeforeStepFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithBeforeStepFail),
		},
	},
}

var failSpecResWithAfterStepFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithAfterStepFail),
		},
	},
}

var failSpecResWithBeforeAndAfterStepFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithBeforeAndAfterStepFail),
		},
	},
}

var failSpecResWithStepFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithStepFail),
		},
	},
}

var failSpecResWithConceptFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithConceptFailure),
		},
	},
}

var skippedSpecRes = &gm.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(true),
	ExecutionTime: proto.Int64(0),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Skipped Specification"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newScenarioItem(skippedScenario),
		},
	},
}

var failSpecResWithAfterSpecFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
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
			newScenarioItem(scenario2),
		},
		PostHookFailure: &gm.ProtoHookFailure{
			ErrorMessage: proto.String("java.lang.RuntimeException"),
			StackTrace:   proto.String(newStackTrace()),
			ScreenShot:   []byte(newScreenshot()),
		},
	},
}

var failSpecResWithBeforeSpecFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newCommentItem("\n"),
			newCommentItem("This is an executable specification file. This file follows markdown syntax."),
			newCommentItem("\n"),
			newCommentItem("To execute this specification, run"),
			newCommentItem("\tgauge specs"),
			newCommentItem("\n"),
			newCommentItem("Comment 1"),
			newCommentItem("Comment 2"),
			newCommentItem("Comment 3"),
			newScenarioItem(scenario2),
		},
		PreHookFailure: &gm.ProtoHookFailure{
			ErrorMessage: proto.String("java.lang.RuntimeException"),
			StackTrace:   proto.String(newStackTrace()),
			ScreenShot:   []byte(newScreenshot()),
		},
	},
}

var failSpecResWithBeforeAfterSpecFailure = &gm.ProtoSpecResult{
	Failed:        proto.Bool(true),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: proto.String("Failing Specification 1"),
		Tags:        []string{},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
		Items: []*gm.ProtoItem{
			newCommentItem("\n"),
			newCommentItem("This is an executable specification file. This file follows markdown syntax."),
			newCommentItem("\n"),
			newCommentItem("To execute this specification, run"),
			newCommentItem("\tgauge specs"),
			newCommentItem("\n"),
			newCommentItem("Comment 1"),
			newCommentItem("Comment 2"),
			newCommentItem("Comment 3"),
			newScenarioItem(scenario2),
		},
		PreHookFailure: &gm.ProtoHookFailure{
			ErrorMessage: proto.String("java.lang.RuntimeException"),
			StackTrace:   proto.String(newStackTrace()),
			ScreenShot:   []byte(newScreenshot()),
		},
		PostHookFailure: &gm.ProtoHookFailure{
			ErrorMessage: proto.String("java.lang.RuntimeException"),
			StackTrace:   proto.String(newStackTrace()),
			ScreenShot:   []byte(newScreenshot()),
		},
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



var suiteResWithBeforeSuiteFailure = &gm.ProtoSuiteResult{
	SpecResults:       []*gm.ProtoSpecResult{},
	Failed:            proto.Bool(true),
	SpecsFailedCount:  proto.Int32(0),
	ExecutionTime:     proto.Int64(122609),
	SuccessRate:       proto.Float32(0),
	Environment:       proto.String("default"),
	Tags:              proto.String(""),
	ProjectName:       proto.String("Gauge Project"),
	Timestamp:         proto.String("Jul 13, 2016 at 11:49am"),
	SpecsSkippedCount: proto.Int32(0),
	PreHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: proto.String("java.lang.RuntimeException"),
		StackTrace:   proto.String(newStackTrace()),
		ScreenShot:   []byte(newScreenshot()),
	},
}


var suiteResWithBeforeAfterSuiteFailure = &gm.ProtoSuiteResult{
	Failed:            proto.Bool(true),
	SpecsFailedCount:  proto.Int32(0),
	ExecutionTime:     proto.Int64(122609),
	SuccessRate:       proto.Float32(0),
	Environment:       proto.String("default"),
	Tags:              proto.String(""),
	ProjectName:       proto.String("Gauge Project"),
	Timestamp:         proto.String("Jul 13, 2016 at 11:49am"),
	SpecsSkippedCount: proto.Int32(0),
	PreHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: proto.String("java.lang.RuntimeException"),
		StackTrace:   proto.String(newStackTrace()),
		ScreenShot:   []byte(newScreenshot()),
	},
	PostHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: proto.String("java.lang.RuntimeException"),
		StackTrace:   proto.String(newStackTrace()),
		ScreenShot:   []byte(newScreenshot()),
	},
}

var suiteRes = newProtoSuiteRes(false, 1, 1, 60, nil, nil,passSpecRes1, passSpecRes2, passSpecRes3, failSpecResWithAfterScenarioFailure, skipSpecRes1)
var suiteResWithAfterSuiteFailure = newProtoSuiteRes(true,1,1,60,nil,newProtoHookFailure(), passSpecRes1, passSpecRes2,
	passSpecRes3, failSpecResWithAfterScenarioFailure, skipSpecRes1)
var suiteResWithBeforeScenarioFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithBeforeScenarioFailure)
var suiteResWithAfterScenarioFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithAfterScenarioFailure)
var suiteResWithBeforeAndAfterScenarioFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithBeforeAndAfterScenarioFailure)
var suiteResWithMultipleScenarios = newProtoSuiteRes(true,1,0,0,nil,nil,specResWithMultipleScenarios)
var suiteResWithBeforeStepFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithBeforeStepFailure)
var suiteResWithAfterStepFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithAfterStepFailure)
var suiteResWithBeforeAndAfterStepFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithBeforeAndAfterStepFailure)
var suiteResWithStepFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithStepFailure)
var suiteResWithBeforeSpecFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithBeforeSpecFailure)
var suiteResWithAfterSpecFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithAfterSpecFailure)
var suiteResWithBeforeAfterSpecFailure = newProtoSuiteRes(true,1,0,0,nil,nil,failSpecResWithBeforeAfterSpecFailure)
var suiteResWithConceptFailure = newProtoSuiteRes(true,1,0,60,nil,nil,failSpecResWithConceptFailure)
var suiteResWithSkippedSpec = newProtoSuiteRes(false,0,1,0,nil,nil,skippedSpecRes)
var suiteResWithAllPass = newProtoSuiteRes(false, 0, 0, 100,nil,nil, passSpecRes2)

func newProtoHookFailure() *gm.ProtoHookFailure {
	return &gm.ProtoHookFailure{
		ErrorMessage: proto.String("java.lang.RuntimeException"),
		StackTrace:   proto.String(newStackTrace()),
		ScreenShot:   []byte(newScreenshot()),
	}
}

func newProtoSuiteRes(failed bool, failCount, skipCount int32, succRate float32,preHook,postHook *gm.ProtoHookFailure, specRes ...*gm.ProtoSpecResult) *gm.ProtoSuiteResult {
	return &gm.ProtoSuiteResult{
		SpecResults: specRes,
		Failed: proto.Bool(failed),
		SpecsFailedCount: proto.Int32(failCount),
		ExecutionTime:     proto.Int64(122609),
		SuccessRate:       proto.Float32(succRate),
		Environment:       proto.String("default"),
		Tags:              proto.String(""),
		ProjectName:       proto.String("Gauge Project"),
		Timestamp:         proto.String("Jul 13, 2016 at 11:49am"),
		SpecsSkippedCount: proto.Int32(skipCount),
		PostHookFailure:postHook,
		PreHookFailure:preHook,
	}
}

type HTMLGenerationTest struct {
	name         string
	res          *gm.ProtoSuiteResult
	expectedFile string
}

var HTMLGenerationTests = []*HTMLGenerationTest{
	{"happy path", suiteRes, "pass.html"},
	{"after suite failure", suiteResWithAfterSuiteFailure, "after_suite_fail.html"},
	{"before spec failure", suiteResWithBeforeSpecFailure, "before_spec_fail.html"},
	{"after spec failure", suiteResWithAfterSpecFailure, "after_spec_fail.html"},
	{"skipped specification",suiteResWithSkippedSpec,"skipped_spec.html"},
	{"both before and after spec failure", suiteResWithBeforeAfterSpecFailure, "before_after_spec_fail.html"},
	{"before scenario failure", suiteResWithBeforeScenarioFailure, "before_scenario_fail.html"},
	{"after scenario failure", suiteResWithAfterScenarioFailure, "after_scenario_fail.html"},
	{"both before and after scenario failure", suiteResWithBeforeAndAfterScenarioFailure, "before_after_scenario_fail.html"},
	{"multiple scenarios", suiteResWithMultipleScenarios, "multiple_scenarios.html"},
	{"before step failure", suiteResWithBeforeStepFailure, "before_step_fail.html"},
	{"after step failure", suiteResWithAfterStepFailure, "after_step_fail.html"},
	{"both before after step failure", suiteResWithBeforeAndAfterStepFailure, "before_after_step_fail.html"},
	{"step failure", suiteResWithStepFailure, "step_fail.html"},
	{"concept failure", suiteResWithConceptFailure, "concept_fail.html"},
}

func TestHTMLGeneration(t *testing.T) {
	for _, test := range HTMLGenerationTests {
		content, err := ioutil.ReadFile(filepath.Join("_testdata", "integration", test.expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}

		buf := new(bytes.Buffer)
		generateSpecPage(test.res, buf)

		want := removeNewline(string(content))
		got := removeNewline(buf.String())

		if got != want {
			t.Errorf("%s:\nwant:\n%q\ngot:\n%q\n", test.name, want, got)
		}

	}
}

func TestIndexPageGeneration(t *testing.T){
	content, err := ioutil.ReadFile(filepath.Join("_testdata","integration", "pass_index.html"))
	if err != nil {
		t.Errorf("Error reading expected HTML file: %s", err.Error())
	}

	buf := new(bytes.Buffer)
	generateIndexPage(suiteResWithAllPass, buf)

	want := removeNewline(string(content))
	got := removeNewline(buf.String())

	if got != want {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}
