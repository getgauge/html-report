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

func formatTime(ms int64) string {
	return time.Unix(0, ms*int64(time.Millisecond)).UTC().Format(execTimeFormat)
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
