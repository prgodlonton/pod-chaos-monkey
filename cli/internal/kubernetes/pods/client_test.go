//go:build !integration

package pods_test

import (
	"context"
	"cp2/cli/internal/kubernetes/pods"
	"errors"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	kt "k8s.io/client-go/testing"
	"testing"
	"time"
)

const (
	timeoutDur = time.Second * 20
)

func TestDeletePassesCorrectNamespaceAndPodName(t *testing.T) {
	as := assert.New(t)

	var wasCalled bool
	reaction := func(action kt.Action) (bool, runtime.Object, error) {
		defer func() {
			wasCalled = true
		}()

		del, ok := action.(kt.DeleteActionImpl)
		if !ok {
			t.Fatalf("action is not of type DeleteActionImpl")
		}
		as.Equal("namespace", del.Namespace)
		as.Equal("pod-name", del.Name)

		return true, nil, nil
	}

	cs := fake.NewSimpleClientset()
	cs.Fake.PrependReactor("delete", "pods", reaction)

	ctx, cfn := context.WithTimeout(context.Background(), timeoutDur)
	defer cfn()

	cl := pods.NewClient(cs, "namespace")
	as.Nil(cl.Delete(ctx, "pod-name"))

	as.True(wasCalled)
}

func TestDeleteReturnsError(t *testing.T) {
	as := assert.New(t)

	err := errors.New("server error")
	reaction := func(action kt.Action) (bool, runtime.Object, error) {
		return true, nil, err
	}

	cs := fake.NewSimpleClientset()
	cs.Fake.PrependReactor("delete", "pods", reaction)

	ctx, cfn := context.WithTimeout(context.Background(), timeoutDur)
	defer cfn()

	cl := pods.NewClient(cs, "namespace")
	as.ErrorIs(cl.Delete(ctx, "pod-name"), err)
}

func TestListPassesCorrectNamespaceAndSelectors(t *testing.T) {
	as := assert.New(t)

	var wasCalled bool
	reaction := func(action kt.Action) (bool, runtime.Object, error) {
		defer func() {
			wasCalled = true
		}()

		list, ok := action.(kt.ListActionImpl)
		if !ok {
			t.Fatalf("action is not of type ListActionImpl")
		}
		as.Equal("namespace", list.Namespace)
		as.Equal("field1=value1,field2=value2", list.ListRestrictions.Labels.String())

		return true, &v1.PodList{}, nil
	}

	cs := fake.NewSimpleClientset()
	cs.Fake.PrependReactor("list", "pods", reaction)

	ctx, cfn := context.WithTimeout(context.Background(), timeoutDur)
	defer cfn()

	cl := pods.NewClient(cs, "namespace")
	_, _ = cl.List(ctx, "field1=value1,field2=value2")

	as.True(wasCalled)
}

func TestListExtractsPodsNames(t *testing.T) {
	as := assert.New(t)

	reaction := func(action kt.Action) (bool, runtime.Object, error) {
		return true, &v1.PodList{
			Items: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
				},
			},
		}, nil
	}

	cs := fake.NewSimpleClientset()
	cs.Fake.PrependReactor("list", "pods", reaction)

	ctx, cfn := context.WithTimeout(context.Background(), timeoutDur)
	defer cfn()

	cl := pods.NewClient(cs, "namespace")
	list, err := cl.List(ctx, "")

	as.Contains(list, "pod1")
	as.Contains(list, "pod2")
	as.Nil(err)
}

func TestListReturnsError(t *testing.T) {
	as := assert.New(t)

	listErr := errors.New("api server error")
	reaction := func(action kt.Action) (bool, runtime.Object, error) {
		return true, nil, listErr
	}

	cs := fake.NewSimpleClientset()
	cs.Fake.PrependReactor("list", "pods", reaction)

	ctx, cfn := context.WithTimeout(context.Background(), timeoutDur)
	defer cfn()

	cl := pods.NewClient(cs, "namespace")
	_, err := cl.List(ctx, "field1=value1")

	as.ErrorIs(err, listErr)
}
