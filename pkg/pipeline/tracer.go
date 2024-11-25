package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type tracerKeyType int

const (
	tracerEnabledKey tracerKeyType = iota
	tracerKey
)

var mark = struct{}{}

type (
	ctxKeysType string

	stopwatch interface {
		done()
		start(context.Context)
		canceled()
		fail(error)
	}

	tracer struct {
		name  string
		nodes []*tracerNode
		mtx   *sync.Mutex
		start time.Time
		end   time.Time

		skin Skin

		sourceNotes []string
		sinkNotes   []string
	}

	annotations struct {
		of    Traceable
		notes []string
	}

	tracerNode struct {
		pipe        Traceable
		branches    *branches
		mtx         *sync.Mutex
		startTime   time.Time
		endTime     time.Time
		error       error
		cancelled   bool
		annotations *annotations
	}

	dummyTask    struct{}
	dummyJotter  struct{}
	sourceJotter struct{ tracer *tracer }
	stageJotter  struct{ node *tracerNode }
	sinkJotter   struct{ tracer *tracer }

	TracerPointer struct {
		name    string
		pointer int
	}

	tracerNodes struct {
		mtx   *sync.Mutex
		nodes []*tracerNode
	}

	branches struct {
		inner sync.Map
	}
)

func (j *sourceJotter) Note(note string) {
	j.tracer.sourceNotes = append(j.tracer.sourceNotes, note)
}
func (j *sourceJotter) LazyNote(noteFn func() string) {
	j.Note(noteFn())
}

func (j *stageJotter) Note(note string) {
	defer j.node.mtx.Unlock()
	j.node.mtx.Lock()

	j.node.annotations.notes = append(j.node.annotations.notes, note)
}
func (j *stageJotter) LazyNote(noteFn func() string) {
	j.Note(noteFn())
}

func (j *sinkJotter) Note(note string) {
	j.tracer.sinkNotes = append(j.tracer.sinkNotes, note)
}
func (j *sinkJotter) LazyNote(noteFn func() string) {
	j.Note(noteFn())
}

func SourceNote(ctx context.Context) Jotter {
	if disabled(ctx) {
		return &dummyJotter{}
	}

	return &sourceJotter{tracer: ctx.Value(tracerKey).(*tracer)}
}

func SinkNote(ctx context.Context) Jotter {
	if disabled(ctx) {
		return &dummyJotter{}
	}

	return &sinkJotter{tracer: ctx.Value(tracerKey).(*tracer)}
}

func WithNote(ctx context.Context) Jotter {
	if disabled(ctx) {
		return &dummyJotter{}
	}

	annotationsOf := ctx.Value(tracerKey).(*tracer).findAnnotationsOf(ctx)
	if annotationsOf == nil {
		return &dummyJotter{}
	}

	return &stageJotter{node: annotationsOf}
}

func TracerOn(ctx context.Context) bool {
	return !disabled(ctx)
}

func newTracer(ctx context.Context, p *Pipeline) context.Context {
	var (
		pointer    = &TracerPointer{name: "root"}
		rootTracer = root(p)

		pointerCtx = context.WithValue(ctx, rootTracer.stagePointer(), pointer)
		tracerCtx  = context.WithValue(pointerCtx, tracerKey, rootTracer)
	)

	return context.WithValue(tracerCtx, tracerEnabledKey, mark)
}

func (t *tracer) traceExecution(ctx context.Context, node *tracerNode) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if n := ctx.Value(t.currentBranchKey()); n != nil {
		b := n.(string)
		if b != "" {
			c := ctx.Value(t.currentTracerKey()).(*tracerNode)
			c.branches.append(b, node)

			return
		}
	}

	t.nodes = upsert(t.nodes, node)
}

func (t *tracer) getOrCreate(ctx context.Context, root Traceable) *tracerNode {
	for _, n := range t.nodes {
		if n.pipe == root {
			return n
		}
	}

	if x := ctx.Value(t.currentTracerKey()); x != nil {
		cn := x.(*tracerNode)

		if n := ctx.Value(t.currentBranchKey()); n != nil {
			bn := n.(string)

			for _, n := range cn.branches.get(bn) {
				if n.pipe == root {
					return n
				}
			}
		}
	}

	return newNode(root)
}

func (t *tracer) findAnnotationsOf(ctx context.Context) *tracerNode {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	pointer := ctx.Value(t.stagePointer()).(*TracerPointer)

	if n := ctx.Value(t.currentBranchKey()); n != nil {
		b := n.(string)
		if b != "" {
			c := ctx.Value(t.currentTracerKey()).(*tracerNode)
			return c.branches.get(b)[pointer.pointer-1]
		}
	}

	return t.nodes[pointer.pointer-1]
}

func (t *tracer) currentTracerKey() ctxKeysType {
	return ctxKeysType(fmt.Sprintf("%v.tracer", t.name))
}

