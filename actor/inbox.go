package actor

import (
	"runtime"
	"sync/atomic"

	"github.com/khulnasoft/goactors/ringbuffer"
)

const (
	defaultThroughput = 300
	messageBatchSize  = 1024 * 4
)

const (
	stopped int32 = iota
	starting
	idle
	running
)

// Scheduler is an interface that defines the scheduling behavior for processing messages.
type Scheduler interface {
	Schedule(fn func())
	Throughput() int
}

type goscheduler int

func (goscheduler) Schedule(fn func()) {
	go fn()
}

func (sched goscheduler) Throughput() int {
	return int(sched)
}

// NewScheduler creates a new Scheduler with the specified throughput.
func NewScheduler(throughput int) Scheduler {
	return goscheduler(throughput)
}

// Inboxer is an interface that defines the behavior of an inbox for processing messages.
type Inboxer interface {
	Send(Envelope)
	Start(Processer)
	Stop() error
}

// Inbox represents an inbox for processing messages with concurrency handling.
type Inbox struct {
	rb         *ringbuffer.RingBuffer[Envelope]
	proc       Processer
	scheduler  Scheduler
	procStatus int32
}

// NewInbox creates a new Inbox with the specified size.
func NewInbox(size int) *Inbox {
	return &Inbox{
		rb:         ringbuffer.New[Envelope](int64(size)),
		scheduler:  NewScheduler(defaultThroughput),
		procStatus: stopped,
	}
}

// Send adds a message to the inbox and schedules its processing.
func (in *Inbox) Send(msg Envelope) {
	in.rb.Push(msg)
	in.schedule()
}

// schedule schedules the processing of messages in the inbox.
func (in *Inbox) schedule() {
	if atomic.CompareAndSwapInt32(&in.procStatus, idle, running) {
		in.scheduler.Schedule(in.process)
	}
}

// process processes the messages in the inbox.
func (in *Inbox) process() {
	in.run()
	atomic.CompareAndSwapInt32(&in.procStatus, running, idle)
}

// run runs the message processing loop.
func (in *Inbox) run() {
	i, t := 0, in.scheduler.Throughput()
	for atomic.LoadInt32(&in.procStatus) != stopped {
		if i > t {
			i = 0
			runtime.Gosched()
		}
		i++

		if msgs, ok := in.rb.PopN(messageBatchSize); ok && len(msgs) > 0 {
			in.proc.Invoke(msgs)
		} else {
			return
		}
	}
}

// Start starts the inbox with the specified Processer.
func (in *Inbox) Start(proc Processer) {
	// transition to "starting" and then "idle" to ensure no race condition on in.proc
	if atomic.CompareAndSwapInt32(&in.procStatus, stopped, starting) {
		in.proc = proc
		atomic.SwapInt32(&in.procStatus, idle)
		in.schedule()
	}
}

// Stop stops the inbox and sets its status to stopped.
func (in *Inbox) Stop() error {
	atomic.StoreInt32(&in.procStatus, stopped)
	return nil
}
