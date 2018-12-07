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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getgauge/html-report/gauge_messages"

	"github.com/getgauge/html-report/env"
)

func NewSearchIndex() *SearchIndex {
	var i SearchIndex
	i.Tags = make(map[string][]string)
	i.Specs = make(map[string][]string)
	return &i
}

func (i *SearchIndex) hasValueForTag(tag string, spec string) bool {
	for _, s := range i.Tags[tag] {
		if s == spec {
			return true
		}
	}
	return false
}

func (i *SearchIndex) hasSpec(specHeading string, specFileName string) bool {
	for _, s := range i.Specs[specHeading] {
		if s == specFileName {
			return true
		}
	}
	return false
}

func (i *SearchIndex) AddRawSpec(r *gauge_messages.ProtoSpec) {
	specFileName := toHTMLFileName(r.FileName, projectRoot)
	for _, t := range r.Tags {
		if !i.hasValueForTag(t, specFileName) {
			i.Tags[t] = append(i.Tags[t], specFileName)
		}
	}
	specHeading := r.SpecHeading
	if !i.hasSpec(specHeading, specFileName) {
		i.Specs[specHeading] = append(i.Specs[specHeading], specFileName)
	}
}

func (i *SearchIndex) AddRawItem(r *gauge_messages.ProtoItem) {
	specFileName := toHTMLFileName(r.FileName, projectRoot)
	if r.ItemType == gauge_messages.ProtoItem_Scenario {
		for _, t := range r.Scenario.Tags {
			if !i.hasValueForTag(t, specFileName) {
				i.Tags[t] = append(i.Tags[t], specFileName)
			}
		}
	}
}

func (i *SearchIndex) add(r *spec) {
	specFileName := toHTMLFileName(r.FileName, projectRoot)
	for _, t := range r.Tags {
		if !i.hasValueForTag(t, specFileName) {
			i.Tags[t] = append(i.Tags[t], specFileName)
		}
	}
	for _, s := range r.Scenarios {
		for _, t := range s.Tags {
			if !i.hasValueForTag(t, specFileName) {
				i.Tags[t] = append(i.Tags[t], specFileName)
			}
		}
	}
	specHeading := r.SpecHeading
	if !i.hasSpec(specHeading, specFileName) {
		i.Specs[specHeading] = append(i.Specs[specHeading], specFileName)
	}
}

func (i *SearchIndex) Write(dir string) error {
	env.CreateDirectory(filepath.Join(dir, "js"))
	f, err := os.Create(filepath.Join(dir, "js", "search_index.js"))
	if err != nil {
		return err
	}
	defer f.Close()
	s, err := json.Marshal(i)
	if err != nil {
		return err
	}
	f.WriteString(fmt.Sprintf("var index = %s;", s))
	return nil
}

func generateSearchIndex(suiteRes *SuiteResult, reportsDir string) error {
	index := NewSearchIndex()
	for _, r := range suiteRes.SpecResults {
		index.add(r)
	}
	return index.Write(reportsDir)
}
