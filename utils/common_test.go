package utils

import (
	"strings"
	"testing"
)

func Test_convertOrdinalToString(t *testing.T) {
	t.Run("number is 0", func(t *testing.T) {
		assert := NewAssert(t)
		assert(convertOrdinalToString(0)).Equals("")
	})

	t.Run("number is 1", func(t *testing.T) {
		assert := NewAssert(t)
		assert(convertOrdinalToString(1)).Equals("1st")
	})

	t.Run("number is 2", func(t *testing.T) {
		assert := NewAssert(t)
		assert(convertOrdinalToString(2)).Equals("2nd")
	})

	t.Run("number is 3", func(t *testing.T) {
		assert := NewAssert(t)
		assert(convertOrdinalToString(3)).Equals("3rd")
	})

	t.Run("number is large than 3", func(t *testing.T) {
		assert := NewAssert(t)
		assert(convertOrdinalToString(4)).Equals("4th")
		assert(convertOrdinalToString(100)).Equals("100th")
	})
}

func Test_getFileLine(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		assert(strings.Contains(getFileLine(0), "builder/utils/common_test.go")).IsTrue()
	})

	t.Run("skip is overflow", func(t *testing.T) {
		assert := NewAssert(t)
		assert(getFileLine(1000)).Equals("")
	})
}

func Test_addPrefixPerLine(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		assert := NewAssert(t)
		assert(addPrefixPerLine("", "prefix")).Equals("prefix")
	})

	t.Run("one line string", func(t *testing.T) {
		assert := NewAssert(t)
		assert(addPrefixPerLine("line1", "prefix")).Equals("prefixline1")
	})

	t.Run("multi line string", func(t *testing.T) {
		assert := NewAssert(t)
		assert(addPrefixPerLine("line1\nline2", "prefix")).Equals("prefixline1\nprefixline2")
	})
}
