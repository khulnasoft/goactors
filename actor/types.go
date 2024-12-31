package actor

import "sync"

type InternalError struct {
	From string
	Err  error
}

type poisonPill struct {
	wg       *sync.WaitGroup
	graceful bool
}
type (
	Initialized struct{}
	Started     struct{}
	Stopped     struct{}
)
