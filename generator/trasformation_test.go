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

var suiteRes = &gauge_messages.ProtoSuiteResult{
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

var o = &overview{
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

var transformTests = []transformTest{
	{"transforms to overview", suiteRes, o},
}

func TestTransform(t *testing.T) {
	for _, test := range transformTests {
		got := toOverview(test.input)
		want := test.output.(*overview)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: \n want:\n%q\ngot:\n%q\n", test.name, want, got)
		}
	}
}
