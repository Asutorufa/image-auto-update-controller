package main

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

func main() {
	kcli, err := newKubernetes()
	if err != nil {
		panic(err)
	}

	criAddress := os.Getenv("CRI_ENDPOINT_ADDRESS")
	if criAddress == "" {
		criAddress = "/run/k0s/containerd.sock"
	}

	duration := time.Hour * 24

	if d, _ := strconv.ParseUint(os.Getenv("UPDATE_TICKER"), 10, 64); d > 0 {
		duration = time.Hour * time.Duration(d)
	}

	c, err := newCri(criAddress)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(duration)

	for range ticker.C {
		if err := Check(kcli, c); err != nil {
			slog.Error("check failed", "err", err)
		}
	}
}
