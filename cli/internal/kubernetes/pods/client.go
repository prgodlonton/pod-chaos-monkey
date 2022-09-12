package pods

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Deleter interface {
	Delete(ctx context.Context, name string) error
}

type Lister interface {
	List(ctx context.Context, selectors string) ([]string, error)
}

type DeleterLister interface {
	Deleter
	Lister
}

type Option interface {
	Apply(*Client)
}

type Client struct {
	cs        kubernetes.Interface
	namespace string
}

func NewClient(cs kubernetes.Interface, namespace string, opts ...Option) *Client {
	client := &Client{
		cs:        cs,
		namespace: namespace,
	}
	for _, opt := range opts {
		opt.Apply(client)
	}
	return client
}

func (c *Client) Delete(ctx context.Context, name string) error {
	if err := c.cs.CoreV1().Pods(c.namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("cannot delete pod: %w", err)
	}
	return nil
}

func (c *Client) List(ctx context.Context, selectors string) ([]string, error) {
	pods, err := c.cs.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selectors,
		//Limit:         c.limit,
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
