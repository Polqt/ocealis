package service

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Scheduler struct {
	cron  *cron.Cron
	drift DriftService
	log   *zap.Logger
	ctx   context.Context
}

func NewScheduler(drift DriftService, log *zap.Logger) *Scheduler {
	return &Scheduler{
		cron:  cron.New(),
		drift: drift,
		log:   log,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.ctx = ctx

	// Existing bottle releases and scheduled releases are handled in the same way -
	// they just have different release times.
	// So we can use the same scheduler to check for both.
	s.cron.AddFunc("*/15 * * * *", func() {
		if err := s.drift.Tick(s.ctx); err != nil {
			s.log.Error("drift tick failed", zap.Error(err))
		}
	})

	// This is a safety net to catch any scheduled releases that might
	// be missed if the server restarts between ticks.
	// It runs every minute and checks for any scheduled releases that are due.
	s.cron.AddFunc("* * * * *", func() {
		if err := s.drift.ReleaseScheduled(s.ctx); err != nil {
			s.log.Error("scheduled release failed", zap.Error(err))
		}
	})

	s.cron.Start()
	s.log.Info("scheduler started - drift tick every 15 mins.")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	s.log.Info("scheduler stopped")
}
