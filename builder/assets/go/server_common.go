package _rt_package_name_

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	ErrInternal    = 1000
	ErrAPINotFound = 2000
	ErrAPIExec     = 2001
	ErrAPICustom   = 2002
	ErrDBCustom    = 3000
)

var gAPIMap = map[string]func(ctx *Context, data []byte) *Return{}

type Request interface {
	Action() string
	Data() []byte
	Cookie(name string) (*http.Cookie, error)
	Header(name string) string
}

type Response interface {
	SetHeader(name string, value string)
	WriteHeader(code int)
	WriteJson(data []byte) (int, error)
}

func NewContext(request Request, response Response) *Context {
	return &Context{
		request:  request,
		response: response,
	}
}

type Context struct {
	request  Request
	response Response
}

func (p *Context) Request() Request {
	return p.request
}

func (p *Context) Response() Response {
	return p.response
}

func (p *Context) Close(dbCommit bool) error {
	return nil
}

type Return struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func RegisterHandler(action string, handler func(ctx *Context, data []byte) *Return) {
	gAPIMap[action] = handler
}

func JsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func evalAction(w Response, r Request) (ret *Return) {
	ctx := NewContext(r, w)
	action := ctx.Request().Action()
	data := ctx.Request().Data()
	fn, ok := gAPIMap[action]

	if !ok {
		return &Return{
			Code:    ErrAPINotFound,
			Message: fmt.Sprintf("api %s not found", action),
		}
	}

	defer func() {
		if reason := recover(); reason != nil {
			_ = ctx.Close(false)
			ret = &Return{
				Code:    ErrAPIExec,
				Message: fmt.Sprintf("action exec error: %s", reason),
			}
		} else if ret != nil {
			if closeErr := ctx.Close(ret.Code == 0); closeErr != nil {
				ret = &Return{
					Code:    ErrInternal,
					Message: fmt.Sprintf("close error: %s", closeErr.Error()),
				}
			}
		} else {
			_ = ctx.Close(false)
			ret = &Return{
				Code:    ErrInternal,
				Message: "return is nil",
			}
		}
	}()

	return fn(ctx, data)
}

func apiHandler(cors bool, w Response, r Request) {
	// Set CORS headers
	if cors {
		w.SetHeader("Access-Control-Allow-Origin", "*")

		// Handle preflight OPTIONS request only when CORS is enabled
		if r.Header("Access-Control-Request-Method") == "OPTIONS" {
			w.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	ret := evalAction(w, r)
	if ret.Code == 0 {
		ret.Code = http.StatusOK
	}

	w.SetHeader("Content-Type", "application/json")

	if retBytes, err := json.Marshal(ret); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.WriteJson([]byte(`{"code":500,"message":"Internal Server Error"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = w.WriteJson(retBytes)
	}
}
