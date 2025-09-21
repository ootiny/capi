package main

import (
	"github.com/ootiny/capi/builder"
	"github.com/ootiny/capi/utils"
)

func main() {
	if err := builder.Build(); err != nil {
		panic(utils.DebugError(err))
	}
}
