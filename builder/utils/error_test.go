package utils

import (
	"fmt"
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

func Test_Error_Code(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert((*Error)(nil).Code()).Equals(0)
	})

	t.Run("error is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(Errorf("error message").SetCode(12).Code()).Equals(ErrCodeGeneral)
	})
}

func Test_Error_SetCode(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert((*Error)(nil).SetCode(2)).Equals(nil)
	})

	t.Run("error is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		err := Errorf("error message").SetCode(2)
		assert(err.code).Equals(2)
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

func Test_WrapError(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(WrapError(nil)).Equals(nil)
	})

	t.Run("error is not nil, trace on", func(t *testing.T) {
		SetTraceError(true)
		assert := NewAssert(t)
		err1 := WrapError(Errorf("error message"))
		assert(err1.message).Equals("error message")
		assert(err1.code).Equals(ErrCodeGeneral)
		assert(len(err1.traces)).Equals(2)
		assert(strings.Contains(err1.traces[1], "builder/utils/error_test.go")).IsTrue()

		err2 := WrapError(fmt.Errorf("error2"))
		assert(err2.message).Equals("error2")
		assert(err2.code).Equals(ErrCodeGeneral)
		assert(len(err2.traces)).Equals(1)
		assert(strings.Contains(err2.traces[0], "builder/utils/error_test.go")).IsTrue()
	})

	t.Run("error is not nil, trace off", func(t *testing.T) {
		SetTraceError(false)
		assert := NewAssert(t)
		err1 := WrapError(Errorf("error message"))
		assert(err1.message).Equals("error message")
		assert(err1.code).Equals(ErrCodeGeneral)
		assert(len(err1.traces)).Equals(0)

		err2 := WrapError(fmt.Errorf("error2"))
		assert(err2.message).Equals("error2")
		assert(err2.code).Equals(ErrCodeGeneral)
		assert(len(err2.traces)).Equals(0)
	})
}

func Test_DebugError(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		assert := NewAssert(t)
		assert(DebugError(nil)).Equals("")
	})

	t.Run("error is not nil, trace on", func(t *testing.T) {
		SetTraceError(true)
		assert := NewAssert(t)
		debugMessage := DebugError(Errorf("error %d", 19))

		assert(strings.Contains(debugMessage, "Message: error 19")).IsTrue()
		assert(strings.Contains(debugMessage, fmt.Sprintf("Code: %d", ErrCodeGeneral))).IsTrue()
		assert(strings.Contains(debugMessage, "builder/utils/error_test.go")).IsTrue()
	})

	t.Run("error is not nil, trace off", func(t *testing.T) {
		SetTraceError(false)
		assert := NewAssert(t)
		debugMessage := DebugError(Errorf("error message"))
		assert(debugMessage).Equals("Message: error message\nCode: 1\n")
	})
}
