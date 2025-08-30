package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"unsafe"
)

func captureStdout(fn func()) string {
	oldStdout := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	func() {
		defer func() {
			_ = recover()
		}()
		fn()
	}()

	outCH := make(chan string)
	// copy the output in a separate goroutine so print can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outCH <- buf.String()
	}()

	os.Stdout = oldStdout
	_ = w.Close()
	ret := <-outCH
	_ = r.Close()
	return ret
}

type fakeTesting struct {
	onFail func()
}

func (p *fakeTesting) Fail() {
	if p.onFail != nil {
		p.onFail()
	}
}

func testFailHelper(fn func(_ func(_ ...any) *Assert)) (bool, string) {
	retCH := make(chan bool, 1)

	retValue := captureStdout(func() {
		fn(func(args ...any) *Assert {
			return &Assert{
				t: &fakeTesting{
					onFail: func() {
						retCH <- true
					},
				},
				args: args,
			}
		})
	})

	select {
	case <-retCH:
		return true, retValue
	default:
		return false, retValue
	}
}

func TestNewAssert(t *testing.T) {
	t.Run("t is nil", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewAssert(nil)
		assert(o(true)).Equals(&Assert{t: nil, args: []any{true}})
	})

	t.Run("args is nil", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewAssert(t)
		assert(o()).Equals(&Assert{t: t, args: nil})
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewAssert(t)
		assert(o(true, 1)).Equals(&Assert{t: t, args: []any{true, 1}})
	})
}

func TestRpcAssert_Fail(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o().Fail("error"); source = getFileLine(0) }()
		})).Equals(true, "\terror\n\t"+source+"\n")
	})
}

func TestRpcAssert_Equals(t *testing.T) {
	t.Run("arguments is empty", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o().Equals(); source = getFileLine(0) }()
		})).Equals(true, "\targuments is empty\n\t"+source+"\n")
	})

	t.Run("arguments is empty", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(1).Equals(1, 2); source = getFileLine(0) }()
		})).Equals(true, "\targuments length not match\n\t"+source+"\n")
	})

	t.Run("arguments does not equal", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(1).Equals(2); source = getFileLine(0) }()
		})).Equals(true, fmt.Sprintf(
			"\t1st argument does not equal\n\t"+
				"want:\n\t%s\n\tgot:\n\t%s\n\t%s\n",
			"int(2)",
			"int(1)",
			source,
		))
	})

	t.Run("arguments type does not equal", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(1).Equals(int64(1)); source = getFileLine(0) }()
		})).Equals(true, fmt.Sprintf(
			"\t1st argument does not equal\n\t"+
				"want:\n\t%s\n\tgot:\n\t%s\n\t%s\n",
			"int64(1)",
			"int(1)",
			source,
		))

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			v1 := map[int]any{3: "OK", 4: []byte(nil)}
			v2 := map[int]any{3: "OK", 4: nil}
			func() { o(v1).Equals(v2); source = getFileLine(0) }()
		})).Equals(true, fmt.Sprintf(
			"\t1st argument does not equal\n\t"+
				"want:\n\t%s\n\tgot:\n\t%s\n\t%s\n",
			"map[int]interface {}(map[3:OK 4:<nil>])",
			"map[int]interface {}(map[3:OK 4:[]])",
			source,
		))

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			v1 := []int{1, 2, 3}
			v2 := []int64{1, 2, 3}
			func() { o(v1).Equals(v2); source = getFileLine(0) }()
		})).Equals(true, fmt.Sprintf(
			"\t1st argument does not equal\n\t"+
				"want:\n\t%s\n\tgot:\n\t%s\n\t%s\n",
			"[]int64([1 2 3])",
			"[]int([1 2 3])",
			source,
		))

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			v1 := []int{1, 2, 3}
			v2 := []int{1, 3, 2}
			func() { o(v1).Equals(v2); source = getFileLine(0) }()
		})).Equals(true, fmt.Sprintf(
			"\t1st argument does not equal\n\t"+
				"want:\n\t%s\n\tgot:\n\t%s\n\t%s\n",
			"[]int([1 3 2])",
			"[]int([1 2 3])",
			source,
		))

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			v1 := make([]any, 0)
			func() { o(v1).Equals(nil); source = getFileLine(0) }()
		})).Equals(true, fmt.Sprintf(
			"\t1st argument does not equal\n\t"+
				"want:\n\t%s\n\tgot:\n\t%s\n\t%s\n",
			"<nil>(<nil>)",
			"[]interface {}([])",
			source,
		))

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			v1 := map[string]any{}
			func() { o(v1).Equals(nil); source = getFileLine(0) }()
		})).Equals(true, fmt.Sprintf(
			"\t1st argument does not equal\n\t"+
				"want:\n\t%s\n\tgot:\n\t%s\n\t%s\n",
			"<nil>(<nil>)",
			"map[string]interface {}(map[])",
			source,
		))
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(3).Equals(3)
		assert(nil).Equals(nil)
		assert((any)(nil)).Equals(nil)
		assert([]any(nil)).Equals(nil)
		assert(map[string]any(nil)).Equals(nil)
		assert((*Assert)(nil)).Equals(nil)
		assert(nil).Equals((*Assert)(nil))
		assert(nil).Equals((any)(nil))
		assert([]int{1, 2, 3}).Equals([]int{1, 2, 3})
		assert(map[int]string{3: "OK", 4: "NO"}).
			Equals(map[int]string{4: "NO", 3: "OK"})
		assert(1, 2, 3).Equals(1, 2, 3)
	})
}

