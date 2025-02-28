package main

import (
	"context"

	containerd "github.com/containerd/containerd/v2/client"
)

type Containerd struct {
	Client *containerd.Client
}

func newContainerd(namespace, address string) (*Containerd, error) {
	client, err := containerd.New(address, containerd.WithDefaultNamespace(namespace))
	if err != nil {
		return nil, err
	}

	return &Containerd{
		Client: client,
	}, nil
}

func (c *Containerd) GetDigest(ctx context.Context, name string) (string, error) {
	ig, err := c.Client.ImageService().Get(ctx, name)
	if err != nil {
		return "", err
	}

	return ig.Target.Digest.String(), nil
}

func (c *Containerd) FetchDigest(ctx context.Context, name string) (string, error) {
	ig, err := c.Client.Fetch(ctx, name)
	if err != nil {
		return "", err
	}

	return ig.Target.Digest.String(), nil
}

func (c *Containerd) Pull(ctx context.Context, name string) error {
	_, err := c.Client.Pull(ctx, name)
	return err
}

func (c *Containerd) RemoveImage(ctx context.Context, name string) error {
	err := c.Client.ImageService().Delete(ctx, name)
	return err
}
