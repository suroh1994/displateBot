package backend

import (
	"context"
	"database/sql"
	"displateBot/displate"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Store interface {
	UpdateDatabase(displate.Client, context.Context, time.Duration, func([]displate.Displate))
	LimitedEditionDisplates() []displate.Displate
	AvailableDisplates() []displate.Displate
	UpcomingDisplates() []displate.Displate
	AddChat(int64)
	RemoveChat(int64)
	Chats() []int64
	Close() error
}

type store struct {
	displates          []displate.Displate
	availableDisplates []displate.Displate
	upcomingDisplates  []displate.Displate
	logger             *slog.Logger
	db                 *sql.DB
	mu                 sync.RWMutex
}

func (s *store) AvailableDisplates() []displate.Displate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.availableDisplates
}

func (s *store) UpcomingDisplates() []displate.Displate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.upcomingDisplates
}

func (s *store) LimitedEditionDisplates() []displate.Displate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.displates
}

func (s *store) AddChat(chatID int64) {
	_, err := s.db.Exec("INSERT OR IGNORE INTO chats (id) VALUES (?)", chatID)
	if err != nil {
		s.logger.Error("failed to add chat to database", "chatID", chatID, "err", err)
	}
}

func (s *store) RemoveChat(chatID int64) {
	_, err := s.db.Exec("DELETE FROM chats WHERE id = ?", chatID)
	if err != nil {
		s.logger.Error("failed to remove chat from database", "chatID", chatID, "err", err)
	}
}

func (s *store) Chats() []int64 {
	rows, err := s.db.Query("SELECT id FROM chats")
	if err != nil {
		s.logger.Error("failed to get chats from database", "err", err)
		return nil
	}
	defer rows.Close()

	var chats []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			s.logger.Error("failed to scan chat id", "err", err)
			continue
		}
		chats = append(chats, id)
	}
	return chats
}

func (s *store) Close() error {
	return s.db.Close()
}

func (s *store) UpdateDatabase(client displate.Client, ctx context.Context, interval time.Duration, onNewDisplates func([]displate.Displate)) {
	newDisplates := s.fetchDisplatesAndUpdateCache(client)
	if len(newDisplates) > 0 {
		onNewDisplates(newDisplates)
	}
	for {
		select {
		case <-time.After(interval):
			newDisplates := s.fetchDisplatesAndUpdateCache(client)
			if len(newDisplates) > 0 {
				onNewDisplates(newDisplates)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *store) fetchDisplatesAndUpdateCache(client displate.Client) []displate.Displate {
	displates, err := client.GetLimitedEditionDisplates()
	if err != nil {
		s.logger.Error("failed to update database: failed to get displates", "err", err)
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	newDisplates := make([]displate.Displate, 0)
	if len(s.displates) > 0 {
		oldDisplatesMap := make(map[int]struct{})
		for _, d := range s.displates {
			oldDisplatesMap[d.ID] = struct{}{}
		}
		for _, d := range displates {
			if _, ok := oldDisplatesMap[d.ID]; !ok {
				newDisplates = append(newDisplates, d)
			}
		}
	}

	s.displates = displates
	s.availableDisplates = displate.FilterDisplates(s.displates, func(d displate.Displate) bool {
		return d.Edition.Status == displate.StatusAvailable
	})
	s.upcomingDisplates = displate.FilterDisplates(s.displates, func(d displate.Displate) bool {
		return d.Edition.Status == displate.StatusUpcoming
	})
	return newDisplates
}

func NewStore(logger *slog.Logger, dbPath string) (Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS chats (id INTEGER PRIMARY KEY)")
	if err != nil {
		return nil, fmt.Errorf("failed to create chats table: %w", err)
	}

	return &store{
		displates: make([]displate.Displate, 0),
		logger:    logger,
		db:        db,
	}, nil
}
