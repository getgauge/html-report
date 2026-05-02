/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package testHelper

import (
	"os"
	"testing"
)

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
		t.Logf("Failed to remove %s: %v", key, err)
	}
}
