package monkeys

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

// Saboteur deletes a single pod from the set returned by the pods client at a regular interval.
type Saboteur struct {
	client pods.DeleterLister
	period time.Duration
}

// NewSaboteur creates a new instance of Saboteur
func NewSaboteur(client pods.DeleterLister, period time.Duration) (*Saboteur, error) {
	if period < 0 {
		return nil, NegativePeriodErr
	}
	return &Saboteur{
		client: client,
		period: period,
	}, nil
}

// Havoc starts the tester running concurrently; calls cb if lister or deleter returns an error
func (c *Saboteur) Havoc(ctx context.Context, selector string) error {
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
