package common

import (
	"context"
)

type Closer struct {
	ch chan error
}

// Spawns a closer for the given resource. The closer funtion will be called
// - When the context finalizes
// - When the `Close` method is called
// Guarantees that the close method is called only once.
func SpawnCloser[Resource any](ctx context.Context, resource Resource, resourceCloser func(Resource) error) Closer {
	ch := make(chan error)
	go func() {
		select {
		case <-ctx.Done():
			err := resourceCloser(resource)
			<-ch
			ch <- err
		case <-ch:
			ch <- resourceCloser(resource)
		}
	}()

	return Closer{ch}
}

// Notifies the Closer to close it's associated resource, if it hasn't closed it yet.
// Returns the error, if any, of executing the closer function.
func (c Closer) Close() error {
	c.ch <- nil
	return <-c.ch
}
