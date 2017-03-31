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
	"fmt"
	"html"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"

	htmldiff "github.com/documize/html-diff"
	gm "github.com/getgauge/html-report/gauge_messages"
)

var scenario1 = &gm.ProtoScenario{
	ScenarioHeading: "Vowel counts in single word",
	ExecutionStatus: gm.ExecutionStatus_PASSED,
	Tags:            []string{"foo", "bar"},
	ExecutionTime:   113163,
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
			ExecutionResult: &gm.ProtoExecutionResult{Failed: false, ExecutionTime: 211316},
		}),
		newConceptItem("Outer Concept", []*gm.ProtoItem{
			newStepItem(false, false, []*gm.Fragment{newTextFragment("Outer Concept Step 1")}),
			newConceptItem("Inner Concept", []*gm.ProtoItem{
				newStepItem(false, false, []*gm.Fragment{newTextFragment("Inner Concept Step 1")}),
				newStepItem(false, false, []*gm.Fragment{newTextFragment("Inner Concept Step 2")}),
			}, &gm.ProtoStepExecutionResult{
				ExecutionResult: &gm.ProtoExecutionResult{Failed: false, ExecutionTime: 211316},
			}),
			newStepItem(false, false, []*gm.Fragment{newTextFragment("Outer Concept Step 2")}),
		}, &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{Failed: false, ExecutionTime: 211316},
		}),
	},
	TearDownSteps: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Teardown Step1")}),
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Teardown Step2")}),
	},
}

var scenarioWithConceptFailure = &gm.ProtoScenario{
	ScenarioHeading: "Vowel counts in single word",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	Tags:            []string{"foo", "bar"},
	ExecutionTime:   113163,
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Step1")}),
		newConceptItem("Outer Concept", []*gm.ProtoItem{
			newStepItem(false, false, []*gm.Fragment{newTextFragment("Outer Concept Step 1")}),
			newConceptItem("Inner Concept", []*gm.ProtoItem{
				newStepItem(false, false, []*gm.Fragment{newTextFragment("Inner Concept Step 1")}),
				failedStep,
				newStepItem(false, true, []*gm.Fragment{newTextFragment("Inner Concept Step 3")}),
			}, &gm.ProtoStepExecutionResult{ExecutionResult: &gm.ProtoExecutionResult{Failed: true, ExecutionTime: 113163}}),
			newStepItem(false, true, []*gm.Fragment{newTextFragment("Outer Concept Step 2")}),
		}, &gm.ProtoStepExecutionResult{ExecutionResult: &gm.ProtoExecutionResult{Failed: true, ExecutionTime: 113163}}),
	},
}

var scenario2 = &gm.ProtoScenario{
	ScenarioHeading: "Vowel counts in multiple words",
	ExecutionStatus: gm.ExecutionStatus_PASSED,
	ExecutionTime:   113163,
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
	ScenarioHeading: "skipped scenario",
	ExecutionStatus: gm.ExecutionStatus_SKIPPED,
	ExecutionTime:   0,
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
	ScenarioHeading: "Scenario Heading",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	ExecutionTime:   113163,
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, false, []*gm.Fragment{newTextFragment("Some step")}),
	},
	PostHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: "java.lang.RuntimeException",
		StackTrace:   newStackTrace(),
		ScreenShot:   []byte(newScreenshot()),
	},
}

var scenarioWithBeforeHookFail = &gm.ProtoScenario{
	ScenarioHeading: "Scenario Heading",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	ExecutionTime:   113163,
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, true, []*gm.Fragment{newTextFragment("Some step")}),
	},
	PreHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: "java.lang.RuntimeException",
		StackTrace:   newStackTrace(),
		ScreenShot:   []byte(newScreenshot()),
	},
}

var scenarioWithBeforeAndAfterHookFail = &gm.ProtoScenario{
	ScenarioHeading: "Scenario Heading",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	ExecutionTime:   113163,
	ScenarioItems: []*gm.ProtoItem{
		newStepItem(false, true, []*gm.Fragment{newTextFragment("Some step")}),
	},
	PreHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: "java.lang.RuntimeException",
		StackTrace:   newStackTrace(),
		ScreenShot:   []byte(newScreenshot()),
	},
	PostHookFailure: &gm.ProtoHookFailure{
		ErrorMessage: "java.lang.RuntimeException",
		StackTrace:   newStackTrace(),
		ScreenShot:   []byte(newScreenshot()),
	},
}

