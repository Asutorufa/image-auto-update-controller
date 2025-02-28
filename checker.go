package main

import (
	"context"
	"iter"
	"log/slog"
	"slices"

	"k8s.io/apimachinery/pkg/types"
)

func Check(k *Kubernetes, cri *Cri) error {
	pods, err := k.ListPods(context.Background())
	if err != nil {
		return err
	}

	podsByImage := podsByImage(pods)

	var needRestart set[PodEntry]
	var oldPodsDigests set[string]
	for image, pods := range podsByImage {
		od, err := cri.ImageId(context.Background(), image)
		if err != nil {
			slog.Error("get digest failed", "err", err, "image", image)
			continue
		}

		nd, err := cri.PullImage(context.Background(), image)
		if err != nil {
			slog.Error("fetch digest failed", "err", err, "image", image)
			continue
		}

		if nd == od {
			continue
		}

		slog.Info("pull image", "image", image, "old", od, "new", nd)

		oldPodsDigests.add(od)
		needRestart.copyFrom(pods)
	}

	for v := range needRestart.Range() {
		slog.Info("restart pod", "name", v)

		if err := k.RemovePods(context.Background(), v); err != nil {
			slog.Error("remove pod failed", "err", err, "name", v)
		}
	}

	slog.Info("remove old images", "ids", slices.Collect(oldPodsDigests.Range()))

	for v := range oldPodsDigests.Range() {
		if err := cri.RemoveImage(context.Background(), v); err != nil {
			slog.Error("remove image failed", "err", err, "image", v)
		}
	}

	slog.Info("remove unused images")

	err = cri.RemoveUnusedImages(context.Background())
	if err != nil {
		slog.Error("remove unused images failed", "err", err)
	}

	return nil
}

func podsByImage(pods []Pod) map[string]*set[PodEntry] {
	maps := make(map[string]*set[PodEntry])

	for _, p := range pods {
		for _, image := range p.Image {
			s, ok := maps[image]
			if !ok {
				s = &set[PodEntry]{}
				maps[image] = s
			}

			s.add(PodEntry{
				Name:      p.Name,
				Namespace: p.Namepsace,
				Uid:       p.Uid,
			})
		}
	}

	return maps
}

type PodEntry struct {
	Name      string
	Namespace string
	Uid       types.UID
}

type set[T comparable] struct {
	m map[T]struct{}
}

func (s *set[T]) add(v T) {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}

	s.m[v] = struct{}{}
}

func (s *set[T]) Range() iter.Seq[T] {
	return func(yield func(T) bool) {
		for k := range s.m {
			if !yield(k) {
				return
			}
		}
	}
}

func (s *set[T]) copyFrom(ss *set[T]) {
	for v := range ss.Range() {
		s.add(v)
	}
}
