package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/shabbyrobe/enumgen"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	bmc := enumgen.Command{}

	var fs = flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)
	bmc.Flags(fs)

	args := os.Args[1:]
	if len(args) == 0 {
		help(&bmc, fs)
		return nil
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			help(&bmc, fs)
			return nil
		}
		return err
	}

	err := bmc.Run(fs.Args()...)
	if enumgen.IsUsageError(err) {
		help(&bmc, fs)
		fmt.Println()
		fmt.Println("error:", err)
		return nil
	}
	return err
}

func help(bmc *enumgen.Command, fs *flag.FlagSet) {
	fmt.Println(bmc.Usage())
	fmt.Println("Flags:")
	fs.SetOutput(os.Stdout)
	fs.PrintDefaults()
}
