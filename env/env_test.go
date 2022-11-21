package env

import "testing"

func TestMaxRecvMsgSize(t *testing.T) {
	t.Run("empty value should return default", func(t *testing.T) {
		t.Setenv(gaugeMaxMessageSize, "")
		v := GetMaxMessageSize()
		if v != 1024 {
			t.Errorf("Expected 1024, got %d", v)
		}
	})

	t.Run("non-numeric should return default", func(t *testing.T) {
		t.Setenv(gaugeMaxMessageSize, "abcd")
		v := GetMaxMessageSize()
		if v != 1024 {
			t.Errorf("Expected 1024, got %d", v)
		}
	})

	t.Run("numeric should return set value", func(t *testing.T) {
		t.Setenv(gaugeMaxMessageSize, "2048")
		v := GetMaxMessageSize()
		if v != 2048 {
			t.Errorf("Expected 2048, got %d", v)
		}
	})
}
