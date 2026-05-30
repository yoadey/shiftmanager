package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// Advisory lock keys. Each scheduled job uses a distinct key so that only a
// single instance across the cluster runs the job at a time.
const (
	lockKeyExpireReservations int64 = 4711001
	lockKeySendReminders      int64 = 4711002
)

// ReservationExpirer is the subset of the registration usecase needed to expire
// stale reservations.
type ReservationExpirer interface {
	ExpireReservations(ctx context.Context) (int, error)
}

// ReminderSender sends shift reminder emails for upcoming shifts. It is optional;
// when nil the reminder job is skipped.
type ReminderSender interface {
	SendDueReminders(ctx context.Context) (int, error)
}

// Scheduler runs background maintenance jobs on fixed intervals. Each job
// acquires a PostgreSQL advisory lock before executing so that exactly one
// application instance performs the work at any given time.
type Scheduler struct {
	pool      *pgxpool.Pool
	expirer   ReservationExpirer
	reminders ReminderSender
	log       zerolog.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new Scheduler. reminders may be nil to disable the reminder job.
func New(pool *pgxpool.Pool, expirer ReservationExpirer, reminders ReminderSender, log zerolog.Logger) *Scheduler {
	return &Scheduler{
		pool:      pool,
		expirer:   expirer,
		reminders: reminders,
		log:       log.With().Str("component", "scheduler").Logger(),
	}
}

// Start launches the background jobs. It returns immediately; call Stop to
// terminate them gracefully.
func (s *Scheduler) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.run(ctx, "expire-reservations", 15*time.Minute, lockKeyExpireReservations, s.expireReservations)
	s.run(ctx, "send-reminders", time.Hour, lockKeySendReminders, s.sendReminders)

	s.log.Info().Msg("scheduler started")
}

// Stop signals all jobs to terminate and waits for them to finish.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	s.log.Info().Msg("scheduler stopped")
}

// run starts a ticker-driven loop for a single job.
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

// execute acquires the advisory lock, runs the job, and releases the lock.
func (s *Scheduler) execute(ctx context.Context, name string, lockKey int64, job func(context.Context) (int, error)) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		s.log.Error().Err(err).Str("job", name).Msg("acquire connection")
		return
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, lockKey).Scan(&locked); err != nil {
		s.log.Error().Err(err).Str("job", name).Msg("acquire advisory lock")
		return
	}
	if !locked {
		// Another instance holds the lock; skip this run.
		return
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, lockKey)
	}()

	n, err := job(ctx)
	if err != nil {
		s.log.Error().Err(err).Str("job", name).Msg("job failed")
		return
	}
	s.log.Info().Str("job", name).Int("affected", n).Msg("job completed")
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
