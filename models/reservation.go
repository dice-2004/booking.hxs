package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// ReservationStatus は予約の状態を表す
type ReservationStatus string

const (
	StatusPending   ReservationStatus = "pending"   // 予約中
	StatusCompleted ReservationStatus = "completed" // 完了
	StatusCancelled ReservationStatus = "cancelled" // キャンセル済み
)

// Reservation は予約情報を表す構造体
type Reservation struct {
	ID        string            `json:"id"`         // 予約ID（推測しにくい英数字列）
	UserID    string            `json:"user_id"`    // 予約者のDiscord ID
	Username  string            `json:"username"`   // 予約者の表示名
	Date      string            `json:"date"`       // 予約日（YYYY-MM-DD形式）
	StartTime string            `json:"start_time"` // 開始時間（HH:MM形式）
	EndTime   string            `json:"end_time"`   // 終了時間（HH:MM形式）
	Comment   string            `json:"comment"`    // コメント（オプション）
	Status    ReservationStatus `json:"status"`     // 予約状態
	CreatedAt time.Time         `json:"created_at"` // 作成日時
	UpdatedAt time.Time         `json:"updated_at"` // 更新日時
	ChannelID string            `json:"channel_id"` // 予約が行われたチャンネルID
}

// GenerateReservationID は推測しにくいランダムな予約IDを生成する
func GenerateReservationID() (string, error) {
	bytes := make([]byte, 16) // 16バイト = 32文字の16進数文字列
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetDateTime は予約日時をtime.Time型で返す
func (r *Reservation) GetDateTime(timeStr string) (time.Time, error) {
	layout := "2006-01-02 15:04"
	dateTimeStr := r.Date + " " + timeStr
	return time.Parse(layout, dateTimeStr)
}

// GetStartDateTime は予約開始日時をtime.Time型で返す
func (r *Reservation) GetStartDateTime() (time.Time, error) {
	return r.GetDateTime(r.StartTime)
}

// GetEndDateTime は予約終了日時をtime.Time型で返す
func (r *Reservation) GetEndDateTime() (time.Time, error) {
	return r.GetDateTime(r.EndTime)
}

// OverlapsWith は他の予約と時間が重複しているかチェックする
func (r *Reservation) OverlapsWith(other *Reservation) (bool, error) {
	// キャンセル済みの予約は重複チェックしない
	if r.Status == StatusCancelled || other.Status == StatusCancelled {
		return false, nil
	}

	// 日付が異なる場合は重複しない
	if r.Date != other.Date {
		return false, nil
	}

	r1Start, err := r.GetStartDateTime()
	if err != nil {
		return false, err
	}

	r1End, err := r.GetEndDateTime()
	if err != nil {
		return false, err
	}

	r2Start, err := other.GetStartDateTime()
	if err != nil {
		return false, err
	}

	r2End, err := other.GetEndDateTime()
	if err != nil {
		return false, err
	}

	// 時間の重複をチェック
	// r1の開始時間がr2の期間内にある、またはr1の終了時間がr2の期間内にある
	// またはr1がr2を完全に含む場合
	return (r1Start.Before(r2End) && r1End.After(r2Start)), nil
}
