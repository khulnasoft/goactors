package actor

import (
	"context"
	"log/slog"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/khulnasoft/goactors/safemap"
)

// Context represents the context of an actor, containing information about the actor's state, its children, and its parent.
type Context struct {
	pid      *PID
	sender   *PID
	engine   *Engine
	receiver Receiver
	message  any
	// the context of the parent if we are a child.
	// we need this parentCtx, so we can remove the child from the parent Context
	// when the child dies.
	parentCtx *Context
	children  *safemap.SafeMap[string, *PID]
	context   context.Context
}

// newContext creates a new Context instance.
func newContext(ctx context.Context, e *Engine, pid *PID) *Context {
	return &Context{
		context:  ctx,
		engine:   e,
		pid:      pid,
		children: safemap.New[string, *PID](),
	}
}

// Context returns a context.Context, user defined on spawn or
// a context.Background as default
func (c *Context) Context() context.Context {
	return c.context
}

// Receiver returns the underlying receiver of this Context.
func (c *Context) Receiver() Receiver {
	return c.receiver
}

// Request sends a request to the given PID with a specified timeout and returns a Response.
func (c *Context) Request(pid *PID, msg any, timeout time.Duration) *Response {
	return c.engine.Request(pid, msg, timeout)
}

// Respond sends the given message to the sender of the current received message.
func (c *Context) Respond(msg any) {
	if c.sender == nil {
		slog.Warn("context got no sender", "func", "Respond", "pid", c.PID())
		return
	}
	c.engine.Send(c.sender, msg)
}

// SpawnChild spawns a child actor with the given Producer and name, and returns its PID.
func (c *Context) SpawnChild(p Producer, name string, opts ...OptFunc) *PID {
	options := DefaultOpts(p)
	options.Kind = c.PID().ID + pidSeparator + name
	for _, opt := range opts {
		opt(&options)
	}
	// Check if we got an ID, generate otherwise
	if len(options.ID) == 0 {
		id := strconv.Itoa(rand.Intn(math.MaxInt))
		options.ID = id
	}
	proc := newProcess(c.engine, options)
	proc.context.parentCtx = c
	pid := c.engine.SpawnProc(proc)
	c.children.Set(pid.ID, pid)

	return proc.PID()
}

// SpawnChildFunc spawns a child actor with the given function and name, and returns its PID.
func (c *Context) SpawnChildFunc(f func(*Context), name string, opts ...OptFunc) *PID {
	return c.SpawnChild(newFuncReceiver(f), name, opts...)
}

// Send sends the given message to the specified PID, setting the sender to the current Context's PID.
func (c *Context) Send(pid *PID, msg any) {
	c.engine.SendWithSender(pid, msg, c.pid)
}

// SendRepeat sends the given message to the specified PID at the specified interval, and returns a SendRepeater.
func (c *Context) SendRepeat(pid *PID, msg any, interval time.Duration) SendRepeater {
	sr := SendRepeater{
		engine:   c.engine,
		self:     c.pid,
		target:   pid.CloneVT(),
		interval: interval,
		msg:      msg,
		cancelch: make(chan struct{}, 1),
	}
	sr.start()
	return sr
}

// Forward forwards the current received message to the specified PID, setting the sender to the current Context's PID.
func (c *Context) Forward(pid *PID) {
	c.engine.SendWithSender(pid, c.message, c.pid)
}

// GetPID returns the PID of the process found by the given id, or nil if not found.
func (c *Context) GetPID(id string) *PID {
	proc := c.engine.Registry.getByID(id)
	if proc != nil {
		return proc.PID()
	}
	return nil
}

// Parent returns the PID of the parent process, or nil if there is no parent.
func (c *Context) Parent() *PID {
	if c.parentCtx != nil {
		return c.parentCtx.pid
	}
	return nil
}

// Child returns the PID of the child process with the given id, or nil if not found.
func (c *Context) Child(id string) *PID {
	pid, _ := c.children.Get(id)
	return pid
}

// Children returns a slice of PIDs of all child processes.
func (c *Context) Children() []*PID {
	pids := make([]*PID, c.children.Len())
	i := 0
	c.children.ForEach(func(_ string, child *PID) {
		pids[i] = child
		i++
	})
	return pids
}

// PID returns the PID of the current process.
func (c *Context) PID() *PID {
	return c.pid
}

// Sender returns the PID of the sender of the current received message, or nil if there is no sender.
func (c *Context) Sender() *PID {
	return c.sender
}

// Engine returns the underlying Engine.
func (c *Context) Engine() *Engine {
	return c.engine
}

// Message returns the current received message.
func (c *Context) Message() any {
	return c.message
}
