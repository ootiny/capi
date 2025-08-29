package utils

import (
	"testing"
)

func TestFirstError(t *testing.T) {
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
