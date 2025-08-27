package utils

import (
	"fmt"
	"runtime"
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

func getFileLine(skip int) string {
	if _, file, line, ok := runtime.Caller(int(skip) + 1); ok && line > 0 {
		return fmt.Sprintf("%s:%d", file, line)
	} else {
		return ""
	}
}

type Error struct {
	message string
	code    int
	traces  []string
}

func (p *Error) Error() string {
	if p == nil {
		return ""
	}

	return p.message
}

func (p *Error) SetCode(code int) error {
	if p == nil {
		return nil
	}

	p.code = code
	return p
}

func (p *Error) AddHeader(format string, args ...any) error {
	if p == nil {
		return nil
	}

	p.message = fmt.Sprintf(format, args...) + ": " + p.message
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
	ret := e.message
	for _, trace := range e.traces {
		ret += "\n\t" + trace
	}
	return ret
}
