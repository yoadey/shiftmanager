package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	lockKeyExpireReservations int64 = 4711001
	lockKeySendReminders      int64 = 4711002
	lockKeyNotifications      int64 = 4711003
)

// ReservationExpirer is the subset of the registration usecase needed to expire
// stale reservations.
type ReservationExpirer interface {
	ExpireReservations(ctx context.Context) (int, error)
}

// ReminderSender sends shift reminder emails for upcoming shifts.
type ReminderSender interface {
	SendDueReminders(ctx context.Context) (int, error)
}

// NotificationRunner runs the daily notification jobs.
type NotificationRunner interface {
	RunDailyNotifications(ctx context.Context) (int, error)
}

// Scheduler runs background maintenance jobs on fixed intervals. When connected
// to PostgreSQL it uses advisory locks so only one replica runs each job at a
// time. On SQLite (test mode) it runs jobs directly without locking.
type Scheduler struct {
	db            *sql.DB
	expirer       ReservationExpirer
	reminders     ReminderSender
	notifications NotificationRunner
	log           zerolog.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new Scheduler. db may be nil to disable advisory locking
// (jobs still run, but without distributed coordination — suitable for
// single-instance deployments and test mode).
func New(db *sql.DB, expirer ReservationExpirer, reminders ReminderSender, notifications NotificationRunner, log zerolog.Logger) *Scheduler {
	return &Scheduler{
		db:            db,
		expirer:       expirer,
		reminders:     reminders,
		notifications: notifications,
		log:           log.With().Str("component", "scheduler").Logger(),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.run(ctx, "expire-reservations", 15*time.Minute, lockKeyExpireReservations, s.expireReservations)
	s.run(ctx, "send-reminders", time.Hour, lockKeySendReminders, s.sendReminders)
	s.run(ctx, "notifications", 24*time.Hour, lockKeyNotifications, s.runNotifications)

	s.log.Info().Msg("scheduler started")
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	s.log.Info().Msg("scheduler stopped")
}

func (s *Scheduler) run(ctx context.Context, name string, interval time.Duration, lockKey int64, job func(context.Context) (int, error)) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.execute(ctx, name, lockKey, job)
			}
		}
	}()
}

func (s *Scheduler) execute(ctx context.Context, name string, lockKey int64, job func(context.Context) (int, error)) {
	// When a *sql.DB is available and supports PostgreSQL advisory locks, use
	// them to ensure only one replica runs the job. Otherwise run directly.
	if s.db != nil {
		locked, err := s.tryAdvisoryLock(ctx, lockKey)
		if err != nil {
			// Advisory lock not supported (e.g. SQLite); run without coordination.
			s.log.Debug().Str("job", name).Msg("advisory lock unavailable, running without lock")
		} else if !locked {
			return // another instance holds the lock
		} else {
			defer func() { _ = s.releaseAdvisoryLock(context.Background(), lockKey) }()
		}
	}

	n, err := job(ctx)
	if err != nil {
		s.log.Error().Err(err).Str("job", name).Msg("job failed")
		return
	}
	s.log.Info().Str("job", name).Int("affected", n).Msg("job completed")
}

func (s *Scheduler) tryAdvisoryLock(ctx context.Context, key int64) (bool, error) {
	var locked bool
	err := s.db.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, key).Scan(&locked)
	if err != nil {
		return false, fmt.Errorf("advisory lock: %w", err)
	}
	return locked, nil
}

func (s *Scheduler) releaseAdvisoryLock(ctx context.Context, key int64) error {
	_, err := s.db.ExecContext(ctx, `SELECT pg_advisory_unlock($1)`, key)
	return err
}

func (s *Scheduler) expireReservations(ctx context.Context) (int, error) {
	if s.expirer == nil {
		return 0, nil
	}
	return s.expirer.ExpireReservations(ctx)
}

func (s *Scheduler) sendReminders(ctx context.Context) (int, error) {
	if s.reminders == nil {
		return 0, nil
	}
	return s.reminders.SendDueReminders(ctx)
}

func (s *Scheduler) runNotifications(ctx context.Context) (int, error) {
	if s.notifications == nil {
		return 0, nil
	}
	return s.notifications.RunDailyNotifications(ctx)
}
