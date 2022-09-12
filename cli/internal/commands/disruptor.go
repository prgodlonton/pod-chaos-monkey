package commands

import (
	"context"
	"cp2/cli/internal/kubernetes/pods"
	"errors"
	"math/rand"
	"time"
)

var (
	NegativePeriodErr = errors.New("period cannot be negative")
)

// PodDisruptor deletes a single victim pod from a list of pods at a regular interval.
type PodDisruptor struct {
	client pods.DeleterLister
	period time.Duration
}

// NewPodDisruptor creates a new instance of PodDisruptor
func NewPodDisruptor(client pods.DeleterLister, period time.Duration) (*PodDisruptor, error) {
	if period < 0 {
		return nil, NegativePeriodErr
	}
	return &PodDisruptor{
		client: client,
		period: period,
	}, nil
}

// Disrupt selects a victim pod from a list to be deleted at a regular interval
func (c *PodDisruptor) Disrupt(ctx context.Context, selector string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.period):
			list, err := c.client.List(ctx, selector)
			if err != nil {
				return err
			}
			if len(list) > 0 {
				victim := list[rand.Intn(len(list))]
				if err := c.client.Delete(ctx, victim); err != nil {
					return err
				}
			}
		}
	}
}
