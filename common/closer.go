package common

import (
	"context"
)

type Closer struct {
	ch chan error
}

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

func (c Closer) Close() error {
	c.ch <- nil
	return <-c.ch
}
