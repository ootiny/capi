package utils

import (
	"strings"
	"testing"
)

func Test_SetTraceError(t *testing.T) {
	t.Run("set true", func(t *testing.T) {
		SetTraceError(false)
		SetTraceError(true)

		assert := NewAssert(t)
		assert(IsTraceError()).Equals(true)
	})

	t.Run("set false", func(t *testing.T) {
		SetTraceError(true)
		SetTraceError(false)

		assert := NewAssert(t)
		assert(IsTraceError()).Equals(false)
	})
}

func Test_FirstError(t *testing.T) {
	t.Run("all errors is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(FirstError(nil, nil, nil)).Equals(nil)
	})

	t.Run("first error is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(FirstError(nil, Errorf("error2"), nil)).Equals(Errorf("error2"))
	})

	t.Run("first error is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(FirstError(Errorf("error1"), nil, Errorf("error3"))).Equals(Errorf("error1"))
	})
}

func Test_Error_Error(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert((*Error)(nil).Error()).Equals("")
	})

	t.Run("error is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(Errorf("error message").Error()).Equals("error message")
	})
}

func Test_Error_SetCode(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert((*Error)(nil).SetCode(2)).Equals(nil)
	})

	t.Run("error is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(Errorf("error message").SetCode(2)).Equals(Errorf("error message").SetCode(2))
	})
}

func Test_Error_AddHeader(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert((*Error)(nil).AddHeader("header")).Equals(nil)
	})

	t.Run("error is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(Errorf("error message").AddHeader("header %s", "test")).Equals(Errorf("header test: error message"))
	})
}

func Test_Errorf(t *testing.T) {
	t.Run("trace on", func(t *testing.T) {
		SetTraceError(true)
		assert := NewAssert(t)
		err := Errorf("error %s", "message")
		assert(err.message).Equals("error message")
		assert(err.code).Equals(ErrCodeGeneral)
		assert(len(err.traces)).Equals(1)
		assert(strings.Contains(err.traces[0], "builder/utils/error_test.go")).IsTrue()
	})

	t.Run("trace off", func(t *testing.T) {
		SetTraceError(false)
		assert := NewAssert(t)
		err := Errorf("error %s", "message")
		assert(err.message).Equals("error message")
		assert(err.code).Equals(ErrCodeGeneral)
		assert(len(err.traces)).Equals(0)
	})
}
