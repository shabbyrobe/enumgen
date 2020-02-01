package enumgen

import "text/template"

type templateData struct {
	switches
	Receiver  string
	Type      string
	Constants *constants
	Unknown   string
}

var genTpl = template.Must(template.New("").Parse(genTplText))
var intTpl = template.Must(template.New("").Parse(intTplText))
var strTpl = template.Must(template.New("").Parse(strTplText))

const genTplText = genWithName + genWithLookup + genWithIsValid + genWithValues

const genWithName = `
{{ if .WithName }}
func ({{.Receiver}} {{.Type}}) Name() string {
	switch {{.Receiver}} {
	{{- range .Constants.NameOrder }}
	case {{ .Name }}:
		return {{ printf "%q" .Name }}
	{{- end }}
	default:
		return ""
	}
}
{{ end }}
`

const genWithLookup = `
{{ if .WithLookup }}
func ({{.Receiver}} {{.Type}}) Lookup(name string) (value {{.Type}}, ok bool) {
	switch name {
	{{- range .Constants.NameOrder }}
	case {{ printf "%q" .Name }}:
		return {{.Name}}, true
	{{- end }}
	default:
		return {{ .Constants.Empty }}, false
	}
}
{{ end }}
`

const genWithIsValid = `
{{ if .WithIsValid }}
func ({{.Receiver}} {{.Type}}) IsValid() bool {
	switch {{.Receiver}} {
	{{- range .Constants.NameOrder }}
	case {{ .Name }}:
	{{- end }}
	default:
		return false
	}
	return true
}
{{ end }}
`

const genWithValues = `
{{ if and .Constants.IsNamedType .WithValues }}
var {{.Type}}Values = []{{.Type}}{
	{{- range .Constants.ValueOrder }}
	{{ .Name }},
	{{- end }}
}
{{ end }}
`

const intTplText = genIntWithString + genIntWithMarshal + genIntWithFlag

const genIntWithString = `
{{ if .WithString }}
func ({{.Receiver}} {{.Type}}) String() string {
	switch {{.Receiver}} {
	{{- range .Constants.NameOrder }}
	case {{ .Name }}:
		return "{{ .Name }}({{.Value}})"
	{{- end }}
	default:
		return {{ printf "%q" .Unknown }}
	}
}
{{ end }}
`

const genIntWithMarshal = `
{{ if .WithMarshal }}
func ({{.Receiver}} {{.Type}}) MarshalText() (text []byte, err error) {
	switch {{.Receiver}} {
	{{- range .Constants.NameOrder }}
	case {{ .Name }}:
		return []byte({{printf "%q" .Value}}), nil
	{{- end }}
	default:
		return fmt.Errorf("could not marshal enum %T containing invalid value %q", {{.Receiver}}, s)
	}
}

func ({{.Receiver}} *{{.Type}}) UnmarshalText(text []byte) (err error) {
	switch string({{.Receiver}}) {
	{{- range .Constants.NameOrder }}
	case {{ printf "%q" .Name }}, {{ printf "%q" .Value }}:
		*{{$.Receiver}} = {{.Name}}
	{{- end }}
	default:
		return fmt.Errorf("could not marshal enum %T containing invalid value %q", {{.Receiver}}, s)
	}
	return nil
}
{{ end }}
`

const genIntWithFlag = `
{{ if .FlagMode.WithFlag }}
func ({{.Receiver}} *{{.Type}}) Set(s string) error {
	switch strings.ToLower(s) {
	{{- range .Constants.NameOrder }}
	case {{ printf "%q" .LowerName }}:
		*{{$.Receiver}} = {{ .Name }}
	{{- end }}
	default:
		parsed, err := strconv.ParseInt(s, 10, {{ .Constants.IntParseBits }})
		if err != nil {
			return err
		}
		*{{$.Receiver}} = {{.Type}}(parsed)
	}
	return nil
}
{{ end }}

{{ if eq .FlagMode "get" }}
func ({{.Receiver}} {{.Type}}) Get() interface{} {
	return {{.Receiver}}
}
{{ end }}
`

const strTplText = genStrWithString + genStrWithMarshal + genStrWithFlag + genStrWithValuesString

const genStrWithString = `
{{ if .WithString }}
func ({{.Receiver}} {{.Type}}) String() string {
	switch {{.Receiver}} {
	{{- range .Constants.NameOrder }}
	case {{ .Name }}:
		return {{ .Value }}
	{{- end }}
	default:
		return {{ printf "%q" .Unknown }}
	}
}
{{ end }}
`

const genStrWithMarshal = `
{{ if .WithMarshal }}
func ({{.Receiver}} {{.Type}}) MarshalText() (text []byte, err error) {
	switch {{.Receiver}} {
	{{- range .Constants.NameOrder }}
	case {{ .Name }}:
		return []byte({{.Name}}), nil
	{{- end }}
	default:
		return fmt.Errorf("could not marshal enum %T containing invalid value %q", {{.Receiver}}, s)
	}
}

func ({{.Receiver}} *{{.Type}}) UnmarshalText(text []byte) (err error) {
	switch string({{.Receiver}}) {
	{{- range .Constants.NameOrder }}
	case {{ .Value }}:
		*{{$.Receiver}} = {{.Name}}
	{{- end }}
	default:
		return fmt.Errorf("could not marshal enum %T containing invalid value %q", {{.Receiver}}, s)
	}
	return nil
}
{{ end }}
`

const genStrWithFlag = `
{{ if .FlagMode.WithFlag }}
func ({{.Receiver}} *{{.Type}}) Set(s string) error {
	*{{.Receiver}} = {{.Type}}(s)
	if !{{.Receiver}}.IsValid() {
		return fmt.Errorf("enum %T received invalid value %q", {{.Receiver}}, s)
	}
	return nil
}
{{ end }}

{{ if eq .FlagMode "get" }}
func ({{.Receiver}} {{.Type}}) Get() interface{} {
	return {{.Receiver}}
}
{{ end }}
`

const genStrWithValuesString = `
{{ if .WithValuesString }}
func ({{.Type}}) ValuesString() string {
	return {{.Constants.ValuesString | printf "%q"}}
}
{{ end }}
`
