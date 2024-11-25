package pipeline

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

const (
	mapper       = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	RenderServer = "http://plantuml.com/plantuml/svg/~1"
	secret       = "6368616e676520746869732070617373"

	methodFunction = 3
	function       = 2

	hSixtyThree = 0x3f
	hFifteen    = 0xF
	hThree      = 0x3
)

var key, _ = hex.DecodeString(secret)

func (p *Pipeline) Diagram() string {
	output := "@startuml\nstart\n"

	output += p.skin()
	output += p.title()
	output += p.source()
	output += p.flow()
	output += p.sink()

	output += "stop\n@enduml\n"

	return output
}

func (p *Pipeline) title() string {
	var output string

	output += fmt.Sprintf("\ntitle\n\n<size:42>%s</size>\n\n\nend title\n", p.Name)
	if p.Description != "" {
		output += fmt.Sprintf("\nright footer <font size=13>//%v//</font>\n \n\n\n\n", p.Description)
	}

	return output
}

func (p *Pipeline) skin() string {
	return string(p.BlueprintSkin)
}

func (p *Pipeline) source() string {
	var output string

	output += drawStage(p.Source)
	output += "note left\n <font size=\"16\">//source ⟶ //</font>  \n end note\n"

	if p.SourceComments != "" {
		output += fmt.Sprintf("note right\n%s\n end note\n", p.SourceComments)
	}

	return output
}

func (p *Pipeline) flow() string {
	var output string

	for _, pipe := range p.Flow {
		output += pipe.draw(p.BlueprintSkin)
	}

	return output
}

func (p *Pipeline) sink() string {
	var output string

	output += "(★) \n"
	output += drawStage(p.Sink)
	output += "note right\n <font size=\"16\">// ⟵ sink//</font> \n  end note\n"

	if p.SinkComments != "" {
		output += fmt.Sprintf("note left\n%s\n end note\n", p.SinkComments)
	}

	return output
}

func (fp *flowsPercent) asList() string {
	var out string

	for k, v := range fp.tagged {
		out += fmt.Sprintf("** %s: //%.2f%%//\n", k, v)
	}

	return out
}

func drawStage(r interface{}) string {
	return fmt.Sprintf(": %s ; \n", formattedResolver(r))
}

func drawCanceledStage(r interface{}) string {
	return fmt.Sprintf(": --%s-- ; \n", formattedResolver(r))
}

func drawFailedStage(r interface{}, err error) string {
	out := fmt.Sprintf(": %s\n", formattedResolver(r))
	out += "----\n"
	out += fmt.Sprintf("<font size=\"$errorSize\" color=\"$errorColor\"> ☠ %s</font> ;\n", err.Error())

	return out
}

func formattedResolver(r interface{}) string {
	fullName, parts := resolverName(r)

	switch len(parts) {
	case methodFunction:
		return fmt.Sprintf(
			"<font color=\"$packageColor\">//%s.//</font>%s.**//%s//**",
			parts[0],
			parts[1],
			strings.Replace(parts[2], "-fm", "", 1),
		)

	case function:
		return fmt.Sprintf(
			"<font color=\"$packageColor\">//%s.//</font>**//%s//**",
			parts[0],
			strings.Replace(parts[1], "-fm", "", 1),
		)

	default:
		return fullName
	}
}

func resolverName(r interface{}) (string, []string) {
	fullName := funcName(r)
	slash := strings.Split(fullName, "/")
	split := strings.Split(strings.Join(slash[FunctionsNamePrefixPrune:], "/"), ".")

	return fullName, split
}

func funcName(r interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(r).Pointer()).Name()
}

func fontMultiline(raw string, fontTag string) string {
	var out string

	rows := strings.Split(raw, "\n")
	for i, r := range rows {
		out += fmt.Sprintf(" %s%s</font>", fontTag, r)
		if i < len(rows)-1 {
			out += "\n"
		}
	}

	return out
}
