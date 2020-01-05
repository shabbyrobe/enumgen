package enumgen

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"strings"
)

const Usage = `enumgen: turn a bag of constants into something a bit more useful

Usage: enumgen [options] <input>...

Inputs:

The <input> argument is a list of types contained in the current package, to which
methods will be added.

Methods generated:

- fmt.Stringer.String()
- IsValid() bool
- Lookup(str string) <T>
- encoding.TextMarshaler.MarshalText (with -textmarshal)
- encoding.TextUnmarshaler.UnmarshalText (with -textmarshal)
- flag.Value.Set(s string) (with -flagval)

If the underlying type is a string:

- ValuesString() (with -strvalues) - string of comma separated values (useful
  for things like flag help)
`

type usageError string

func (u usageError) Error() string { return string(u) }

func IsUsageError(err error) bool {
	_, ok := err.(usageError)
	return ok
}

type Command struct {
	switches
	tags   string
	pkg    string
	format bool
	out    string
}

func (cmd *Command) Flags(flags *flag.FlagSet) {
	flags.StringVar(&cmd.pkg, "pkg", ".", "package name to search for types")
	flags.StringVar(&cmd.out, "out", "enum_gen.go", "output file name")
	flags.StringVar(&cmd.tags, "tags", "", "comma-separated list of build tags")
	flags.BoolVar(&cmd.format, "format", true, "run gofmt on result")

	flags.BoolVar(&cmd.switches.WithName, "name", true, "generate Name()")
	flags.BoolVar(&cmd.switches.WithLookup, "lookup", true, "generate Lookup()")
	flags.BoolVar(&cmd.switches.WithFlagVal, "flag", true, "generate flag.Value")
	flags.BoolVar(&cmd.switches.WithIsValid, "isvalid", true, "generate IsValid()")
	flags.BoolVar(&cmd.switches.WithString, "string", true, "generate String()")
	flags.BoolVar(&cmd.switches.WithMarshal, "marshal", false, "EXPERIMENTAL: generate encoding.TextMarshaler/TextUnmarshaler")
	flags.BoolVar(&cmd.switches.WithValuesString, "strvalues", false, "generate ValuesString() if underlying enum type is string")
}

func (cmd *Command) Synopsis() string { return "Generate enum-ish helpers from a bag of constants" }

func (cmd *Command) Usage() string { return Usage }

func (cmd *Command) Run(args ...string) error {
	if cmd.pkg == "" {
		return usageError("-pkg not set")
	}
	if cmd.out == "" {
		return usageError("-out not set")
	}

	tags := strings.Split(cmd.tags, ",")

	g := &generator{
		switches: cmd.switches,
		format:   cmd.format,
	}

	pkg, err := g.parsePackage(cmd.pkg, tags)
	if err != nil {
		return err
	}

	for _, typeName := range args {
		cns, err := g.extract(pkg, typeName)
		if err != nil {
			return err
		}
		if err := g.generate(cns); err != nil {
			return err
		}
	}

	out, err := g.Output(cmd.out, pkg)
	if err != nil {
		return err
	}

	var write bool
	existing, err := ioutil.ReadFile(cmd.out)
	if os.IsNotExist(err) || err == nil {
		write = true
	} else if err != nil {
		return err
	} else if !bytes.Equal(out, existing) {
		write = true
	}

	if write {
		return ioutil.WriteFile(cmd.out, out, 0644)
	}
	return nil
}
