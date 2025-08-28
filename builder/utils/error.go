package utils

import (
	"fmt"
	"sync/atomic"
)

const (
	ErrCodeGeneral = 1 // ErrCodeGeneral 表示通用错误码。
)

var (
	errorTracingEnabled = atomic.Bool{}
)

func FirstError(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}

	return nil
}

func SetTraceError(traceError bool) {
	errorTracingEnabled.Store(traceError)
}

func IsTraceError() bool {
	return errorTracingEnabled.Load()
}

type Error struct {
	message string
	code    int
	traces  []string
}

func (p *Error) Error() string {
	if p != nil {
		return p.message
	}

	return ""
}

func (p *Error) SetCode(code int) error {
	if p != nil {
		p.code = code
	}

	return p
}

func (p *Error) AddHeader(format string, args ...any) error {
	if p != nil {
		p.message = fmt.Sprintf(format, args...) + ": " + p.message
	}

	return p
}

func convertToError(err error) *Error {
	if err == nil {
		return nil
	}

	if v, ok := err.(*Error); ok {
		return v
	} else {
		return &Error{
			message: err.Error(),
			code:    ErrCodeGeneral,
			traces:  []string{},
		}
	}
}

func Errorf(format string, args ...any) *Error {
	ret := &Error{
		message: fmt.Sprintf(format, args...),
		code:    ErrCodeGeneral,
		traces:  []string{},
	}

	if IsTraceError() {
		ret.traces = append(ret.traces, getFileLine(1))
	}

	return ret
}

func WrapError(err error) *Error {
	ret := convertToError(err)

	if ret == nil {
		return nil
	}

	if IsTraceError() {
		ret.traces = append(ret.traces, getFileLine(1))
	}

	return ret
}

func DebugError(err error) string {
	e := convertToError(err)
	ret := fmt.Sprintf("Message: %s\n", e.message)
	ret += fmt.Sprintf("Code: %d\n", e.code)

	for i := len(e.traces) - 1; i >= 0; i-- {
		ret += fmt.Sprintf("\t[Stack %04d]: %s\n", i+1, e.traces[i])
	}
	return ret
}
