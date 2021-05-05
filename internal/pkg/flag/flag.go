package flag

import (
	"fmt"
	"os"
	"reflect"

	"github.com/jessevdk/go-flags"
)

// used in go build
var version string

func GetVersion() string {
	return version
}

func ParseFlags(cfg interface{}) {
	parser := flags.NewParser(cfg, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	v := reflect.ValueOf(cfg).Elem()
	if version := v.FieldByName("PrintVersion"); version.IsValid() && version.Bool() {
		fmt.Fprintln(os.Stdout, GetVersion())
		os.Exit(0)
	}
}
