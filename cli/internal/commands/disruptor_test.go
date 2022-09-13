//go:build !integration

package commands_test

import (
	"context"
	"cp2/cli/internal/commands"
	"cp2/cli/internal/kubernetes/pods"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testTimeout = time.Second
)

type MockDeleter struct {
	Callback  func(ctx context.Context, name string) error
	WasCalled bool
}

func (md *MockDeleter) Delete(ctx context.Context, name string) error {
	defer func() {
		md.WasCalled = true
	}()
	if md.Callback != nil {
		return md.Callback(ctx, name)
	}
	return nil
}

var _ pods.Deleter = &MockDeleter{}

type MockLister struct {
	Callback  func(ctx context.Context, selector string) ([]string, error)
	WasCalled bool
}

func (ml *MockLister) List(ctx context.Context, selector string) ([]string, error) {
	defer func() {
		ml.WasCalled = true
	}()
	if ml.Callback != nil {
		return ml.Callback(ctx, selector)
	}
	return []string{}, nil
}

var _ pods.Lister = &MockLister{}

type MockDeleterLister struct {
	MockDeleter
	MockLister
}

var _ pods.DeleterLister = &MockDeleterLister{}

func TestPodDisruptorDeletesPodTakenFromList(t *testing.T) {
	as := assert.New(t)

	ctx, cfn := context.WithCancel(context.Background())
	defer cfn()

	dl := &MockDeleterLister{
		MockDeleter: MockDeleter{
			Callback: func(_ context.Context, name string) error {
				defer func() {
					cfn()
				}()
				as.Equal("pod1", name)
				return nil
			},
		},
		MockLister: MockLister{
			Callback: func(_ context.Context, _ string) ([]string, error) {
				return []string{
					"pod1",
				}, nil
			},
		},
	}
	disruptor, err := commands.NewPodDisruptor(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create disruptor: %v", err)
	}

	as.Nil(disruptor.Disrupt(ctx, ""))
}

func TestPodDisruptorPassesSelectorToLister(t *testing.T) {
	as := assert.New(t)

	ctx, cfn := context.WithCancel(context.Background())
	defer cfn()

	dl := &MockDeleterLister{
		MockLister: MockLister{
			Callback: func(_ context.Context, selector string) ([]string, error) {
				defer func() {
					cfn()
				}()
				as.Equal("field1=value1", selector)
				return []string{}, nil
			},
		},
	}
	disruptor, err := commands.NewPodDisruptor(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create disruptor: %v", err)
	}

	as.Nil(disruptor.Disrupt(ctx, "field1=value1"))
	as.True(dl.MockLister.WasCalled)
}

func TestPodDisruptorDoesNothingWhenPodListIsEmpty(t *testing.T) {
	as := assert.New(t)

	ctx, cfn := context.WithTimeout(context.Background(), testTimeout)
	defer cfn()

	dl := &MockDeleterLister{
		MockDeleter: MockDeleter{
			Callback: func(_ context.Context, name string) error {
				t.Fatalf("pod delete unexpectedly called")
				return nil
			},
		},
		MockLister: MockLister{
			Callback: func(_ context.Context, _ string) ([]string, error) {
				return []string{}, nil
			},
		},
	}
	disruptor, err := commands.NewPodDisruptor(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create disruptor: %v", err)
	}

	as.Nil(disruptor.Disrupt(ctx, ""))
	as.False(dl.MockDeleter.WasCalled)
	as.True(dl.MockLister.WasCalled)
}

func TestPodDisruptorReturnsWhenPassedCanceledContext(t *testing.T) {
	ctx, cfn := context.WithTimeout(context.Background(), testTimeout)
	cfn()

	disruptor, err := commands.NewPodDisruptor(&MockDeleterLister{}, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create disruptor: %v", err)
	}

	assert.Nil(t, disruptor.Disrupt(ctx, ""))
}

func TestPodDisruptorReturnsErrorWhenListerReturnsError(t *testing.T) {
	expectedErr := errors.New("intentional error")
	dl := &MockDeleterLister{
		MockLister: MockLister{
			Callback: func(_ context.Context, _ string) ([]string, error) {
				return []string{}, expectedErr
			},
		},
	}
	disruptor, err := commands.NewPodDisruptor(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create disruptor: %v", err)
	}

	err = disruptor.Disrupt(context.Background(), "")

	as := assert.New(t)
	as.ErrorIs(err, expectedErr)
	as.True(dl.MockLister.WasCalled)
}

func TestPodDisruptorReturnsErrorWhenDeleterReturnsError(t *testing.T) {
	expectedErr := errors.New("intentional error")
	dl := &MockDeleterLister{
		MockDeleter: MockDeleter{
			Callback: func(_ context.Context, _ string) error {
				return expectedErr
			},
		},

		MockLister: MockLister{
			Callback: func(_ context.Context, actualSelector string) ([]string, error) {
				return []string{
					"pod1",
				}, nil
			},
		},
	}
	disruptor, err := commands.NewPodDisruptor(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create disruptor: %v", err)
	}

	err = disruptor.Disrupt(context.Background(), "")

	as := assert.New(t)
	as.ErrorIs(err, expectedErr)
	as.True(dl.MockDeleter.WasCalled)
	as.True(dl.MockLister.WasCalled)
}
