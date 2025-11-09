# 予約データ自動クリーンアップ機能

## 概要

このアプリケーションには、予約データが永久的に蓄積しないように、自動的にクリーンアップする機能が実装されています。

## クリーンアップの仕組み

### 1. 期限切れ予約の自動完了

**動作:**
- 終了時刻が過ぎた `pending`（予約中）ステータスの予約を自動的に `completed`（完了）に変更します
- 実行頻度: **毎日午前3時**に実行
- Bot起動時が深夜0時台の場合は、起動直後にも実行されます

**目的:**
- 終了時刻が過ぎても手動で完了されていない予約を自動的に完了状態にすることで、予約状態を正確に保ちます

### 2. 古い予約データの削除

**動作:**
- `completed`（完了）または `cancelled`（キャンセル済み）ステータスの予約で、最終更新日時から **30日以上経過** したものを自動削除します
- 実行頻度: **毎日午前3時10分**に実行
- Bot起動時が深夜0時台の場合は、起動直後にも実行されます

**保持期間:**
- デフォルト: **30日間**
- 完了済み・キャンセル済みの予約は、その状態に変更されてから30日間保持されます

**対象:**
- ✅ `completed` (完了) ステータスの予約
- ✅ `cancelled` (キャンセル済み) ステータスの予約
- ❌ `pending` (予約中) ステータスの予約は削除されません

### 3. 判定基準

予約は `UpdatedAt` フィールド（最終更新日時）を基準に削除されます：

```
現在時刻 - 30日 > UpdatedAt の場合に削除
```

**例:**
- 今日が 2025年11月9日の場合
- 2025年10月9日以前に完了/キャンセルされた予約が削除対象
- 2025年10月10日以降に完了/キャンセルされた予約は保持

## 実装詳細

### storage/storage.go

#### `AutoCompleteExpiredReservations()`
```go
// 終了時刻が過ぎたpending予約を自動的にcompletedに変更する
func (s *Storage) AutoCompleteExpiredReservations() (int, error)
```
- 戻り値: 自動完了した予約の数、エラー

#### `CleanupOldReservations(retentionDays int)`
```go
// 古い完了済み・キャンセル済み予約を削除する
// retentionDays: 保持期間（日数）
func (s *Storage) CleanupOldReservations(retentionDays int) (int, error)
```
- 引数: `retentionDays` - 保持期間（日数）
- 戻り値: 削除した予約の数、エラー

#### `DeleteReservation(id string)`
```go
// 指定されたIDの予約を削除する
func (s *Storage) DeleteReservation(id string) error
```
- 引数: `id` - 削除する予約のID
- 戻り値: エラー

### main.go

定期実行タスクとして実装されています：

```go
// 定期的に期限切れ予約を自動完了（毎日午前3時）
go func() {
    for {
        now := time.Now()
        // 次の午前3時を計算
        next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
        if !now.Before(next) {
            // 今日の3時を過ぎている場合は明日の3時
            next = next.Add(24 * time.Hour)
        }

        // 次の実行時刻まで待機
        duration := time.Until(next)
        log.Printf("Next auto-complete scheduled at: %s (in %v)", next.Format("2006-01-02 15:04:05"), duration)
        time.Sleep(duration)

        // 終了時刻が過ぎたpending予約を自動完了
        completedCount, err := store.AutoCompleteExpiredReservations()
        if err != nil {
            log.Printf("Failed to auto-complete expired reservations: %v", err)
        } else if completedCount > 0 {
            log.Printf("Auto-completed %d expired reservation(s)", completedCount)
            if err := store.Save(); err != nil {
                log.Printf("Failed to save after auto-completion: %v", err)
            }
        }
    }
}()

// 定期的に古い予約データをクリーンアップ（毎日午前3時10分）
go func() {
    for {
        now := time.Now()
        // 次の午前3時10分を計算
        next := time.Date(now.Year(), now.Month(), now.Day(), 3, 10, 0, 0, now.Location())
        if !now.Before(next) {
            // 今日の3時10分を過ぎている場合は明日の3時10分
            next = next.Add(24 * time.Hour)
        }

        // 次の実行時刻まで待機
        duration := time.Until(next)
        log.Printf("Next cleanup scheduled at: %s (in %v)", next.Format("2006-01-02 15:04:05"), duration)
        time.Sleep(duration)

        // 古い完了済み・キャンセル済み予約を削除（30日以上前）
        deletedCount, err := store.CleanupOldReservations(30)
        if err != nil {
            log.Printf("Failed to cleanup old reservations: %v", err)
        } else if deletedCount > 0 {
            log.Printf("Cleaned up %d old reservation(s)", deletedCount)
            if err := store.Save(); err != nil {
                log.Printf("Failed to save after cleanup: %v", err)
            }
        }
    }
}()
```

## ログ出力

クリーンアップ処理が実行されると、以下のようなログが出力されます：

### 起動時
```
Next auto-complete scheduled at: 2025-11-10 03:00:00 (in 8h45m32s)
Next cleanup scheduled at: 2025-11-10 03:10:00 (in 8h55m32s)
```

### 実行時
```
Auto-completed 3 expired reservation(s)
Cleaned up 5 old reservation(s)
```

### 何もない場合
```
Auto-complete check completed: no expired reservations found
Cleanup check completed: no old reservations to remove
```

## カスタマイズ

### 保持期間の変更

`main.go` の以下の行を編集することで、保持期間を変更できます：

```go
deletedCount, err := store.CleanupOldReservations(30)  // 30を希望する日数に変更
```

### 実行時刻の変更

`main.go` の以下の部分を編集することで、実行時刻を変更できます：

**期限切れ予約の自動完了の実行時刻:**
```go
// 午前3時に実行
next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
```

時刻を変更する例：
```go
// 午前2時に実行
next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())

// 午後11時に実行
next := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, now.Location())
```

**古い予約データの削除の実行時刻:**
```go
// 午前3時10分に実行
next := time.Date(now.Year(), now.Month(), now.Day(), 3, 10, 0, 0, now.Location())
```

時刻を変更する例：
```go
// 午前4時30分に実行
next := time.Date(now.Year(), now.Month(), now.Day(), 4, 30, 0, 0, now.Location())
```

## メリット

1. **ディスク容量の節約**: 古い予約データが自動的に削除されるため、JSONファイルのサイズが無限に増大することを防ぎます

2. **パフォーマンスの維持**: データ量が適切に管理されるため、アプリケーションのパフォーマンスが保たれます

3. **データの整理**: 古くて不要なデータが自動的に削除されるため、データが整理された状態を保てます

4. **自動化**: 手動でのメンテナンス作業が不要です

## 注意事項

- `pending` ステータスの予約は、たとえ古くても削除されません（時刻が過ぎると自動的に `completed` に変更されます）
- 削除された予約は復元できません
- 保持期間を短くしすぎると、必要なデータまで削除される可能性があります
- クリーンアップ実行時にはデータが自動保存されます
