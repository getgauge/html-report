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

var specRes1 = &gauge_messages.ProtoSpecResult{
	Failed:        proto.Bool(false),
	Skipped:       proto.Bool(false),
	ExecutionTime: proto.Int64(211316),
	ProtoSpec: &gauge_messages.ProtoSpec{
		SpecHeading: proto.String("specRes1"),
		Tags:        []string{"tag1", "tag2"},
		FileName:    proto.String("/tmp/gauge/specs/foobar.spec"),
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

var suiteRes2 = &gauge_messages.ProtoSuiteResult{
	SpecResults: []*gauge_messages.ProtoSpecResult{specRes1, specRes2, specRes3},
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
			&specsMeta{
				SpecName: "specRes1",
				ExecTime: "00:03:31",
				Failed:   false,
				Skipped:  false,
				Tags:     []string{"tag1", "tag2"},
			},
			&specsMeta{
				SpecName: "specRes2",
				ExecTime: "00:03:31",
				Failed:   true,
				Skipped:  false,
				Tags:     []string{"tag1", "tag2", "tag3"},
			},
			&specsMeta{
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