var stepWithBeforeHookFail = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step,
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{
				Failed:        true,
				ExecutionTime: 211316,
			},
			PreHookFailure: &gm.ProtoHookFailure{
				ErrorMessage: "java.lang.RuntimeException",
				StackTrace:   newStackTrace(),
				ScreenShot:   []byte(newScreenshot()),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This is a failing step")},
	},
}

var stepWithAfterHookFail = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step,
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{
				Failed:        true,
				ExecutionTime: 211316,
			},
			PostHookFailure: &gm.ProtoHookFailure{
				ErrorMessage: "java.lang.RuntimeException",
				StackTrace:   newStackTrace(),
				ScreenShot:   []byte(newScreenshot()),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This is a failing step")},
	},
}

var stepWithBeforeAndAfterHookFail = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step,
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{
				Failed:        true,
				ExecutionTime: 211316,
			},
			PreHookFailure: &gm.ProtoHookFailure{
				ErrorMessage: "java.lang.RuntimeException",
				StackTrace:   newStackTrace(),
				ScreenShot:   []byte(newScreenshot()),
			},
			PostHookFailure: &gm.ProtoHookFailure{
				ErrorMessage: "java.lang.RuntimeException",
				StackTrace:   newStackTrace(),
				ScreenShot:   []byte(newScreenshot()),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This is a failing step")},
	},
}

var failedStep = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step,
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			ExecutionResult: &gm.ProtoExecutionResult{
				Failed:        true,
				ExecutionTime: 211316,
				ErrorMessage:  "java.lang.RuntimeException",
				StackTrace:    newStackTrace(),
				ScreenShot:    []byte(newScreenshot()),
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This is a failing step")},
	},
}

var stepNotExecuted = &gm.ProtoItem{
	ItemType: gm.ProtoItem_Step,
	Step: &gm.ProtoStep{
		StepExecutionResult: &gm.ProtoStepExecutionResult{
			Skipped: true,
			ExecutionResult: &gm.ProtoExecutionResult{
				ExecutionTime: 0,
			},
		},
		Fragments: []*gm.Fragment{newTextFragment("This step is skipped because previous one failed")},
	},
}

var scenarioWithBeforeStepFail = &gm.ProtoScenario{
	ScenarioHeading: "Scenario Heading",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	ExecutionTime:   113163,
	ScenarioItems:   []*gm.ProtoItem{stepWithBeforeHookFail, stepNotExecuted},
}

var scenarioWithAfterStepFail = &gm.ProtoScenario{
	ScenarioHeading: "Scenario Heading",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	ExecutionTime:   113163,
	ScenarioItems:   []*gm.ProtoItem{stepWithAfterHookFail, stepNotExecuted},
}

var scenarioWithBeforeAndAfterStepFail = &gm.ProtoScenario{
	ScenarioHeading: "Scenario Heading",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	ExecutionTime:   113163,
	ScenarioItems:   []*gm.ProtoItem{stepWithBeforeAndAfterHookFail, stepNotExecuted},
}

var scenarioWithStepFail = &gm.ProtoScenario{
	ScenarioHeading: "Scenario Heading",
	ExecutionStatus: gm.ExecutionStatus_FAILED,
	ExecutionTime:   113163,
	ScenarioItems:   []*gm.ProtoItem{newStepItem(false, false, []*gm.Fragment{newTextFragment("passing step")}), failedStep, stepNotExecuted},
}

var passSpecRes1 = &gm.ProtoSpecResult{
	Failed:        false,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Passing Specification 1",
		Tags:        []string{"tag1", "tag2"},
		FileName:    "passing_specification_1.spec",
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
	Failed:        false,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		FileName:    "passing_specification_2.spec",
		SpecHeading: "Passing Specification 2",
		Tags:        []string{},
	},
}

var passSpecRes3 = &gm.ProtoSpecResult{
	Failed:        false,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		FileName:    "passing_specification_3.spec",
		SpecHeading: "Passing Specification 3",
		Tags:        []string{"foo"},
	},
}

var failSpecResWithAfterScenarioFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithAfterHookFail),
		},
	},
}

var failSpecResWithBeforeScenarioFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithBeforeHookFail),
		},
	},
}

var failSpecResWithBeforeAndAfterScenarioFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithBeforeAndAfterHookFail),
		},
	},
}

var specResWithMultipleScenarios = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithStepFail),
			newScenarioItem(scenario2),
		},
	},
}

var failSpecResWithBeforeStepFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithBeforeStepFail),
		},
	},
}

var failSpecResWithAfterStepFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithAfterStepFail),
		},
	},
}

var failSpecResWithBeforeAndAfterStepFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithBeforeAndAfterStepFail),
		},
	},
}

var failSpecResWithStepFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithStepFail),
		},
	},
}

var failSpecResWithConceptFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification",
		Tags:        []string{},
		FileName:    "failing_specification.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(scenarioWithConceptFailure),
		},
	},
}

var skippedSpecRes = &gm.ProtoSpecResult{
	Failed:        false,
	Skipped:       true,
	ExecutionTime: 0,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Skipped Specification",
		Tags:        []string{},
		FileName:    "skipped_specification.spec",
		Items: []*gm.ProtoItem{
			newScenarioItem(skippedScenario),
		},
	},
}

var failSpecResWithAfterSpecFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
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
			ErrorMessage: "java.lang.RuntimeException",
			StackTrace:   newStackTrace(),
			ScreenShot:   []byte(newScreenshot()),
		},
	},
}

var failSpecResWithBeforeSpecFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
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
			ErrorMessage: "java.lang.RuntimeException",
			StackTrace:   newStackTrace(),
			ScreenShot:   []byte(newScreenshot()),
		},
	},
}

var failSpecResWithBeforeAfterSpecFailure = &gm.ProtoSpecResult{
	Failed:        true,
	Skipped:       false,
	ExecutionTime: 211316,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Failing Specification 1",
		Tags:        []string{},
		FileName:    "failing_specification_1.spec",
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
			ErrorMessage: "java.lang.RuntimeException",
			StackTrace:   newStackTrace(),
			ScreenShot:   []byte(newScreenshot()),
		},
		PostHookFailure: &gm.ProtoHookFailure{
			ErrorMessage: "java.lang.RuntimeException",
			StackTrace:   newStackTrace(),
			ScreenShot:   []byte(newScreenshot()),
		},
	},
}

var skipSpecRes1 = &gm.ProtoSpecResult{
	Failed:        false,
	Skipped:       true,
	ExecutionTime: 0,
	ProtoSpec: &gm.ProtoSpec{
		SpecHeading: "Skipped Specification 1",
		FileName:    "skipped_specification_1.spec",
		Tags:        []string{"bar"},
	},
}

var errorSpecResults = &gm.ProtoSpecResult{
	Failed:        true,
	ExecutionTime: 0,
	ProtoSpec: &gm.ProtoSpec{
		FileName:    "error_specification.spec",
		SpecHeading: "Error Spec",
		Tags:        []string{"bar"},
	},
	Errors: []*gm.Error{
		{Type: gm.Error_PARSE_ERROR, Message: "message"},
	},
}

var suiteRes = newSuiteResult(false, 1, 1, 60, nil, nil, passSpecRes1, passSpecRes2, passSpecRes3, failSpecResWithAfterScenarioFailure, skipSpecRes1)
var suiteResWithAfterSuiteFailure = newSuiteResult(true, 1, 1, 60, nil, newProtoHookFailure(), passSpecRes1, passSpecRes2,
	passSpecRes3, failSpecResWithAfterScenarioFailure, skipSpecRes1)
var suiteResWithBeforeScenarioFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithBeforeScenarioFailure)
var suiteResWithAfterScenarioFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithAfterScenarioFailure)
var suiteResWithBeforeAndAfterScenarioFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithBeforeAndAfterScenarioFailure)
var suiteResWithMultipleScenarios = newSuiteResult(true, 1, 0, 0, nil, nil, specResWithMultipleScenarios)
var suiteResWithBeforeStepFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithBeforeStepFailure)
var suiteResWithAfterStepFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithAfterStepFailure)
var suiteResWithBeforeAndAfterStepFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithBeforeAndAfterStepFailure)
var suiteResWithStepFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithStepFailure)
var suiteResWithBeforeSpecFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithBeforeSpecFailure)
var suiteResWithAfterSpecFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithAfterSpecFailure)
var suiteResWithBeforeAfterSpecFailure = newSuiteResult(true, 1, 0, 0, nil, nil, failSpecResWithBeforeAfterSpecFailure)
var suiteResWithConceptFailure = newSuiteResult(true, 1, 0, 60, nil, nil, failSpecResWithConceptFailure)
var suiteResWithSkippedSpec = newSuiteResult(false, 0, 1, 0, nil, nil, skippedSpecRes)
var suiteResWithAllPass = newSuiteResult(false, 0, 0, 100, nil, nil, passSpecRes2)
var suiteResWithSpecError = newSuiteResult(true, 1, 0, 0.0, nil, nil, errorSpecResults)

