package help

import (
	"flag"
	"fmt"
	"os"
)

func init() {
	flag.Usage = func() {
		fmt.Println(help)
		flag.CommandLine.PrintDefaults()
	}

	flag.BoolFunc("help", "shows startup help (this)", func(s string) error {
		flag.Usage()
		return nil
	})

	flag.BoolFunc("example-scheme", "print example scheme config", func(s string) error {
		fmt.Println(exampleScheme)
		return nil
	})

	flag.Parse()

	if flag.NFlag() > 0 {
		os.Exit(0)
	}
}