func TestRpcAssert_IsNil(t *testing.T) {
	t.Run("arguments is empty", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o().IsNil(); source = getFileLine(0) }()
		})).Equals(true, "\targuments is empty\n\t"+source+"\n")
	})

	t.Run("arguments is not nil", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		getFL := getFileLine

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o([]any{}).IsNil(); source = getFL(0) }()
		})).Equals(true, "\t1st argument is not nil\n\t"+source+"\n")

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(map[string]any{}).IsNil(); source = getFL(0) }()
		})).Equals(true, "\t1st argument is not nil\n\t"+source+"\n")

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(uintptr(0)).IsNil(); source = getFL(0) }()
		})).Equals(true, "\t1st argument is not nil\n\t"+source+"\n")

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(nil, 0).IsNil(); source = getFL(0) }()
		})).Equals(true, "\t2nd argument is not nil\n\t"+source+"\n")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(nil).IsNil()
		assert(([]any)(nil)).IsNil()
		assert((map[string]any)(nil)).IsNil()
		assert((any)(nil)).IsNil()
		assert((*Assert)(nil)).IsNil()
		assert((unsafe.Pointer)(nil)).IsNil()
		assert(nil, (any)(nil)).IsNil()
	})
}

func TestRpcAssert_IsNotNil(t *testing.T) {
	t.Run("arguments is empty", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o().IsNotNil(); source = getFileLine(0) }()
		})).Equals(true, "\targuments is empty\n\t"+source+"\n")
	})

	t.Run("arguments is nil", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			getFL := getFileLine
			func() { o([]any(nil)).IsNotNil(); source = getFL(0) }()
		})).Equals(true, "\t1st argument is nil\n\t"+source+"\n")

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			v1 := map[string]any(nil)
			func() { o(v1).IsNotNil(); source = getFileLine(0) }()
		})).Equals(true, "\t1st argument is nil\n\t"+source+"\n")

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(nil).IsNotNil(); source = getFileLine(0) }()
		})).Equals(true, "\t1st argument is nil\n\t"+source+"\n")

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(0, nil).IsNotNil(); source = getFileLine(0) }()
		})).Equals(true, "\t2nd argument is nil\n\t"+source+"\n")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(0).IsNotNil()
		assert([]any{}).IsNotNil()
		assert(map[string]any{}).IsNotNil()
		assert(uintptr(0)).IsNotNil()
		assert(0, []any{}).IsNotNil()
	})
}

func TestRpcAssert_IsTrue(t *testing.T) {
	t.Run("arguments is empty", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o().IsTrue(); source = getFileLine(0) }()
		})).Equals(true, "\targuments is empty\n\t"+source+"\n")
	})

	t.Run("arguments is not true", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(nil).IsTrue(); source = getFileLine(0) }()
		})).Equals(true, "\t1st argument is not true\n\t"+source+"\n")

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(true, nil).IsTrue(); source = getFileLine(0) }()
		})).Equals(true, "\t2nd argument is not true\n\t"+source+"\n")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(true).IsTrue()
		assert(true, true).IsTrue()
	})
}

func TestRpcAssert_IsFalse(t *testing.T) {
	t.Run("arguments is empty", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""
		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o().IsFalse(); source = getFileLine(0) }()
		})).Equals(true, "\targuments is empty\n\t"+source+"\n")
	})

	t.Run("arguments is not false", func(t *testing.T) {
		assert := NewAssert(t)
		source := ""

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(nil).IsFalse(); source = getFileLine(0) }()
		})).Equals(true, "\t1st argument is not false\n\t"+source+"\n")

		assert(testFailHelper(func(o func(_ ...any) *Assert) {
			func() { o(false, nil).IsFalse(); source = getFileLine(0) }()
		})).Equals(true, "\t2nd argument is not false\n\t"+source+"\n")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(false).IsFalse()
		assert(false, false).IsFalse()
	})
}
