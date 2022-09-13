package main

import (
	"context"
	"cp2/cli/internal/commands"
	"cp2/cli/internal/kubernetes/pods"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		Use:     "pod-chaos-monkey <namespace> [flags]",
		Short:   "Runs the pod resiliency tests",
		Example: "pod-chaos-monkey workloads --local --selector app=nginx,env=dev --interval 10s",
		Args:    cobra.ExactArgs(1),
		RunE:    command,
	}
	flags := rootCommand.Flags()
	flags.Bool("local", false, "flags whether this program is running locally")
	flags.String("selector", "", "restricts the test scope to the pods matching the selector")
	flags.Duration("interval", 10*time.Second, "specifies time interval between pod deletions")
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}

func command(cmd *cobra.Command, args []string) error {
	runningLocally, podSelector, interval, err := extractFlags(cmd.Flags())
	if err != nil {
		return err
	}
	disruptor, err := newPodDisruptor(runningLocally, args[0], interval)
	if err != nil {
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cfn := context.WithCancel(context.Background())
	defer cfn()

	errs := make(chan error, 1)
	go func() {
		if err := disruptor.Disrupt(ctx, podSelector); err != nil {
			errs <- fmt.Errorf("cannot start pod disruptor: %w", err)
		}
		close(errs)
	}()

	select {
	case <-signals:
		cfn()
		return nil
	case err := <-errs:
		return err
	}
}

func extractFlags(flags *pflag.FlagSet) (bool, string, time.Duration, error) {
	local, err := flags.GetBool("local")
	if err != nil {
		return false, "", time.Duration(0), err
	}
	selector, err := flags.GetString("selector")
	if err != nil {
		return false, "", time.Duration(0), err
	}
	interval, err := flags.GetDuration("interval")
	if err != nil {
		return false, "", time.Duration(0), err
	}
	return local, selector, interval, nil
}

func newPodDisruptor(runningLocally bool, namespace string, interval time.Duration) (*commands.PodDisruptor, error) {
	var cs *kubernetes.Clientset
	var err error
	if runningLocally {
		cs, err = newLocalClientset()
	} else {
		cs, err = newInClusterClientset()
	}
	if err != nil {
		return nil, err
	}
	d, err := commands.NewPodDisruptor(
		pods.NewServiceLayerClient(cs, namespace),
		interval,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create pod disruptor: %w", err)
	}
	return d, nil
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
