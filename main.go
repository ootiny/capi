package main

import (
	"log"

	"github.com/ootiny/capi/builder"
)

func main() {
	if err := builder.Build(); err != nil {
		log.Panicf("Failed to build: %v", err)
	}
}
