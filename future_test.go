package future_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/knzm/future"
)

func TestFutureGet(t *testing.T) {
	ctx := context.Background()
	f := future.New(ctx, func(s future.Sender) error {
		return s.Send("done", 0)
	})
	defer f.Cancel()
	v, err := f.Get()
	if err != nil {
		t.Errorf("Got an error: %v", err)
	}
	if s, ok := v.(string); !ok || s != "done" {
		t.Errorf("Got an unknown value: %v", v)
	}
}

func TestFutureGetCancel(t *testing.T) {
	ctx := context.Background()
	f := future.New(ctx, func(s future.Sender) error {
		time.Sleep(1 * time.Second)
		return s.Send("done", 0)
	})
	f.Cancel()
	_, err := f.Get()
	if err == nil {
		t.Logf("Succeeded before cancellation")
	} else if err.Error() != "CANCELED" {
		t.Errorf("Got an unknown error: %v", err)
	}
}

func TestFutureChannel(t *testing.T) {
	ctx := context.Background()
	f := future.New(ctx, func(s future.Sender) error {
		for i := 0; i < 5; i++ {
			if err := s.Send(i, 0); err != nil {
				return err
			}
		}
		return nil
	})
	defer f.Cancel()

	var ret []future.Value
	for v := range f.GetChannel() {
		ret = append(ret, v)
	}

	err := f.GetLastError()
	if err != nil {
		t.Errorf("Got an error: %v", err)
	}

	expected := []future.Value{0, 1, 2, 3, 4}
	if !reflect.DeepEqual(ret, expected) {
		t.Errorf("Got an unknown value: %v", ret)
	}
}

func TestFutureChannelCancel(t *testing.T) {
	ctx := context.Background()
	f := future.New(ctx, func(s future.Sender) error {
		for i := 0; i < 5; i++ {
			if err := s.Send(i, 0); err != nil {
				return err
			}
		}
		return nil
	})
	defer f.Cancel()

	var ret []future.Value
	i := 0
	for v := range f.GetChannel() {
		ret = append(ret, v)
		if i > 1 {
			f.Cancel()
			break
		}
		i++
	}

	err := f.GetLastError()
	if err == nil {
		t.Logf("Succeeded before cancellation")
	} else if err.Error() != "CANCELED" {
		t.Errorf("Got an unknown error: %v", err)
	}

	expected := []future.Value{0, 1, 2}
	if !reflect.DeepEqual(ret, expected) {
		t.Errorf("Got an unknown value: %v", ret)
	}
}
