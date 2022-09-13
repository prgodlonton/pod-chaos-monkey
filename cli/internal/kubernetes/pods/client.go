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

// Lister list all pods matching the podSelector.
type Lister interface {
	List(ctx context.Context, podSelector string) ([]string, error)
}

// DeleterLister compound interface combining Deleter and Lister
type DeleterLister interface {
	Deleter
	Lister
}

// PodServiceLayer provides delete and list pods functions.
// These wrap the K8 CoreV1 API calls.
type PodServiceLayer struct {
	clientset kubernetes.Interface
	namespace string
}

// NewServiceLayerClient creates a new instance of the PodServiceLayer for the given namespace
func NewServiceLayerClient(cs kubernetes.Interface, namespace string) *PodServiceLayer {
	return &PodServiceLayer{
		clientset: cs,
		namespace: namespace,
	}
}

// Delete deletes the pod with given name from the namespace
func (c *PodServiceLayer) Delete(ctx context.Context, podName string) error {
	if err := c.clientset.CoreV1().Pods(c.namespace).Delete(ctx, podName, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("cannot delete pod: %w", err)
	}
	return nil
}

// List lists all pods within the namespace matching the selector
func (c *PodServiceLayer) List(ctx context.Context, podSelector string) ([]string, error) {
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: podSelector,
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
