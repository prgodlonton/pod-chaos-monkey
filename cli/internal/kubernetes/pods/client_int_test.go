//go:build integration

package pods_test

import (
	"context"
	"cp2/cli/internal/kubernetes/pods"
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
	"time"
)

const (
	namespace  = "workloads"
	pollingDur = time.Second * 2
	timeoutDur = time.Second * 20
)

var (
	cs *kubernetes.Clientset
)

func TestMain(m *testing.M) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot run tests: %v\n", err)
		os.Exit(1)
	}
	cs, err = kubernetes.NewForConfig(config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot run tests: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestListReturnsPodsMatchingSelector(t *testing.T) {
	ctx, cfn := context.WithTimeout(context.Background(), timeoutDur)
	defer cfn()

	client := pods.NewServiceLayerClient(cs, namespace)
	list, err := client.List(ctx, "app=nginx,env=dev")

	as := assert.New(t)
	as.Len(list, 3)
	as.Nil(err)
}

func TestDeleteTerminatesVictimPod(t *testing.T) {
	ctx, cfn := context.WithTimeout(context.Background(), timeoutDur)
	defer cfn()

	selectors := "app=nginx,env=dev"
	client := pods.NewServiceLayerClient(cs, namespace)
	list, err := client.List(ctx, selectors)

	as := assert.New(t)
	as.Len(list, 3)
	as.Nil(err)

	if len(list) != 3 {
		t.Fatalf("expected 3 pods running; found %d", len(list))
	}
	victim := list[0]
	as.Nil(client.Delete(ctx, victim))

	for {
		select {
		case <-ctx.Done():
			t.Fatal("test timed out")
		case <-time.After(pollingDur):
			list, _ = client.List(ctx, selectors)
			if !contains(list, victim) {
				return
			}
		}
	}
}

func contains(slice []string, element string) bool {
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}
