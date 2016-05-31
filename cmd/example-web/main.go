package main

import (
	"os"

	"github.com/pkg/errors"

	"github.com/agalitsyn/goapi/service"
)

func main() {
	if err := service.Start(); err != nil {
		err = errors.Wrap(err, "Service start failed")
		errors.Fprint(os.Stdout, err)
		os.Exit(1)
	}
}
