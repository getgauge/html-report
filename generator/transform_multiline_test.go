/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package generator

import (
	"testing"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
)

// Multiline string test helpers
func newMultilineStringParam(name, value string) *gm.Parameter {
	return &gm.Parameter{
		Name:          name,
		ParameterType: gm.Parameter_Special_String,
		Value:         value,
	}
}

func newMultilineParamFragment(name, value string) *gm.Fragment {
	return &gm.Fragment{
		FragmentType: gm.Fragment_Parameter,
		Parameter:    newMultilineStringParam(name, value),
	}
}

// Test data for multiline steps
var protoStepWithMultilineString = &gm.ProtoStep{
	Fragments: []*gm.Fragment{
		newTextFragment("Step with JSON "),
		newMultilineParamFragment("file:data.json", "{\n  \"name\": \"Gauge\",\n  \"type\": \"Testing\"\n}"),
	},
	StepExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{
			Failed:        false,
			ExecutionTime: 211316,
		},
	},
}

var protoStepWithMixedMultiline = &gm.ProtoStep{
	Fragments: []*gm.Fragment{
		newTextFragment("Step with "),
		newMultilineParamFragment("file:config.txt", "key=value\nanother=setting"),
		newTextFragment(" and "),
		newMultilineParamFragment("file:simple.txt", "simple value"),
	},
	StepExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{
			Failed:        false,
			ExecutionTime: 211316,
		},
	},
}

var protoStepWithEmptyMultiline = &gm.ProtoStep{
	Fragments: []*gm.Fragment{
		newTextFragment("Step with empty content "),
		newMultilineParamFragment("file:empty.txt", ""),
	},
	StepExecutionResult: &gm.ProtoStepExecutionResult{
		ExecutionResult: &gm.ProtoExecutionResult{
			Failed:        false,
			ExecutionTime: 211316,
		},
	},
}

// Test cases for multiline string support
func TestToFragmentsWithMultilineStringParameter(t *testing.T) {
	multilineContent := "line1\nline2\nline3"
	protoFragments := []*gm.Fragment{
		newMultilineParamFragment("file:multiline.txt", multilineContent),
	}

	want := []*fragment{
		{
			FragmentKind: multilineFragmentKind,
			Text:         multilineContent,
		},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "Multiline string parameter", want, got)
}

func TestToFragmentsWithSingleLineStringParameter(t *testing.T) {
	singleLineContent := "single line value"
	protoFragments := []*gm.Fragment{
		newMultilineParamFragment("file:sample.txt", singleLineContent),
	}

	want := []*fragment{
		{
			FragmentKind: specialStringFragmentKind,
			Name:         "file:sample.txt",
			Text:         singleLineContent,
			FileName:     "sample.txt",
		},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "Single line string parameter", want, got)
}

func TestToFragmentsWithMixedTypesIncludingMultiline(t *testing.T) {
	multilineContent := "key=value\nanother=setting"
	protoFragments := []*gm.Fragment{
		newTextFragment("Step with "),
		newMultilineParamFragment("file:config.txt", multilineContent),
		newTextFragment(" and "),
		newMultilineParamFragment("file:simple.txt", "simple value"),
	}

	want := []*fragment{
		{FragmentKind: textFragmentKind, Text: "Step with "},
		{FragmentKind: multilineFragmentKind, Text: multilineContent},
		{FragmentKind: textFragmentKind, Text: " and "},
		{FragmentKind: specialStringFragmentKind, Name: "file:simple.txt", Text: "simple value", FileName: "simple.txt"},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "Mixed types including multiline", want, got)
}

func TestToStepWithMultilineStringParameter(t *testing.T) {
	want := &step{
		Fragments: []*fragment{
			{FragmentKind: textFragmentKind, Text: "Step with JSON "},
			{FragmentKind: multilineFragmentKind, Text: "{\n  \"name\": \"Gauge\",\n  \"type\": \"Testing\"\n}"},
		},
		Result: &result{
			Status:        pass,
			ExecutionTime: "00:03:31",
		},
	}

	got := toStep(protoStepWithMultilineString)
	checkEqual(t, "Step with multiline string parameter", want, got)
}

