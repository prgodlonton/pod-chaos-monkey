package pods

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Deleter deletes the pod with name.
type Deleter interface {
	Delete(ctx context.Context, name string) error
}

// Lister list all pods matching the selectors.
type Lister interface {
	List(ctx context.Context, selectors string) ([]string, error)
}

// DeleterLister compound interface combining Deleter and Lister
type DeleterLister interface {
	Deleter
	Lister
}

// Client provides delete and list pods functions.
// These wrap the K8 CoreV1 API calls.
type Client struct {
	cs        kubernetes.Interface
	namespace string
}

// NewClient creates a new instance of the Client for the given namespace
func NewClient(cs kubernetes.Interface, namespace string) *Client {
	return &Client{
		cs:        cs,
		namespace: namespace,
	}
}

// Delete deletes the pod with given name from the namespace
func (c *Client) Delete(ctx context.Context, name string) error {
	if err := c.cs.CoreV1().Pods(c.namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("cannot delete pod: %w", err)
	}
	return nil
}

// List lists all pods within the namespace matching the selector
func (c *Client) List(ctx context.Context, selectors string) ([]string, error) {
	pods, err := c.cs.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selectors,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list pods: %w", err)
	}
	names := make([]string, 0)
	for _, pod := range pods.Items {
		names = append(names, pod.Name)
	}
	return names, nil
}