func (t *tracer) currentBranchKey() ctxKeysType {
	return ctxKeysType(fmt.Sprintf("%v.branch", t.name))
}

func (t *tracer) stagePointer() ctxKeysType {
	return ctxKeysType(fmt.Sprintf("%v.pointer", t.name))
}

func (n *tracerNode) start(ctx context.Context) {
	defer n.mtx.Unlock()

	t := ctx.Value(tracerKey).(*tracer)

	n.mtx.Lock()
	n.startTime = now()
	ctx.Value(t.stagePointer()).(*TracerPointer).pointer++
}

func (n *tracerNode) done() {
	defer n.mtx.Unlock()

	n.mtx.Lock()
	n.endTime = now()
}

func (n *tracerNode) canceled() {
	defer n.mtx.Unlock()

	n.mtx.Lock()
	n.endTime = n.startTime
	n.cancelled = true
}

func (n *tracerNode) fail(err error) {
	defer n.mtx.Unlock()

	n.mtx.Lock()
	n.endTime = now()
	n.error = err
}

func traceMe(ctx context.Context, pipe Traceable) stopwatch {
	if disabled(ctx) {
		return &dummyTask{}
	}

	node := newNode(pipe)
	tracer := ctx.Value(tracerKey).(*tracer)
	tracer.traceExecution(ctx, node)

	return node
}

func openBranch(ctx context.Context, root Traceable, name string) context.Context {
	if disabled(ctx) {
		return ctx
	}

	t := ctx.Value(tracerKey).(*tracer)
	t.mtx.Lock()
	defer t.mtx.Unlock()

	parentPointer := ctx.Value(t.stagePointer()).(*TracerPointer)
	rootNode := t.getOrCreate(ctx, root)
	branchCtx := context.WithValue(ctx, t.currentBranchKey(), name)
	pointerCtx := context.WithValue(branchCtx, t.stagePointer(), &TracerPointer{name: parentPointer.name + "." + name})

	return context.WithValue(pointerCtx, t.currentTracerKey(), rootNode)
}

func root(p *Pipeline) *tracer {
	if p.Name == "" {
		p.Name = fmt.Sprintf("[no title for %v]", &p)
	}

	return &tracer{
		name:  p.Name,
		nodes: make([]*tracerNode, 0),
		mtx:   &sync.Mutex{},
		start: now(),
		skin:  p.TraceSkin,
	}
}

func now() time.Time {
	return time.Now()
}

func newNode(pipe Traceable) *tracerNode {
	return &tracerNode{
		pipe:        pipe,
		branches:    newBranches(),
		mtx:         &sync.Mutex{},
		startTime:   now(),
		annotations: annotation(pipe),
	}
}

func upsert(s []*tracerNode, node *tracerNode) []*tracerNode {
	if s == nil {
		return []*tracerNode{node}
	}

	for _, n := range s {
		if n.pipe == node.pipe {
			return s
		}
	}

	return append(s, node)
}

func disabled(ctx context.Context) bool {
	return ctx.Value(tracerEnabledKey) == nil
}

func annotation(pipe Traceable) *annotations {
	return &annotations{
		of:    pipe,
		notes: []string{},
	}
}

func (*dummyTask) start(context.Context)    {}
func (*dummyTask) done()                    {}
func (*dummyTask) canceled()                {}
func (*dummyTask) fail(error)               {}
func (*dummyJotter) Note(string)            {}
func (*dummyJotter) LazyNote(func() string) {}

func newTracerNodes() *tracerNodes {
	return &tracerNodes{mtx: &sync.Mutex{}}
}

func (t *tracerNodes) append(nodes ...*tracerNode) *tracerNodes {
	t.mtx.Lock()
	t.nodes = append(t.nodes, nodes...)
	t.mtx.Unlock()

	return t
}

func (t *tracerNodes) dump() []*tracerNode {
	t.mtx.Lock()
	out := make([]*tracerNode, len(t.nodes))

	copy(out, t.nodes)
	t.mtx.Unlock()

	return out
}

func newBranches() *branches {
	return &branches{inner: sync.Map{}}
}

func (b *branches) append(branch string, node *tracerNode) {
	current, loaded := b.inner.LoadOrStore(branch, newTracerNodes().append(node))
	if loaded {
		current.(*tracerNodes).append(node)
	}
}

func (b *branches) dump() map[string][]*tracerNode {
	out := make(map[string][]*tracerNode)

	b.inner.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(*tracerNodes).dump()
		return true
	})

	return out
}

func (b *branches) get(branch string) []*tracerNode {
	var (
		value, _ = b.inner.Load(branch)
		tNodes   = value.(*tracerNodes)
	)

	tNodes.mtx.Lock()

	var (
		nodes = tNodes.nodes
		out   = make([]*tracerNode, len(nodes))
	)

	copy(out, nodes)
	tNodes.mtx.Unlock()

	return out
}
