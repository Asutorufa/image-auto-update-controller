package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Kubernetes struct {
	Client *kubernetes.Clientset
}

func newKubernetes() (*Kubernetes, error) {
	var apiConfig *clientcmdapi.Config
	var err error
	apiConfig, _ = clientcmd.NewDefaultClientConfigLoadingRules().Load()
	var config *rest.Config
	if apiConfig != nil {
		config, err = clientcmd.NewDefaultClientConfig(*apiConfig, nil).ClientConfig()
	}
	if config == nil || err != nil {
		fmt.Println(err)
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		Client: cli,
	}, nil
}

type Pod struct {
	Name      string
	Uid       types.UID
	Namepsace string
	Image     []string
}

func (k *Kubernetes) ListPods(ctx context.Context) ([]Pod, error) {
	ls, err := k.Client.CoreV1().Pods("").List(ctx, v1.ListOptions{
		LabelSelector: Label + "=true",
	})
	if err != nil {
		return nil, err
	}

	var names []Pod
	for _, l := range ls.Items {
		var images []string
		for _, c := range l.Spec.InitContainers {
			images = append(images, c.Image)
		}

		for _, c := range l.Spec.Containers {
			images = append(images, c.Image)
		}

		for _, c := range l.Spec.EphemeralContainers {
			images = append(images, c.Image)
		}

		names = append(names, Pod{
			Name:      l.Name,
			Namepsace: l.Namespace,
			Uid:       l.UID,
			Image:     images,
		})
	}

	return names, nil
}

func (k *Kubernetes) RemovePods(ctx context.Context, pe PodEntry) error {
	event, err := k.Client.CoreV1().Pods(pe.Namespace).Watch(ctx, v1.ListOptions{
		LabelSelector: Label + "=true",
	})
	if err != nil {
		return fmt.Errorf("watch pods: %w", err)
	}
	defer event.Stop()

	wait := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				wait <- ctx.Err()
				return
			case e, ok := <-event.ResultChan():
				if !ok {
					wait <- fmt.Errorf("watch channel closed")
					return
				}

				pod, ok := e.Object.(*corev1.Pod)
				if !ok {
					continue
				}

				if pod.Name != pe.Name || pod.Namespace != pe.Namespace {
					continue
				}

				if e.Type == watch.Deleted {
					wait <- nil
					return
				}
			}
		}
	}()

	err = k.Client.CoreV1().Pods(pe.Namespace).Delete(ctx, pe.Name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return <-wait
}

const (
	Label = "asutorufa.github.io/image-updater"
)
