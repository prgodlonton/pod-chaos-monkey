package main

import (
	"context"
	"cp2/cli/internal/commands/monkeys"
	"cp2/cli/internal/kubernetes/pods"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	rootCommand *cobra.Command
)

func init() {
	rootCommand = &cobra.Command{
		Use:     "pod-chaos-monkey <namespace>",
		Short:   "Runs the pod resiliency tests",
		Example: "pod-chaos-monkey workloads --local --selector app=nginx,env=dev --interval 10s",
		Args:    cobra.ExactArgs(1),
		RunE:    command,
	}
	flags := rootCommand.Flags()
	flags.Bool("local", false, "flags whether this program running locally")
	flags.String("selector", "", "restricts the test scope to the pods matching the selector")
	flags.Duration("interval", 10*time.Second, "specifies time interval between pod deletions")
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}

func command(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	local, err := flags.GetBool("local")
	if err != nil {
		return err
	}
	selector, err := flags.GetString("selector")
	if err != nil {
		return err
	}
	interval, err := flags.GetDuration("interval")
	if err != nil {
		return err
	}

	cs, err := newClientset(local)
	if err != nil {
		return err
	}
	deleter, err := monkeys.NewSaboteur(pods.NewClient(cs, args[0]), interval)
	if err != nil {
		return fmt.Errorf("cannot create tester: %w", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cfn := context.WithCancel(context.Background())
	defer cfn()

	errs := make(chan error)
	go func() {
		if err := deleter.Havoc(ctx, selector); err != nil {
			errs <- fmt.Errorf("cannot start tester: %w", err)
		}
		close(errs)
	}()

	select {
	case <-signals:
		cfn()
	case err := <-errs:
		return err
	}
	return nil
}

func newClientset(local bool) (*kubernetes.Clientset, error) {
	if local {
		return newLocalClientset()
	}
	return newInClusterClientset()
}

func newLocalClientset() (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch local cluster config: %w", err)
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot build kubernetes client: %v\n", err)
	}
	return cs, nil
}

func newInClusterClientset() (*kubernetes.Clientset, error) {
	c, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot fetch cluster config: %w", err)
	}
	cs, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, fmt.Errorf("cannot get k8 client: %w", err)
	}
	return cs, nil
}
