//go:build !integration

package monkeys_test

import (
	"context"
	"cp2/cli/internal/commands/monkeys"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

type MockDeleterLister struct {
	MockDeleter
	MockLister
}

func TestPodDeleterReturnsImmediatelyWhenPassedCanceledContext(t *testing.T) {
	deleter, err := monkeys.NewSaboteur(&MockDeleterLister{}, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create pod deleter: %v", err)
	}

	ctx, cfn := context.WithCancel(context.Background())
	cfn()

	assert.Nil(t, deleter.Havoc(ctx, ""))
}

func TestPodDeleterCallsListerWithSelector(t *testing.T) {
	as := assert.New(t)
	expectedSelector := "app=nginx"

	ctx, cfn := context.WithCancel(context.Background())
	defer cfn()

	dl := &MockDeleterLister{
		MockLister: MockLister{
			Callback: func(_ context.Context, actualSelector string) ([]string, error) {
				defer func() {
					cfn()
				}()
				as.Equal(expectedSelector, actualSelector)
				return []string{}, nil
			},
		},
	}
	deleter, err := monkeys.NewSaboteur(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create pod deleter: %v", err)
	}

	as.Nil(deleter.Havoc(ctx, expectedSelector))
	as.True(dl.MockLister.WasCalled)
}

func TestPodDeleterReturnsErrorWhenListerReturnsError(t *testing.T) {
	dl := &MockDeleterLister{
		MockLister: MockLister{
			Callback: func(_ context.Context, actualSelector string) ([]string, error) {
				return []string{}, errors.New("intentional error")
			},
		},
	}
	deleter, err := monkeys.NewSaboteur(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create pod deleter: %v", err)
	}

	as := assert.New(t)
	as.NotNil(deleter.Havoc(context.Background(), ""))
	as.True(dl.MockLister.WasCalled)
}

func TestPodDeleterDoesNothingWhenListerReturnsEmptyList(t *testing.T) {
	calls := make(chan struct{})
	dl := &MockDeleterLister{
		MockDeleter: MockDeleter{
			Callback: func(_ context.Context, _ string) error {
				t.Fatalf("delete unexpectedly called")
				return nil
			},
		},
		MockLister: MockLister{
			Callback: func(_ context.Context, actualSelector string) ([]string, error) {
				defer func() {
					calls <- struct{}{}
				}()
				return []string{}, nil
			},
		},
	}
	deleter, err := monkeys.NewSaboteur(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create pod deleter: %v", err)
	}

	ctx, cfn := context.WithCancel(context.Background())
	defer cfn()

	go func() {
		assert.Nil(t, deleter.Havoc(ctx, ""))
	}()

	for i := 0; i < 10; {
		<-calls
		i++
	}
	close(calls)
}

func TestPodDeleterReturnsErrorWhenDeleterReturnsError(t *testing.T) {
	as := assert.New(t)
	dl := &MockDeleterLister{
		MockDeleter: MockDeleter{
			Callback: func(_ context.Context, _ string) error {
				return errors.New("intentional error")
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
	deleter, err := monkeys.NewSaboteur(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create pod deleter: %v", err)
	}

	as.NotNil(deleter.Havoc(context.Background(), ""))
	as.True(dl.MockDeleter.WasCalled)
}

func TestPodDeleterSelectsPodForDeletionFromList(t *testing.T) {
	as := assert.New(t)

	ctx, cfn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cfn()

	pods := []string{
		"pod1",
		"pod2",
	}
	dl := &MockDeleterLister{
		MockDeleter: MockDeleter{
			Callback: func(_ context.Context, name string) error {
				defer func() {
					cfn()
				}()
				as.Contains(pods, name)
				return nil
			},
		},

		MockLister: MockLister{
			Callback: func(_ context.Context, actualSelector string) ([]string, error) {
				return pods, nil
			},
		},
	}
	deleter, err := monkeys.NewSaboteur(dl, time.Duration(0))
	if err != nil {
		t.Fatalf("cannot create pod deleter: %v", err)
	}

	as.Nil(deleter.Havoc(ctx, ""))
	as.True(dl.MockDeleter.WasCalled)
}
