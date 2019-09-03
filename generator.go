package enumgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

const preamble = "Code generated by 'github.com/shabbyrobe/go-enumgen'. DO NOT EDIT."

type kind int

const (
	intKind    kind = 1
	stringKind kind = 2
)

type packageInfo struct {
	fullName string
	name     string
	defs     map[*ast.Ident]types.Object
}

type constantValue struct {
	Name  string
	Value string
	Const *types.Const
}

type constants struct {
	FullName string
	Name     string
	Empty    string
	Kind     kind
	Values   []constantValue
}

type generator struct {
	buf         bytes.Buffer
	format      bool
	withFlagVal bool
	withMarshal bool
}

func (g *generator) Output(fileName string, pkgInfo *packageInfo) ([]byte, error) {
	var out bytes.Buffer
	out.Write([]byte("// " + preamble))
	out.WriteString("\n\n")
	out.WriteString(fmt.Sprintf("package %s", pkgInfo.name))
	io.Copy(&out, &g.buf)

	bts := out.Bytes()
	if g.format {
		var err error
		bts, err = imports.Process(fileName, bts, nil)
		if err != nil {
			return nil, err
		}
	}
	return bts, nil
}

func (g *generator) extract(pkg *packageInfo, typeName string) (*constants, error) {
	var def types.Object
	for _, cur := range pkg.defs {
		if cur == nil {
			continue
		}
		if _, ok := cur.Type().(*types.Named); !ok {
			continue
		}
		if cur.Name() == typeName {
			def = cur
			break
		}
	}
	if def == nil {
		return nil, fmt.Errorf("could not find def for %s", typeName)
	}

	fullName := pkg.fullName + "." + typeName
	cs := &constants{
		FullName: fullName,
		Name:     typeName,
	}

	underlying := def.Type().Underlying().(*types.Basic).Info()
	if underlying&types.IsInteger != 0 {
		cs.Empty = "0"
		cs.Kind = intKind

	} else if underlying&types.IsString != 0 {
		cs.Empty = `""`
		cs.Kind = stringKind

	} else {
		return nil, fmt.Errorf("type %q is not a string or integer type", typeName)
	}

	for _, cur := range pkg.defs {
		if cur == nil {
			continue
		}

		if cur.Type().String() != fullName {
			continue
		}
		if cns, ok := cur.(*types.Const); ok {
			cs.Values = append(cs.Values, constantValue{
				Name:  cns.Name(),
				Value: cns.Val().ExactString(),
				Const: cns,
			})
		}
	}

	sort.Slice(cs.Values, func(i, j int) bool {
		return cs.Values[i].Name < cs.Values[j].Name
	})

	return cs, nil
}

func (g *generator) parsePackage(pkgName string, tags []string) (*packageInfo, error) {
	cfg := &packages.Config{
		Mode:       packages.LoadSyntax,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(cfg, pkgName)
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("package %q not found", pkgName)
	}

	pkg := pkgs[0]
	return &packageInfo{
		fullName: pkg.PkgPath,
		name:     pkg.Name,
		defs:     pkg.TypesInfo.Defs,
	}, nil
}

func (g *generator) generate(cns *constants) error {
	var data = &templateData{
		Receiver:    "v",
		Unknown:     "<unknown>",
		Constants:   cns,
		Type:        cns.Name,
		WithMarshal: g.withMarshal,
		WithFlagVal: g.withFlagVal,
	}
	if err := genTpl.Execute(&g.buf, data); err != nil {
		return err
	}

	var tpl *template.Template
	switch cns.Kind {
	case intKind:
		tpl = intTpl
	case stringKind:
		tpl = strTpl
	default:
		return fmt.Errorf("unsupported kind")
	}
	if err := tpl.Execute(&g.buf, data); err != nil {
		return err
	}

	return nil
}

type templateData struct {
	Receiver    string
	Type        string
	Constants   *constants
	Unknown     string
	WithMarshal bool
	WithFlagVal bool
}

