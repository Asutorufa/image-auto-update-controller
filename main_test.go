package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLisimage(t *testing.T) {
	c, err := newContainerd("k8s.io", "/run/k0s/containerd.sock")
	require.NoError(t, err)

	igs, err := c.Client.ListImages(context.Background())
	require.NoError(t, err)

	for _, ig := range igs {
		t.Log(ig.Name())
	}
}

func TestK8s(t *testing.T) {
	c, err := newKubernetes()
	require.NoError(t, err)

	// cc, err := newContainerd("k8s.io", "/run/k0s/containerd.sock")
	// require.NoError(t, err)

	pods, err := c.ListPods(context.Background())
	require.NoError(t, err)

	t.Log(pods)

	t.Run("remove pod", func(t *testing.T) {
		t.Log(c.RemovePods(t.Context(), PodEntry{
			Name:      "grafana-0",
			Namespace: "metrics",
		}))
	})

	// for _, p := range pods {
	// 	t.Log(p.Name, p.Image)
	// 	_, err := cc.GetDigest(context.Background(), p.Image[0])
	// 	require.NoError(t, err)
	// }
}
