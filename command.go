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

The <input> argument is a list of types contained in the current package, to which
methods will be added.

Things that are always generated:
    func (t <T>) String() string

    func (t <T>) IsValid() bool
        Is the value in 't' a member of the set of constants?
        
    func (t <T>) Lookup(name string) (<T>, ok bool)
        Find the constant for 'name'. 'ok' is false if not found.

    var <T>Values = []T{...}
        Slice containing all values for this enum, sorted by value (not name).

If -textmarshal is passed:
    func (t <T>)  MarshalText() (text []byte, err error)
    func (t *<T>) UnmarshalText(text []byte) (err error)

If -flag=(val|get) are passed (implements flag.Value):
    func (t <T>) Set(s string) error

If -flag=get is passed (implements flag.Getter):
    func (t <T>) Get() interface{}

If -strvalues is passed and the underlying type of your enum is a string:
    func (t <T>) ValuesString() string
        String of all constant values, separated by commas. This is useful for
        things like flag help, where you may want to show the list of possible
        options.
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
	cmd.switches.FlagMode = flagVal

	flags.StringVar(&cmd.pkg, "pkg", ".", "package name to search for types")
	flags.StringVar(&cmd.out, "out", "enum_gen.go", "output file name")
	flags.StringVar(&cmd.tags, "tags", "", "comma-separated list of build tags")
	flags.BoolVar(&cmd.format, "format", true, "run gofmt on result")

	flags.BoolVar(&cmd.switches.WithName, "name", true, "generate Name()")
	flags.BoolVar(&cmd.switches.WithLookup, "lookup", true, "generate Lookup()")
	flags.Var(&cmd.switches.FlagMode, "flag", "'val': flag.Value, 'get': flag.Getter, or 'none'. Default: val")
	flags.BoolVar(&cmd.switches.WithIsValid, "isvalid", true, "generate IsValid()")
	flags.BoolVar(&cmd.switches.WithString, "string", true, "generate String()")
	flags.BoolVar(&cmd.switches.WithMarshal, "marshal", false, "EXPERIMENTAL: generate encoding.TextMarshaler/TextUnmarshaler")
	flags.BoolVar(&cmd.switches.WithValuesString, "strvalues", false, "generate ValuesString() if underlying enum type is string")
	flags.BoolVar(&cmd.switches.WithValues, "values", true, "generate 'var <T>Values = []<T>{}' slice")
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
