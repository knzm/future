package future

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/golang/glog"
)

var (
	errCanceled = errors.New("CANCELED")
	errTimeout  = errors.New("TIMEOUT")
)

type Value interface{}

type Sender interface {
	Send(Value, time.Duration) error
}

type sender struct {
	ctx context.Context
	ch  chan Value
}

func (s *sender) send(v Value) error {
	select {
	case <-s.ctx.Done():
		glog.V(1).Info("Sender.Send() aborted")
		return errCanceled
	case s.ch <- v:
		return nil
	}
}

func (s *sender) sendAndWait(v Value, timeout time.Duration) error {
	select {
	case <-s.ctx.Done():
		glog.V(1).Info("Sender.Send() aborted")
		return errCanceled
	case <-time.After(timeout):
		glog.V(1).Info("Sender.Send() timed out")
		return errTimeout
	case s.ch <- v:
		return nil
	}
}

func (s *sender) Send(v Value, timeout time.Duration) error {
	if timeout == 0 {
		return s.send(v)
	}

	return s.sendAndWait(v, timeout)
}

type Future interface {
	Get() (Value, error)
	GetChannel() <-chan Value
	GetLastError() error
	Cancel()
}

type future struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	ch     chan Value
	error  error
	mutex  *sync.RWMutex
}

type FutureFunc func(sender Sender) error

func New(ctx context.Context, fun FutureFunc) Future {
	ctx, cancel := context.WithCancel(ctx)
	wg := new(sync.WaitGroup)
	mutex := new(sync.RWMutex)
	ch := make(chan Value)
	future := &future{ctx, cancel, wg, ch, nil, mutex}
	future.invoke(fun)
	return future
}

func (f *future) invoke(fun FutureFunc) {
	go func() {
		defer f.wg.Done()
		defer close(f.ch)
		sender := &sender{f.ctx, f.ch}
		err := fun(sender)
		f.setError(err)
	}()
	f.wg.Add(1)
}

func (f *future) Get() (Value, error) {
	select {
	case <-f.ctx.Done():
		glog.V(1).Info("future.Get() aborted")
		return nil, errCanceled
	case v := <-f.ch:
		err := f.GetLastError()
		return v, err
	}
}

func (f *future) GetChannel() <-chan Value {
	return f.ch
}

func (f *future) GetLastError() error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.error
}

func (f *future) setError(err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.error = err
}

func (f *future) Cancel() {
	f.cancel()
	for v := range f.ch {
		glog.V(1).Info("skip extra data: ", v)
	}
	f.wg.Wait()
}