func newProtoHookFailure() *gm.ProtoHookFailure {
	return &gm.ProtoHookFailure{
		ErrorMessage: "java.lang.RuntimeException",
		StackTrace:   newStackTrace(),
		ScreenShot:   []byte(newScreenshot()),
	}
}

func newSuiteResult(failed bool, failCount, skipCount int32, succRate float32, preHook, postHook *gm.ProtoHookFailure, specRes ...*gm.ProtoSpecResult) *SuiteResult {
	return ToSuiteResult(newProtoSuiteRes(failed, failCount, skipCount, succRate, preHook, postHook, specRes...))
}

func newProtoSuiteRes(failed bool, failCount, skipCount int32, succRate float32, preHook, postHook *gm.ProtoHookFailure, specRes ...*gm.ProtoSpecResult) *gm.ProtoSuiteResult {
	return &gm.ProtoSuiteResult{
		SpecResults:       specRes,
		Failed:            failed,
		SpecsFailedCount:  failCount,
		ExecutionTime:     122609,
		SuccessRate:       succRate,
		Environment:       "default",
		Tags:              "",
		ProjectName:       "Gauge Project",
		Timestamp:         "Jul 13, 2016 at 11:49am",
		SpecsSkippedCount: skipCount,
		PostHookFailure:   postHook,
		PreHookFailure:    preHook,
	}
}

type HTMLGenerationTest struct {
	name         string
	res          *SuiteResult
	expectedFile string
}

var HTMLGenerationTests = []*HTMLGenerationTest{
	{"happy path", suiteRes, "pass.html"},
	{"after suite failure", suiteResWithAfterSuiteFailure, "after_suite_fail.html"},
	{"before spec failure", suiteResWithBeforeSpecFailure, "before_spec_fail.html"},
	{"after spec failure", suiteResWithAfterSpecFailure, "after_spec_fail.html"},
	{"skipped specification", suiteResWithSkippedSpec, "skipped_spec.html"},
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
	{"spec error", suiteResWithSpecError, "spec_err.html"},
}

func TestHTMLGeneration(t *testing.T) {
	for _, test := range HTMLGenerationTests {
		content, err := ioutil.ReadFile(filepath.Join("_testdata", "integration", test.expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}

		buf := new(bytes.Buffer)
		var wg sync.WaitGroup
		wg.Add(1)

		generateSpecPage(test.res, test.res.SpecResults[0], buf, &wg)
		wg.Wait()

		want := removeNewline(string(content))
		got := removeNewline(buf.String())
		assertEqual(want, got, test.name, t)
	}
}

func TestIndexPageGeneration(t *testing.T) {
	content, err := ioutil.ReadFile(filepath.Join("_testdata", "integration", "pass_index.html"))
	if err != nil {
		t.Errorf("Error reading expected HTML file: %s", err.Error())
	}

	buf := new(bytes.Buffer)
	var wg sync.WaitGroup
	wg.Add(1)

	generateIndexPage(suiteResWithAllPass, buf, &wg)
	wg.Wait()

	want := removeNewline(string(content))
	got := removeNewline(buf.String())

	assertEqual(want, got, "index", t)
}

func assertEqual(expected, actual, testName string, t *testing.T) {
	if expected != actual {
		diffHTML := compare(expected, actual)
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Errorf("Unable to dump to tmp file. Raw content:\n%s\n", diffHTML)
		}
		fileName := fmt.Sprintf("%s.html", tmpFile.Name())
		ioutil.WriteFile(fileName, []byte(diffHTML), 0644)
		tmpFile.Close()
		t.Errorf("%s -  View Diff Output : %s\n", testName, fileName)
	}
}
func compare(a, b string) string {
	var cfg = &htmldiff.Config{
		InsertedSpan: []htmldiff.Attribute{{Key: "style", Val: "background-color: palegreen;"}},
		DeletedSpan:  []htmldiff.Attribute{{Key: "style", Val: "background-color: lightpink;"}},
		ReplacedSpan: []htmldiff.Attribute{{Key: "style", Val: "background-color: lightskyblue;"}},
		CleanTags:    []string{""},
	}

	res, _ := cfg.HTMLdiff([]string{html.EscapeString(a), html.EscapeString(b)})
	return res[0]
}
