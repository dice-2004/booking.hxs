package storage

import (
	"testing"
	"time"

	"github.com/dice/hxs_reservation_system/models"
)

func TestAutoCompleteExpiredReservations(t *testing.T) {
	store := NewStorage()

	// 過去の予約を作成（終了時刻が過ぎている）
	pastReservation := &models.Reservation{
		ID:        "test-past-1",
		UserID:    "user1",
		Username:  "Test User",
		Date:      time.Now().AddDate(0, 0, -1).Format("2006-01-02"), // 昨日
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.StatusPending,
		CreatedAt: time.Now().AddDate(0, 0, -1),
		UpdatedAt: time.Now().AddDate(0, 0, -1),
		ChannelID: "channel1",
	}

	// 未来の予約を作成（まだ終了していない）
	futureReservation := &models.Reservation{
		ID:        "test-future-1",
		UserID:    "user2",
		Username:  "Test User 2",
		Date:      time.Now().AddDate(0, 0, 1).Format("2006-01-02"), // 明日
		StartTime: "14:00",
		EndTime:   "15:00",
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ChannelID: "channel1",
	}

	// 既に完了している予約
	completedReservation := &models.Reservation{
		ID:        "test-completed-1",
		UserID:    "user3",
		Username:  "Test User 3",
		Date:      time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
		StartTime: "09:00",
		EndTime:   "10:00",
		Status:    models.StatusCompleted,
		CreatedAt: time.Now().AddDate(0, 0, -2),
		UpdatedAt: time.Now().AddDate(0, 0, -2),
		ChannelID: "channel1",
	}

	// 予約を追加
	store.AddReservation(pastReservation)
	store.AddReservation(futureReservation)
	store.AddReservation(completedReservation)

	// 自動完了を実行
	count, err := store.AutoCompleteExpiredReservations()
	if err != nil {
		t.Fatalf("AutoCompleteExpiredReservations failed: %v", err)
	}

	// 1件の予約が完了したはず
	if count != 1 {
		t.Errorf("Expected 1 reservation to be completed, got %d", count)
	}

	// 過去の予約がcompletedになっているか確認
	updated, err := store.GetReservation("test-past-1")
	if err != nil {
		t.Fatalf("Failed to get reservation: %v", err)
	}
	if updated.Status != models.StatusCompleted {
		t.Errorf("Expected status to be completed, got %s", updated.Status)
	}

	// 未来の予約はpendingのままか確認
	future, err := store.GetReservation("test-future-1")
	if err != nil {
		t.Fatalf("Failed to get reservation: %v", err)
	}
	if future.Status != models.StatusPending {
		t.Errorf("Expected status to be pending, got %s", future.Status)
	}

	// 既にcompletedだった予約はそのままか確認
	completed, err := store.GetReservation("test-completed-1")
	if err != nil {
		t.Fatalf("Failed to get reservation: %v", err)
	}
	if completed.Status != models.StatusCompleted {
		t.Errorf("Expected status to be completed, got %s", completed.Status)
	}
}

func TestCleanupOldReservations(t *testing.T) {
	store := NewStorage()

	// 31日前に完了した予約（削除されるはず）
	oldCompleted := &models.Reservation{
		ID:        "test-old-completed",
		UserID:    "user1",
		Username:  "Test User",
		Date:      time.Now().AddDate(0, 0, -31).Format("2006-01-02"),
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.StatusCompleted,
		CreatedAt: time.Now().AddDate(0, 0, -31),
		UpdatedAt: time.Now().AddDate(0, 0, -31),
		ChannelID: "channel1",
	}

	// 31日前にキャンセルした予約（削除されるはず）
	oldCancelled := &models.Reservation{
		ID:        "test-old-cancelled",
		UserID:    "user2",
		Username:  "Test User 2",
		Date:      time.Now().AddDate(0, 0, -31).Format("2006-01-02"),
		StartTime: "12:00",
		EndTime:   "13:00",
		Status:    models.StatusCancelled,
		CreatedAt: time.Now().AddDate(0, 0, -31),
		UpdatedAt: time.Now().AddDate(0, 0, -31),
		ChannelID: "channel1",
	}

	// 29日前に完了した予約（削除されないはず）
	recentCompleted := &models.Reservation{
		ID:        "test-recent-completed",
		UserID:    "user3",
		Username:  "Test User 3",
		Date:      time.Now().AddDate(0, 0, -29).Format("2006-01-02"),
		StartTime: "14:00",
		EndTime:   "15:00",
		Status:    models.StatusCompleted,
		CreatedAt: time.Now().AddDate(0, 0, -29),
		UpdatedAt: time.Now().AddDate(0, 0, -29),
		ChannelID: "channel1",
	}

	// 31日前のpending予約（削除されないはず）
	oldPending := &models.Reservation{
		ID:        "test-old-pending",
		UserID:    "user4",
		Username:  "Test User 4",
		Date:      time.Now().AddDate(0, 0, 1).Format("2006-01-02"), // 未来の日付
		StartTime: "16:00",
		EndTime:   "17:00",
		Status:    models.StatusPending,
		CreatedAt: time.Now().AddDate(0, 0, -31),
		UpdatedAt: time.Now().AddDate(0, 0, -31),
		ChannelID: "channel1",
	}

	// 予約を追加
	store.AddReservation(oldCompleted)
	store.AddReservation(oldCancelled)
	store.AddReservation(recentCompleted)
	store.AddReservation(oldPending)

	// クリーンアップを実行（保持期間30日）
	count, err := store.CleanupOldReservations(30)
	if err != nil {
		t.Fatalf("CleanupOldReservations failed: %v", err)
	}

	// 2件の予約が削除されたはず
	if count != 2 {
		t.Errorf("Expected 2 reservations to be deleted, got %d", count)
	}

	// 古い完了予約が削除されたか確認
	_, err = store.GetReservation("test-old-completed")
	if err == nil {
		t.Error("Expected old completed reservation to be deleted")
	}

	// 古いキャンセル予約が削除されたか確認
	_, err = store.GetReservation("test-old-cancelled")
	if err == nil {
		t.Error("Expected old cancelled reservation to be deleted")
	}

	// 最近の完了予約が残っているか確認
	_, err = store.GetReservation("test-recent-completed")
	if err != nil {
		t.Error("Expected recent completed reservation to exist")
	}

	// 古いpending予約が残っているか確認
	_, err = store.GetReservation("test-old-pending")
	if err != nil {
		t.Error("Expected old pending reservation to exist")
	}
}

func TestDeleteReservation(t *testing.T) {
	store := NewStorage()

	// テスト用予約を作成
	reservation := &models.Reservation{
		ID:        "test-delete-1",
		UserID:    "user1",
		Username:  "Test User",
		Date:      "2025-11-10",
		StartTime: "10:00",
		EndTime:   "11:00",
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ChannelID: "channel1",
	}

	// 予約を追加
	err := store.AddReservation(reservation)
	if err != nil {
		t.Fatalf("Failed to add reservation: %v", err)
	}

	// 予約が存在することを確認
	_, err = store.GetReservation("test-delete-1")
	if err != nil {
		t.Fatalf("Reservation should exist: %v", err)
	}

	// 予約を削除
	err = store.DeleteReservation("test-delete-1")
	if err != nil {
		t.Fatalf("Failed to delete reservation: %v", err)
	}

	// 予約が削除されたことを確認
	_, err = store.GetReservation("test-delete-1")
	if err == nil {
		t.Error("Reservation should be deleted")
	}

	// 存在しない予約の削除を試みる
	err = store.DeleteReservation("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent reservation")
	}
}
