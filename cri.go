package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"k8s.io/apimachinery/pkg/types"
	cri "k8s.io/cri-api/pkg/apis"
	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	remote "k8s.io/cri-client/pkg"
	"k8s.io/klog/v2"
)

type Cri struct {
	imageCri   cri.ImageManagerService
	runtimeCri cri.RuntimeService
}

func newCri(address string) (*Cri, error) {
	var tp trace.TracerProvider = noop.NewTracerProvider()
	logger := klog.Background()

	image, err := remote.NewRemoteImageService(address, time.Second*2, tp, &logger)
	if err != nil {
		return nil, err
	}

	runtime, err := remote.NewRemoteRuntimeService(address, time.Second*2, tp, &logger)
	if err != nil {
		return nil, err
	}

	return &Cri{
		imageCri:   image,
		runtimeCri: runtime,
	}, nil
}

func (c *Cri) ListPods() ([]Pod, error) {
	pods, err := c.runtimeCri.ListPodSandbox(context.Background(), &v1.PodSandboxFilter{
		LabelSelector: map[string]string{
			Label: "true",
		},
	})
	if err != nil {
		return nil, err
	}

	var result []Pod
	for _, p := range pods {
		status, err := c.runtimeCri.PodSandboxStatus(context.Background(), p.GetId(), true)
		if err != nil {
			slog.Error("get container status failed", "err", err, "pod", p.GetId())
			continue
		}

		images := []string{}

		fmt.Println(status.ContainersStatuses)

		for _, v := range status.ContainersStatuses {
			fmt.Println(v.Image, v.ImageId, v.ImageRef)
			images = append(images, v.ImageId)
		}

		result = append(result, Pod{
			Name:      p.GetMetadata().GetName(),
			Namepsace: p.GetMetadata().GetNamespace(),
			Uid:       types.UID(p.GetMetadata().GetUid()),
			Image:     images,
		})
	}

	return result, nil
}

func (c *Cri) ImageId(ctx context.Context, image string) (string, error) {
	imageStatus, err := c.imageCri.ImageStatus(ctx, &v1.ImageSpec{
		Image: image,
	}, true)
	if err != nil {
		return "", err
	}

	return imageStatus.GetImage().GetId(), nil
}

func (c *Cri) PullImage(ctx context.Context, image string) (string, error) {
	return c.imageCri.PullImage(ctx, &v1.ImageSpec{
		Image: image,
	}, nil, nil)
}

func (c *Cri) RemoveUnusedImages(ctx context.Context) error {
	images, err := c.imageCri.ListImages(ctx, nil)
	if err != nil {
		return err
	}

	ids := map[string]struct{}{}
	for _, image := range images {
		// pinned images can't be deleted
		if image.Pinned {
			continue
		}

		ids[image.GetId()] = struct{}{}
	}

	containers, err := c.runtimeCri.ListContainers(ctx, nil)
	if err != nil {
		return err
	}

	for _, container := range containers {
		imageStatus, err := c.imageCri.ImageStatus(ctx, container.GetImage(), true)
		if err != nil {
			slog.Error("get image status failed", "err", err, "image", container.GetImage())
			continue
		}

		delete(ids, imageStatus.GetImage().GetId())
	}

	for id := range ids {
		imageStatus, err := c.imageCri.ImageStatus(ctx, &v1.ImageSpec{Image: id}, true)
		if err != nil {
			slog.Error("get image status failed", "err", err, "image", id)
			continue
		}

		if imageStatus.Image == nil {
			slog.Error("image not found", "image", id)
			continue
		}

		err = c.RemoveImage(ctx, id)
		if err != nil {
			slog.Error("remove image failed", "err", err, "image", id)
			continue
		}

		if len(imageStatus.Image.RepoTags) == 0 {
			// RepoTags is nil when pulling image by repoDigest,
			// so print deleted using that instead.
			for _, repoDigest := range imageStatus.Image.RepoDigests {
				fmt.Printf("Deleted: %s\n", repoDigest)
			}

			return nil
		}

		for _, repoTag := range imageStatus.Image.RepoTags {
			fmt.Printf("Deleted: %s\n", repoTag)
		}
	}

	return nil
}

func (c *Cri) RemoveImage(ctx context.Context, id string) error {
	return c.imageCri.RemoveImage(ctx, &v1.ImageSpec{
		Image: id,
	})
}
