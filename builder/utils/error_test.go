package utils

import (
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	SetTraceError(true)
	defer SetTraceError(false)

}

func TestWrapError(t *testing.T) {
	SetTraceError(true)
	defer SetTraceError(false)

	err := WrapError(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = WrapError(fmt.Errorf("test error"))
	if err == nil {
		t.Fatal("expect error")
	}

	err = WrapError(err)
	if err == nil {
		t.Fatal("expect error")
	}

	t.Log(DebugError(err))
}
