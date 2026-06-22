// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package scheduler builds the cron scheduler. The constructor is pure: it
// registers jobs and returns the *cron.Cron without starting it, leaving
// lifecycle (Start/Stop) to the caller.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Job describes a single scheduled task.
type Job struct {
	Name string // human-readable name, used in logs and errors
	Spec string // cron expression
	Run  func() // work to perform on each tick
}

// New builds a cron scheduler (timezone GMT / Africa/Abidjan) with the given
// jobs registered. It does NOT call Start — the caller owns the lifecycle.
func New(jobs []Job) (*cron.Cron, error) {
	c := cron.New(cron.WithLocation(time.FixedZone("GMT", 0)))
	for _, j := range jobs {
		if _, err := c.AddFunc(j.Spec, j.Run); err != nil {
			return nil, fmt.Errorf("add cron job %q: %w", j.Name, err)
		}
	}
	return c, nil
}

// Wrap adapts a context-aware, error-returning collection function into a
// cron-compatible func() that logs start and failures.
func Wrap(ctx context.Context, name string, fn func(context.Context) error) func() {
	return func() {
		slog.Info("running collection job", "job", name)
		if err := fn(ctx); err != nil {
			slog.Error("collection job failed", "job", name, "err", err)
		}
	}
}
