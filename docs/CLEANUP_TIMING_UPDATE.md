# クリーンアップ実行タイミングの変更

## 📋 変更内容

クリーンアップ機能の実行タイミングを、「Bot起動時刻から24時間ごと」から「**毎日決まった時刻**」に変更しました。

## 🔄 変更前と変更後の比較

### 変更前（問題のある実装）

```go
// 24時間ごとに実行
ticker := time.NewTicker(24 * time.Hour)
defer ticker.Stop()
for {
    // 即座に1回実行
    // クリーンアップ処理...

    // 次のティックまで待機
    <-ticker.C
}
```

**問題点:**
- ❌ Bot起動時刻に依存（15:30起動なら毎日15:30に実行）
- ❌ Botを再起動すると実行時刻がズレる
- ❌ 実行タイミングが予測しにくい

**例:**
- Bot起動: 2025-11-09 15:30
  - 1回目: 15:30（起動直後）
  - 2回目: 2025-11-10 15:30
  - メンテナンスで13:00に再起動 → 以降は毎日13:00に実行される

### 変更後（改善版）

```go
// 毎日決まった時刻に実行
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
    log.Printf("Next cleanup scheduled at: %s (in %v)", next.Format("2006-01-02 15:04:05"), duration)
    time.Sleep(duration)

    // クリーンアップ処理...
}
```

**メリット:**
- ✅ 毎日午前3時に確実に実行される
- ✅ Botを再起動しても実行時刻は変わらない
- ✅ 実行タイミングが予測可能
- ✅ 次回実行時刻がログに表示される

**例:**
- どんな時刻にBotを起動・再起動しても、**必ず毎日午前3時**に実行される

## ⏰ 実行スケジュール

### 期限切れ予約の自動完了
- **実行時刻**: 毎日午前3時00分
- **処理内容**: 終了時刻が過ぎた`pending`予約を`completed`に変更

### 古い予約データの削除
- **実行時刻**: 毎日午前3時10分
- **処理内容**: 30日以上前の`completed`/`cancelled`予約を削除

**なぜ10分ずらす？**
- 自動完了処理が終わってから削除処理を実行するため
- 同時実行によるデータ競合を避けるため

## 📊 ログ出力

### Bot起動時
```
Bot is now running. Press CTRL+C to exit.
Next auto-complete scheduled at: 2025-11-10 03:00:00 (in 8h45m32s)
Next cleanup scheduled at: 2025-11-10 03:10:00 (in 8h55m32s)
```

### 実行時（データがある場合）
```
[2025-11-10 03:00:00] Auto-completed 3 expired reservation(s)
[2025-11-10 03:00:00] Reservations saved successfully
[2025-11-10 03:10:00] Cleaned up 5 old reservation(s)
[2025-11-10 03:10:00] Reservations saved successfully
```

### 実行時（データがない場合）
```
[2025-11-10 03:00:00] Auto-complete check completed: no expired reservations found
[2025-11-10 03:10:00] Cleanup check completed: no old reservations to remove
```

## 🔧 カスタマイズ方法

### 実行時刻の変更

`main.go` を編集して時刻を変更できます：

#### 午前3時 → 午前2時に変更
```go
// 変更前
next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())

// 変更後
next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
```

#### 午前3時 → 午後11時に変更
```go
// 変更前
next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())

// 変更後
next := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, now.Location())
```

#### 午前3時10分 → 午前4時30分に変更
```go
// 変更前
next := time.Date(now.Year(), now.Month(), now.Day(), 3, 10, 0, 0, now.Location())

// 変更後
next := time.Date(now.Year(), now.Month(), now.Day(), 4, 30, 0, 0, now.Location())
```

## 🎯 推奨設定

### サーバー負荷が低い時間帯
- **午前2時〜5時**: サーバー負荷が低く、ユーザーの利用も少ない
- **現在の設定（午前3時）**: 推奨

### 避けるべき時間帯
- **日中（9時〜18時）**: ユーザーの利用が多い
- **深夜0時前後**: 日付が切り替わる時間帯は避ける

## ✅ 動作確認

ビルドエラーなし：
```bash
$ go build -o bin/hxs_reservation_system main.go
$ echo $?
0
```

## 📚 更新されたドキュメント

1. **main.go** - 実装を変更
2. **docs/CLEANUP.md** - 実行タイミングとログ出力を更新
3. **README.md** - 実行時刻を明記

## 🔍 技術的な詳細

### タイムゾーン対応
```go
now.Location()
```
を使用することで、サーバーのタイムゾーンに自動対応します。

### 次回実行時刻の計算
```go
next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
if !now.Before(next) {
    // 今日の3時を過ぎている場合は明日の3時
    next = next.Add(24 * time.Hour)
}
```

- 現在時刻が3時より前 → 今日の3時
- 現在時刻が3時以降 → 明日の3時

### 待機時間の計算
```go
duration := time.Until(next)
time.Sleep(duration)
```

正確な時刻まで待機し、その時刻になったら処理を実行します。

## 🎉 結論

クリーンアップ機能が**毎日決まった時刻（午前3時/3時10分）**に確実に実行されるようになりました！

- ✅ 予測可能な動作
- ✅ Botの再起動に影響されない
- ✅ ログで次回実行時刻が確認できる
- ✅ 運用しやすい
