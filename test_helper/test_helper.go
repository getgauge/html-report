/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package testHelper

import (
	"fmt"
	"html"
	"os"
	"regexp"
	"testing"

	htmldiff "github.com/documize/html-diff"
)

var re = regexp.MustCompile("[\\s]*[\n\t][\\s]*")

func RemoveNewline(s string) string {
	return re.ReplaceAllLiteralString(s, "")
}

func AssertEqual(expected, actual, testName string, t *testing.T) {
	if expected != actual {
		diffHTML := compare(expected, actual)
		tmpFile, err := os.CreateTemp("", "")
		if err != nil {
			t.Errorf("Unable to dump to tmp file. Raw content:\n%s\n", diffHTML)
		}
		fileName := fmt.Sprintf("%s.html", tmpFile.Name())
		err = os.WriteFile(fileName, []byte(diffHTML), 0644)
		if err != nil {
			t.Errorf("Unable to write file %s. Error: %s", fileName, err.Error())
		}

		_ = tmpFile.Close()
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

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func SetEnvOrFail(t *testing.T, key, value string) {
	if err := os.Setenv(key, value); err != nil {
		t.Logf("Failed to set env %s: %v", key, err)
	}
}

func UnsetEnvOrFail(t *testing.T, key string) {
	if err := os.Unsetenv(key); err != nil {
		t.Logf("Failed to unset env %s: %v", key, err)
	}
}

func RemoveOrFail(t *testing.T, key string) {
	if err := os.Remove(key); err != nil {
		t.Logf("Failed to set env %s: %v", key, err)
	}
}
