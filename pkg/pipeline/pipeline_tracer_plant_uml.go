package pipeline

import (
	"context"
	"fmt"
	"time"
)

func (t *tracer) TracedDiagram(txnID string) string {
	t.end = now()

	output := "@startuml \nstart\n"
	output += string(t.skin)

	output += fmt.Sprintf("\n title <font size=32>%v</font>\n \n\n\n\n", t.name)
	output += fmt.Sprintf("\n left header total time:  %.4fms \n\n\n", elapsedTime(t.start, t.end))
	output += ":source; \n"
	output += notes(t.sourceNotes)

	hasNodes := len(t.nodes) > 0
	if hasNodes {
		output += executionArrow(t.start, t.nodes[0].startTime)
		output += traceBranch(t.nodes)
	}

	output += "(★) \n"
	output += ":sink; \n"
	output += notes(t.sinkNotes)

	if hasNodes {
		node := t.nodes[len(t.nodes)-1]
		node.mtx.Lock()
		output += executionArrow(node.endTime, t.end)
		node.mtx.Unlock()
	}

	output += "stop \n"
	output += fmt.Sprintf("right footer <font color=darkRed>★</font> pipeline tracer - //txn//:**%s**\n", txnID)
	output += "@enduml\n"

	return output
}

func executionArrow(start, end time.Time) string {
	if end == start || (start == time.Time{} || end == time.Time{}) {
		return "-[dotted]-> //canceled//; \n"
	}

	return fmt.Sprintf("-[dotted]-> //%.4fms//; \n", elapsedTime(start, end))
}

func TracedLink(ctx context.Context) string {
	if disabled(ctx) {
		return ""
	}

	diagram := ctx.Value(tracerKey).(*tracer).TracedDiagram(txnID(ctx))

	return fmt.Sprint(RenderServer, Encoded(diagram))
}

func traceBranch(b []*tracerNode) string {
	var out string

	for _, node := range b {
		node.mtx.Lock()

		out += node.pipe.traced(node)
		out += executionArrow(node.startTime, node.endTime)

		node.mtx.Unlock()
	}

	return out
}

func elapsedTime(start, end time.Time) float64 {
	if (start == time.Time{} || end == time.Time{}) {
		return 0
	}

	return float64(end.Sub(start)) / 1000000
}

func txnID(_ context.Context) string {
	return ""
}

func notesOf(n *tracerNode) string {
	out := ""

	if n.annotations != nil && len(n.annotations.notes) > 0 {
		out += "note right \n"

		for idx, note := range n.annotations.notes {
			out += note
			out += "\n"

			if idx < len(n.annotations.notes)-1 {
				out += "----\n"
			}
		}

		out += "end note \n"
	}

	return out
}

func notes(notes []string) string {
	out := ""

	if len(notes) > 0 {
		out += "note right \n"

		for idx, note := range notes {
			out += note
			out += "\n"

			if idx < len(notes)-1 {
				out += "----\n"
			}
		}

		out += "end note \n"
	}

	return out
}

func drawFlowVolume(location string, flowVolume flowsPercent) string {
	var output string

	if flowVolume.total > 0 {
		output += fmt.Sprintf(": //%.2f%%// /\n", flowVolume.total)

		if len(flowVolume.tagged) > 0 {
			output += fmt.Sprintf("note %s \n%s end note\n", location, flowVolume.asList())
		}
	}

	return output
}
