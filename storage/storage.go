package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dice/hxs_reservation_system/models"
)

const dataFilePath = "data/reservations.json"

// Storage は予約データを管理する
type Storage struct {
	mu           sync.RWMutex
	Reservations map[string]*models.Reservation `json:"reservations"`
}

// NewStorage は新しいStorageインスタンスを作成する
func NewStorage() *Storage {
	return &Storage{
		Reservations: make(map[string]*models.Reservation),
	}
}

// Load はファイルから予約データを読み込む
func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// ファイルが存在しない場合は新規作成
	if _, err := os.Stat(dataFilePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(dataFilePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &s.Reservations)
}

// Save は予約データをファイルに保存する
func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.Reservations, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dataFilePath, data, 0644)
}

// AddReservation は新しい予約を追加する
func (s *Storage) AddReservation(reservation *models.Reservation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Reservations[reservation.ID]; exists {
		return errors.New("reservation with this ID already exists")
	}

	s.Reservations[reservation.ID] = reservation
	return nil
}

// GetReservation は指定されたIDの予約を取得する
func (s *Storage) GetReservation(id string) (*models.Reservation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reservation, exists := s.Reservations[id]
	if !exists {
		return nil, errors.New("reservation not found")
	}

	return reservation, nil
}

// UpdateReservation は予約情報を更新する
func (s *Storage) UpdateReservation(reservation *models.Reservation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Reservations[reservation.ID]; !exists {
		return errors.New("reservation not found")
	}

	s.Reservations[reservation.ID] = reservation
	return nil
}

// GetAllReservations はすべての予約を取得する
func (s *Storage) GetAllReservations() []*models.Reservation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reservations := make([]*models.Reservation, 0, len(s.Reservations))
	for _, r := range s.Reservations {
		reservations = append(reservations, r)
	}

	return reservations
}

// GetUserReservations は指定されたユーザーの予約を取得する
func (s *Storage) GetUserReservations(userID string) []*models.Reservation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reservations := make([]*models.Reservation, 0)
	for _, r := range s.Reservations {
		if r.UserID == userID {
			reservations = append(reservations, r)
		}
	}

	return reservations
}

// CheckOverlap は時間の重複をチェックする
func (s *Storage) CheckOverlap(newReservation *models.Reservation) (*models.Reservation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, existing := range s.Reservations {
		// 同じIDの場合はスキップ
		if existing.ID == newReservation.ID {
			continue
		}

		overlaps, err := newReservation.OverlapsWith(existing)
		if err != nil {
			return nil, err
		}

		if overlaps {
			return existing, nil
		}
	}

	return nil, nil
}

// DeleteReservation は指定されたIDの予約を削除する
func (s *Storage) DeleteReservation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Reservations[id]; !exists {
		return errors.New("reservation not found")
	}

	delete(s.Reservations, id)
	return nil
}

// AutoCompleteExpiredReservations は終了時刻が過ぎたpending予約を自動的にcompletedに変更する
func (s *Storage) AutoCompleteExpiredReservations() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0

	for _, reservation := range s.Reservations {
		// pending状態の予約のみ対象
		if reservation.Status != models.StatusPending {
			continue
		}

		// 終了時刻を取得
		endDateTime, err := reservation.GetEndDateTime()
		if err != nil {
			return count, fmt.Errorf("failed to parse end time for reservation %s: %w", reservation.ID, err)
		}

		// 終了時刻が過ぎていればcompletedに変更
		if endDateTime.Before(now) {
			reservation.Status = models.StatusCompleted
			reservation.UpdatedAt = now
			count++
		}
	}

	// 変更があった場合は即座に保存
	if count > 0 {
		data, err := json.MarshalIndent(s.Reservations, "", "  ")
		if err != nil {
			return count, err
		}
		if err := os.WriteFile(dataFilePath, data, 0644); err != nil {
			return count, err
		}
	}

	return count, nil
}

// CleanupOldReservations は古い完了済み・キャンセル済み予約を削除する
// retentionDays: 保持期間（日数）
func (s *Storage) CleanupOldReservations(retentionDays int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoffTime := now.AddDate(0, 0, -retentionDays)
	count := 0
	idsToDelete := make([]string, 0)

	for id, reservation := range s.Reservations {
		// 完了済みまたはキャンセル済みの予約のみ対象
		if reservation.Status != models.StatusCompleted && reservation.Status != models.StatusCancelled {
			continue
		}

		// UpdatedAtが保持期間を超えていれば削除対象
		if reservation.UpdatedAt.Before(cutoffTime) {
			idsToDelete = append(idsToDelete, id)
			count++
		}
	}

	// 削除を実行
	for _, id := range idsToDelete {
		delete(s.Reservations, id)
	}

	// 削除があった場合は即座に保存
	if count > 0 {
		data, err := json.MarshalIndent(s.Reservations, "", "  ")
		if err != nil {
			return count, err
		}
		if err := os.WriteFile(dataFilePath, data, 0644); err != nil {
			return count, err
		}
	}

	return count, nil
}