var genTpl = template.Must(template.New("").Parse(genTplText))
var intTpl = template.Must(template.New("").Parse(intTplText))
var strTpl = template.Must(template.New("").Parse(strTplText))

var genTplText = `
func ({{.Receiver}} {{.Type}}) Name() string {
	switch {{.Receiver}} {
	{{- range .Constants.Values }}
	case {{ .Name }}:
		return {{ printf "%q" .Name }}
	{{- end }}
	default:
		return ""
	}
}

func ({{.Receiver}} {{.Type}}) Lookup(name string) (value {{.Type}}, ok bool) {
	switch name {
	{{- range .Constants.Values }}
	case {{ printf "%q" .Name }}:
		return {{.Name}}, true
	{{- end }}
	default:
		return {{ .Constants.Empty }}, false
	}
}

func ({{.Receiver}} {{.Type}}) IsValid() bool {
	switch {{.Receiver}} {
	{{- range .Constants.Values }}
	case {{ .Name }}:
	{{- end }}
	default:
		return false
	}
	return true
}
`

var intTplText = `
func ({{.Receiver}} {{.Type}}) String() string {
	switch {{.Receiver}} {
	{{- range .Constants.Values }}
	case {{ .Name }}:
		return "{{ .Name }}({{.Value}})"
	{{- end }}
	default:
		return {{ printf "%q" .Unknown }}
	}
}

{{ if .WithMarshal }}
func ({{.Receiver}} {{.Type}}) MarshalText() (text []byte, err error) {
	switch {{.Receiver}} {
	{{- range .Constants.Values }}
	case {{ .Name }}:
		return []byte({{printf "%q" .Value}}), nil
	{{- end }}
	default:
		return fmt.Errorf("could not marshal enum %T containing invalid value %q", {{.Receiver}}, s)
	}
}

func ({{.Receiver}} *{{.Type}}) UnmarshalText(text []byte) (err error) {
	switch string({{.Receiver}}) {
	{{- range .Constants.Values }}
	case {{ printf "%q" .Name }}, {{ printf "%q" .Value }}:
		*{{$.Receiver}} = {{.Name}}
	{{- end }}
	default:
		return fmt.Errorf("could not marshal enum %T containing invalid value %q", {{.Receiver}}, s)
	}
	return nil
}
{{ end }}

{{ if .WithFlagVal }}
func ({{.Receiver}} *{{.Type}}) Set(s string) error {
	parsed, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*{{.Receiver}} = {{.Type}}(parsed)
	return nil
}
{{ end }}
`

var strTplText = `
func ({{.Receiver}} {{.Type}}) String() string {
	switch {{.Receiver}} {
	{{- range .Constants.Values }}
	case {{ .Name }}:
		return {{ .Value }}
	{{- end }}
	default:
		return {{ printf "%q" .Unknown }}
	}
}

{{ if .WithMarshal }}
func ({{.Receiver}} {{.Type}}) MarshalText() (text []byte, err error) {
	switch {{.Receiver}} {
	{{- range .Constants.Values }}
	case {{ .Name }}:
		return []byte({{.Name}}), nil
	{{- end }}
	default:
		return fmt.Errorf("could not marshal enum %T containing invalid value %q", {{.Receiver}}, s)
	}
}

func ({{.Receiver}} *{{.Type}}) UnmarshalText(text []byte) (err error) {
	switch string({{.Receiver}}) {
	{{- range .Constants.Values }}
	case {{ .Value }}:
		*{{$.Receiver}} = {{.Name}}
	{{- end }}
	default:
		return fmt.Errorf("could not marshal enum %T containing invalid value %q", {{.Receiver}}, s)
	}
	return nil
}
{{ end }}

{{ if .WithFlagVal }}
func ({{.Receiver}} *{{.Type}}) Set(s string) error {
	*{{.Receiver}} = {{.Type}}(s)
	if !{{.Receiver}}.IsValid() {
		return fmt.Errorf("enum %T received invalid value %q", {{.Receiver}}, s)
	}
	return nil
}
{{ end }}
`
