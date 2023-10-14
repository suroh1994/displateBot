package backend

import (
	"context"
	"displateBot/displate"
	"log/slog"
	"time"
)

type Store interface {
	UpdateDatabase(displate.Client, context.Context)
	LimitedEditionDisplates() []displate.Displate
	AvailableDisplates() []displate.Displate
	UpcomingDisplates() []displate.Displate
}

type store struct {
	displates          []displate.Displate
	availableDisplates []displate.Displate
	upcomingDisplates  []displate.Displate
	logger             *slog.Logger
}

func (s *store) AvailableDisplates() []displate.Displate {
	return s.availableDisplates
}

func (s *store) UpcomingDisplates() []displate.Displate {
	return s.upcomingDisplates
}

func (s *store) LimitedEditionDisplates() []displate.Displate {
	return s.displates
}

func (s *store) UpdateDatabase(client displate.Client, ctx context.Context) {
	s.fetchDisplatesAndUpdateCache(client)
	for {
		select {
		//TODO Implement dynamic update interval close to releases and when sales are about to end (sell out or terminate)
		case <-time.After(1 * time.Hour):
			s.fetchDisplatesAndUpdateCache(client)
		case <-ctx.Done():
			return
		}
	}
}

func (s *store) fetchDisplatesAndUpdateCache(client displate.Client) {
	displates, err := client.GetLimitedEditionDisplates()
	if err != nil {
		s.logger.Error("failed to update database: failed to get displates: %v", err)
		return
	}
	s.displates = displates
	s.availableDisplates = displate.FilterDisplates(s.displates, func(d displate.Displate) bool {
		return d.Edition.Status == displate.StatusAvailable
	})
	s.upcomingDisplates = displate.FilterDisplates(s.displates, func(d displate.Displate) bool {
		return d.Edition.Status == displate.StatusUpcoming
	})
}

func NewStore(logger *slog.Logger) Store {
	return &store{
		displates: make([]displate.Displate, 0),
		logger:    logger,
	}
}
