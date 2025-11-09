# 🗄️ データの取り扱い

このドキュメントでは、予約データとログの管理について説明します。

## 📑 目次

- [データ保存](#データ保存)
- [自動クリーンアップ](#自動クリーンアップ)
- [ログシステム](#ログシステム)
- [データバックアップ](#データバックアップ)
- [カスタマイズ](#カスタマイズ)

---

## データ保存

### 予約データファイル

予約データは **JSON形式** で保存されます。

| ファイル | 説明 |
|---------|------|
| `reservations.json` | 本番環境の予約データ |
| `reservations_dev.json` | 開発環境の予約データ（設定により） |

### データ構造

```json
[
  {
    "id": "a1b2c3d4e5f6g7h8",
    "user_id": "123456789012345678",
    "username": "ユーザー名",
    "date": "2025-11-15",
    "start_time": "14:00",
    "end_time": "15:00",
    "comment": "技術面接",
    "status": "pending",
    "created_at": "2025-11-09T10:00:00Z",
    "updated_at": "2025-11-09T10:00:00Z"
  }
]
```

### ステータス

| ステータス | 説明 | 絵文字 |
|-----------|------|--------|
| `pending` | 予約中 | 📅 |
| `completed` | 完了 | ✅ |
| `cancelled` | キャンセル済み | 🚫 |

### 自動保存のタイミング

予約データは以下のタイミングで自動保存されます：

- ✅ Bot起動時に読み込み
- ✅ 予約作成時（即座に保存）
- ✅ 予約更新時（即座に保存）
- ✅ 予約削除時（即座に保存）
- ✅ **5分ごと**に定期保存
- ✅ Bot終了時に保存

---

## 自動クリーンアップ

予約データが永久に蓄積しないように、自動クリーンアップ機能が実装されています。

### 1. 期限切れ予約の自動完了

**実行時刻**: **毎日午前3時00分**

**動作**:
- 終了時刻が過ぎた `pending`（予約中）予約を自動的に `completed`（完了）に変更

**目的**:
- 予約状態を正確に保つ
- 手動完了忘れを防ぐ

**ログ出力例**:
```
[2025-11-10 03:00:00] Auto-completed 3 expired reservation(s)
[2025-11-10 03:00:00] Reservations saved successfully
```

---

### 2. 古い予約データの自動削除

**実行時刻**: **毎日午前3時10分**

**動作**:
- `completed` または `cancelled` ステータスの予約で、最終更新から **30日以上** 経過したものを自動削除

**対象**:
- ✅ `completed`（完了）ステータスの予約
- ✅ `cancelled`（キャンセル済み）ステータスの予約
- ❌ `pending`（予約中）は削除されません

**判定基準**:
```
現在時刻 - 30日 > UpdatedAt の場合に削除
```

**例**:
- 今日: 2025年11月9日
- 削除対象: 2025年10月9日以前に完了/キャンセルされた予約
- 保持: 2025年10月10日以降の予約

**ログ出力例**:
```
[2025-11-10 03:10:00] Cleaned up 5 old reservation(s)
[2025-11-10 03:10:00] Reservations saved successfully
```

---

### 3. 起動時の動作

Bot起動時が深夜0時〜0時5分の場合、即座にクリーンアップを実行します。

**ログ出力例**:
```
Bot is now running. Press CTRL+C to exit.
Next auto-complete scheduled at: 2025-11-10 03:00:00 (in 8h45m32s)
Next cleanup scheduled at: 2025-11-10 03:10:00 (in 8h55m32s)
```

---

## ログシステム

### ログファイル構造

```
logs/
├── command_stats.json          # コマンド統計（JSON）
├── commands_2025-11.log        # 2025年11月のコマンドログ
├── commands_2025-12.log        # 2025年12月のコマンドログ
├── errors_2025-11.log          # 2025年11月のエラーログ
├── errors_2025-12.log          # 2025年12月のエラーログ
└── ...
```

### 1. コマンドログ

**ファイル名形式**: `commands_YYYY-MM.log`

**内容**: 各行がJSON形式のログエントリ

```json
{
  "timestamp": "2025-11-09T14:30:00Z",
  "command": "reserve",
  "user_id": "123456789012345678",
  "username": "ユーザー名",
  "channel_id": "987654321098765432",
  "success": true,
  "error": "",
  "parameters": {
    "date": "2025-11-15",
    "start_time": "14:00",
    "end_time": "15:00"
  }
}
```

### 2. エラーログ（NEW）

**ファイル名形式**: `errors_YYYY-MM.log`

**内容**: システムエラーとコマンドエラーの詳細

```json
{
  "timestamp": "2025-11-09T14:35:00Z",
  "level": "ERROR",
  "source": "handlers.handleReserve",
  "message": "Failed to save reservations",
  "error": "write error: disk full",
  "details": {
    "user_id": "123456789012345678",
    "reservation_id": "a1b2c3d4e5f6g7h8"
  }
}
```

**記録されるエラー:**
- 予約データの保存/読み込みエラー
- データベース操作エラー
- 自動完了/クリーンアップエラー
- コマンド実行時のエラー

### 3. 統計ファイル

**ファイル名**: `command_stats.json`

**内容**: コマンドの統計情報

```json
{
  "total_commands": 150,
  "command_counts": {
    "reserve": 45,
    "cancel": 12,
    "complete": 38,
    "list": 30,
    "my-reservations": 20,
    "help": 3,
    "feedback": 2
  },
  "user_counts": {
    "123456789012345678": 25,
    "987654321098765432": 18
  },
  "last_updated": "2025-11-09T14:30:00Z",
  "monthly_stats": {
    "2025-11": {
      "year": 2025,
      "month": 11,
      "total_commands": 150,
      "command_counts": {...},
      "user_counts": {...}
    }
  }
}
```

### 4. システムログ（標準出力）

システムイベントは標準出力に出力されます：

**出力例:**
```
💾 Reservations saved successfully
✅ Auto-completed 3 expired reservation(s) and saved
🗑️  Cleaned up 5 old reservation(s) and saved
❌ Failed to save reservations: write error
```

**記録されるイベント:**
- 定期保存の成功/失敗
- 自動完了の実行結果
- クリーンアップの実行結果
- Bot起動/終了


### ログローテーション

**自動ローテーション**:
- 月が変わると自動的に新しいログファイルを作成
- 古いログファイルは保持される

**古いログの削除**:
- **1か月以上前**のログファイルを自動削除（`commands_*.log` と `errors_*.log`）
- 実行頻度: **24時間ごと**
- ディスク容量を節約

### ログの確認方法

```bash
# コマンドログを確認
cat logs/commands_2025-11.log

# エラーログを確認
cat logs/errors_2025-11.log

# リアルタイムでログを監視
tail -f logs/commands_2025-11.log
tail -f logs/errors_2025-11.log

# 特定のコマンドのログを検索
grep '"command":"reserve"' logs/commands_2025-11.log

# コマンドエラーのみを表示
grep '"success":false' logs/commands_2025-11.log

# エラーログから特定のエラーを検索
grep '"level":"ERROR"' logs/errors_2025-11.log

# 統計情報を確認
cat logs/command_stats.json | jq .

# 総コマンド数を確認
cat logs/command_stats.json | jq '.total_commands'

# エラー発生件数をカウント
wc -l logs/errors_2025-11.log
```

---

## データバックアップ

### 予約データのバックアップ

```bash
# 手動バックアップ
cp reservations.json reservations_backup_$(date +%Y%m%d).json

# 定期的なバックアップ（cronで設定）
# 毎日午前2時にバックアップ
0 2 * * * cp /path/to/reservations.json /path/to/backups/reservations_$(date +\%Y\%m\%d).json
```

### ログのバックアップ

```bash
# ログディレクトリ全体をバックアップ
tar -czf logs_backup_$(date +%Y%m%d).tar.gz logs/

# 特定の月のログのみバックアップ
cp logs/commands_2025-11.log backups/
```

### データの復元

```bash
# 予約データを復元
cp reservations_backup_20251109.json reservations.json

# Botを再起動して変更を反映
systemctl restart hxs-reservation-bot
```

---

## カスタマイズ

### 保持期間の変更

`main.go` の以下の行を編集：

```go
// デフォルト: 30日
deletedCount, err := store.CleanupOldReservations(30)

// 例: 60日に変更
deletedCount, err := store.CleanupOldReservations(60)

// 例: 7日に変更（短期間）
deletedCount, err := store.CleanupOldReservations(7)
```

変更後、再ビルドが必要です：
```bash
make build
```

---

### クリーンアップ実行時刻の変更

#### 期限切れ予約の自動完了

`main.go` の以下の部分を編集：

```go
// デフォルト: 午前3時
next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())

// 例: 午前2時に変更
next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())

// 例: 午後11時に変更
next := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, now.Location())
```

#### 古いデータの削除

```go
// デフォルト: 午前3時10分
next := time.Date(now.Year(), now.Month(), now.Day(), 3, 10, 0, 0, now.Location())

// 例: 午前4時30分に変更
next := time.Date(now.Year(), now.Month(), now.Day(), 4, 30, 0, 0, now.Location())
```

---

### データファイルの場所を変更

環境変数で変更できます：

```env
# .env ファイル
DATA_FILE=custom_reservations.json
```

または `main.go` で直接指定：

```go
store = storage.NewStorage("custom_reservations.json")
```

---

### ログディレクトリの変更

`main.go` の以下の行を編集：

```go
// デフォルト: ./logs
logger = logging.NewLogger("./logs")

// 例: /var/log/hxs に変更
logger = logging.NewLogger("/var/log/hxs")
```

---

## メリット

### 自動クリーンアップのメリット

1. **ディスク容量の節約**
   - 古いデータが自動削除され、ファイルサイズが無限に増大しない

2. **パフォーマンスの維持**
   - データ量が適切に管理され、アプリケーションの動作が軽快

3. **データの整理**
   - 不要な古いデータが削除され、常に整理された状態

4. **完全自動化**
   - 手動メンテナンス不要

### ログシステムのメリット

1. **トラブルシューティング**
   - エラー発生時に原因を特定できる

2. **使用状況の把握**
   - どのコマンドがよく使われているか分析できる

3. **ユーザー行動の分析**
   - ユーザーごとの使用頻度を把握できる

4. **月次レポート**
   - 月別の統計情報を確認できる

---

## 注意事項

### クリーンアップに関する注意

- ❌ `pending` ステータスの予約は削除されません
- ❌ 削除されたデータは復元できません
- ⚠️ 保持期間を短くしすぎると必要なデータまで削除される可能性があります
- ✅ 削除前にバックアップを取ることを推奨します

### ログに関する注意

- 🔒 機密情報（トークン、パスワードなど）はログに記録されません
- 📝 フィードバック内容はログに記録されません（メッセージ長のみ記録）
- 🗂️ ログファイルは月ごとに分割されます
- 🧹 1か月以上前のログは自動削除されます

---

## トラブルシューティング

### 予約データが保存されない

**症状**: Bot再起動後にデータが消える

**原因と解決**:
1. ファイル権限を確認
   ```bash
   ls -la reservations.json
   chmod 644 reservations.json
   ```

2. ディスク容量を確認
   ```bash
   df -h
   ```

3. ログを確認
   ```bash
   tail -f logs/commands_2025-11.log
   ```

---

### クリーンアップが実行されない

**症状**: 古いデータが削除されない

**確認方法**:
```bash
# ログで実行時刻を確認
grep "Next cleanup scheduled" logs/commands_2025-11.log
grep "Cleaned up" logs/commands_2025-11.log
```

**解決方法**:
- Botが午前3時に起動しているか確認
- Botが長時間稼働しているか確認（24時間以上）
- ログにエラーがないか確認

---

### ログファイルが作成されない

**症状**: `logs/` ディレクトリが空

**解決方法**:
1. ディレクトリの作成
   ```bash
   mkdir -p logs
   chmod 755 logs
   ```

2. Botを再起動
   ```bash
   make run
   ```

3. ログが作成されたか確認
   ```bash
   ls -la logs/
   ```

---

## 監視とメンテナンス

### 定期的な確認項目

#### 毎日
- [ ] Botが正常に動作しているか
- [ ] ログにエラーがないか

#### 毎週
- [ ] ディスク容量の確認
- [ ] 予約データのバックアップ

#### 毎月
- [ ] ログファイルのサイズ確認
- [ ] 統計情報の確認
- [ ] 不要なバックアップの削除

### ディスク使用量の監視

```bash
# ログディレクトリのサイズ
du -sh logs/

# 予約データのサイズ
ls -lh reservations.json

# 全体のディスク使用量
df -h
```

---

## まとめ

このシステムのデータ管理機能：

✅ **自動保存** - データは自動的に保存される
✅ **自動クリーンアップ** - 古いデータは30日後に自動削除
✅ **自動ログローテーション** - ログは月ごとに分割
✅ **統計機能** - コマンド使用状況を自動記録
✅ **完全自動化** - 手動メンテナンス不要

---

**関連ドキュメント**: [README](../README.md) | [起動ガイド](SETUP.md) | [コマンド](COMMANDS.md) | [systemd](SYSTEMD.md) | [開発](DEVELOPMENT.md)