func TestToStepWithMixedMultilineAndSingleLine(t *testing.T) {
	want := &step{
		Fragments: []*fragment{
			{FragmentKind: textFragmentKind, Text: "Step with "},
			{FragmentKind: multilineFragmentKind, Text: "key=value\nanother=setting"},
			{FragmentKind: textFragmentKind, Text: " and "},
			{FragmentKind: specialStringFragmentKind, Name: "file:simple.txt", Text: "simple value", FileName: "simple.txt"},
		},
		Result: &result{
			Status:        pass,
			ExecutionTime: "00:03:31",
		},
	}

	got := toStep(protoStepWithMixedMultiline)
	checkEqual(t, "Step with mixed multiline and single line", want, got)
}

func TestToStepWithEmptyMultilineString(t *testing.T) {
	want := &step{
		Fragments: []*fragment{
			{FragmentKind: textFragmentKind, Text: "Step with empty content "},
			{FragmentKind: specialStringFragmentKind, Name: "file:empty.txt", Text: "", FileName: "empty.txt"},
		},
		Result: &result{
			Status:        pass,
			ExecutionTime: "00:03:31",
		},
	}

	got := toStep(protoStepWithEmptyMultiline)
	checkEqual(t, "Step with empty multiline string", want, got)
}

func TestToFragmentsWithMultilineStringContainingOnlyNewline(t *testing.T) {
	protoFragments := []*gm.Fragment{
		newMultilineParamFragment("file:newline.txt", "\n"),
	}

	want := []*fragment{
		{
			FragmentKind: multilineFragmentKind,
			Text:         "\n",
		},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "Multiline string with only newline", want, got)
}

func TestToFragmentsWithComplexMultilineContent(t *testing.T) {
	complexContent := `# Configuration File
database:
  host: localhost
  port: 5432
  name: gauge_db

logging:
  level: info
  file: /var/log/gauge.log`
	
	protoFragments := []*gm.Fragment{
		newTextFragment("Load configuration "),
		newMultilineParamFragment("file:config.yaml", complexContent),
	}

	want := []*fragment{
		{FragmentKind: textFragmentKind, Text: "Load configuration "},
		{FragmentKind: multilineFragmentKind, Text: complexContent},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "Complex multiline content", want, got)
}

func TestToFragmentsWithXMLMultilineContent(t *testing.T) {
	xmlContent := `<user>
  <name>John Doe</name>
  <email>john@example.com</email>
  <roles>
    <role>admin</role>
    <role>user</role>
  </roles>
</user>`
	
	protoFragments := []*gm.Fragment{
		newTextFragment("Create user with XML "),
		newMultilineParamFragment("file:user.xml", xmlContent),
	}

	want := []*fragment{
		{FragmentKind: textFragmentKind, Text: "Create user with XML "},
		{FragmentKind: multilineFragmentKind, Text: xmlContent},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "XML multiline content", want, got)
}

func TestToFragmentsWithCodeSnippetMultiline(t *testing.T) {
	codeContent := `func main() {
    fmt.Println("Hello, World!")
    for i := 0; i < 10; i++ {
        fmt.Printf("Count: %d\n", i)
    }
}`
	
	protoFragments := []*gm.Fragment{
		newTextFragment("Execute code: "),
		newMultilineParamFragment("file:main.go", codeContent),
	}

	want := []*fragment{
		{FragmentKind: textFragmentKind, Text: "Execute code: "},
		{FragmentKind: multilineFragmentKind, Text: codeContent},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "Code snippet multiline content", want, got)
}

// Test edge case: string with Windows-style line endings
func TestToFragmentsWithWindowsLineEndings(t *testing.T) {
	windowsContent := "line1\r\nline2\r\nline3"
	protoFragments := []*gm.Fragment{
		newMultilineParamFragment("file:windows.txt", windowsContent),
	}

	want := []*fragment{
		{
			FragmentKind: multilineFragmentKind,
			Text:         windowsContent,
		},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "Windows line endings", want, got)
}

// Test edge case: string with mixed line endings
func TestToFragmentsWithMixedLineEndings(t *testing.T) {
	mixedContent := "line1\nline2\r\nline3"
	protoFragments := []*gm.Fragment{
		newMultilineParamFragment("file:mixed.txt", mixedContent),
	}

	want := []*fragment{
		{
			FragmentKind: multilineFragmentKind,
			Text:         mixedContent,
		},
	}

	got := toFragments(protoFragments)
	checkEqual(t, "Mixed line endings", want, got)
}