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
}

func NewScheduler(drift DriftService, log *zap.Logger) *Scheduler {
	return &Scheduler{
		cron:  cron.New(),
		drift: drift,
		log:   log,
	}
}

func (s *Scheduler) Start() {
	_, err := s.cron.AddFunc("*/15 * * * *", func() {
		if err := s.drift.Tick(context.Background()); err != nil {
			s.log.Error("drift tick failed", zap.Error(err))
		}
	})
	if err != nil {
		s.log.Fatal("failed to register drift job", zap.Error(err))
	}
	s.cron.Start()
	s.log.Info("scheduler started - drift tick every 15 mins.")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	s.log.Info("scheduler stopped")
}
