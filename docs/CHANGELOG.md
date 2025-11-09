# 📋 変更履歴

このドキュメントでは、プロジェクトのバージョンごとの変更内容を記録しています。

---

## [v1.2.0] - 2025-11-09

### ✨ 新機能

#### `/help` コマンド
- すべてのコマンドの一覧と使い方を表示
- コマンドを実行した人にのみ表示（Ephemeralメッセージ）
- パラメータの詳細、プライバシー情報、データ管理情報を含む

#### `/feedback` コマンド
- システムへのご意見・ご要望を **完全匿名** で送信
- 特定のチャンネルに転送される
- コマンド実行自体も本人にしか見えない
- ログにはメッセージ長のみ記録（内容は記録されない）

### 🔧 変更

#### 環境変数
- `FEEDBACK_CHANNEL_ID` を追加
- フィードバック受信チャンネルを設定可能に

#### ドキュメント整理
- 6つのメインドキュメントに統合
  - `README.md` - プロジェクト概要とナビゲーション
  - `docs/SETUP.md` - アプリの起動ガイド
  - `docs/COMMANDS.md` - コマンドリファレンス
  - `docs/DATA_MANAGEMENT.md` - データの取り扱い
  - `docs/SYSTEMD.md` - systemdセットアップ
  - `docs/DEVELOPMENT.md` - 開発者ガイド
  - `docs/CHANGELOG.md` - 変更履歴（本ファイル）

### 📝 更新されたドキュメント

- `README.md` - 機能リストとドキュメント構造を更新
- `docs/SETUP.md` - 新規作成（セットアップ～起動の完全ガイド）
- `docs/COMMANDS.md` - 全コマンドの詳細リファレンス
- `docs/DATA_MANAGEMENT.md` - 新規作成（クリーンアップ＋ログ管理を統合）
- `docs/SYSTEMD.md` - 新規作成（systemd関連を統合）
- `docs/DEVELOPMENT.md` - 開発者向けガイドに刷新
- 設定ファイル（`.env.example`, `.env.development`, `.env.production`）に `FEEDBACK_CHANNEL_ID` を追加
- `setup.sh` - FEEDBACK_CHANNEL_IDの説明を追加

### 🗑️ 廃止されたドキュメント

以下のドキュメントは統合され、削除予定：
- `docs/QUICKSTART.md` → `docs/SETUP.md` に統合
- `docs/CLEANUP.md` → `docs/DATA_MANAGEMENT.md` に統合
- `docs/CLEANUP_TIMING_UPDATE.md` → `docs/DATA_MANAGEMENT.md` に統合
- `docs/LOGGING.md` → `docs/DATA_MANAGEMENT.md` に統合
- `docs/SYSTEMD_SETUP.md` → `docs/SYSTEMD.md` に統合
- `docs/SYSTEMD_QUICK_REFERENCE.md` → `docs/SYSTEMD.md` に統合
- `docs/PROJECT_SUMMARY.md` → 情報が古いため削除予定
- `docs/PROJECT_STRUCTURE.md` → `README.md` と `docs/DEVELOPMENT.md` に統合
- `docs/CHANGELOG_HELP_FEEDBACK.md` → `docs/CHANGELOG.md` に統合
- `docs/FILE_ORGANIZATION.md` → 不要
- `docs/REORGANIZATION.md` → 不要

---

## [v1.1.0] - 2025-11-09

### 🔧 変更

#### クリーンアップ実行タイミングの改善
- **変更前**: Bot起動時刻から24時間ごとに実行（起動時刻に依存）
- **変更後**: 毎日決まった時刻（午前3時/3時10分）に実行

#### 期限切れ予約の自動完了
- 実行時刻: **毎日午前3時00分**
- 終了時刻が過ぎた `pending` 予約を自動的に `completed` に変更

#### 古い予約データの自動削除
- 実行時刻: **毎日午前3時10分**
- 30日以上前の `completed` / `cancelled` 予約を自動削除

#### 起動時の最適化
- Bot起動が深夜0時〜0時5分の場合、即座にクリーンアップを実行
- ログに次回実行時刻と残り時間を表示

### 📝 更新されたドキュメント

- `docs/CLEANUP.md` - 実行時刻を「毎日午前3時」に更新
- `docs/CLEANUP_TIMING_UPDATE.md` - 新規作成（変更詳細を記録）
- `README.md` - クリーンアップ機能の説明を更新

---

## [v1.0.0] - 2025-10-05

### ✨ 初回リリース

#### 基本機能

##### 予約管理コマンド
- **`/reserve`** - 面接の予約を作成
- **`/cancel`** - 予約を取り消し
- **`/complete`** - 予約を完了状態に変更

##### 表示コマンド
- **`/list`** - すべての予約を表示（実行者のみに表示）
- **`/my-reservations`** - 自分の予約のみを表示（実行者のみに表示）

#### データ管理

##### 自動クリーンアップ
- 期限切れ予約の自動完了（当初は1時間ごと）
- 古い予約データの自動削除（30日以上前）

##### データ永続化
- JSON形式で予約データを保存
- 5分ごとに自動保存
- Bot終了時にも保存

#### ログシステム

##### コマンドログ
- すべてのコマンド実行を記録
- 月次ログローテーション（`commands_YYYY-MM.log`）
- 古いログの自動削除（1か月以上前）

##### 統計情報
- コマンド総数をJSON形式で保存
- コマンド別、ユーザー別の統計
- 月別統計も記録

#### 開発環境

##### 環境分離
- 開発環境（`.env.development`）
- 本番環境（`.env.production`）
- 環境切り替えスクリプト（`switch_env.sh`）

##### 自動化
- Makefileによるタスク自動化
- セットアップスクリプト（`setup.sh`）
- 依存関係管理スクリプト（`manage_deps.sh`）

##### systemd対応
- systemdサービスファイル
- 自動起動設定
- サービス管理スクリプト

#### セキュリティとプライバシー

- ✅ 推測しにくい予約IDを自動生成
- ✅ 予約IDは作成者にのみ通知（Ephemeralメッセージ）
- ✅ 時間の重複をチェック
- ✅ 環境変数で機密情報を管理

#### 技術スタック

- **言語**: Go 1.21+
- **ライブラリ**:
  - discordgo v0.27.1
  - godotenv v1.5.1
- **データ保存**: JSON
- **ログ管理**: 月次ローテーション

---

## バージョン番号の規則

このプロジェクトは [Semantic Versioning](https://semver.org/) に従います。

- **MAJOR** (例: v2.0.0): 後方互換性のない変更
- **MINOR** (例: v1.2.0): 後方互換性のある機能追加
- **PATCH** (例: v1.1.1): 後方互換性のあるバグ修正

---

## 今後の予定

### v1.3.0（計画中）

検討中の機能：
- [ ] 予約の編集機能
- [ ] 予約のリマインダー機能
- [ ] カレンダー表示機能
- [ ] 統計レポート機能
- [ ] Webダッシュボード
- [ ] データベース対応（PostgreSQL, MySQL）

---

## コントリビューション

バグ報告や機能要望は、Discord Botの `/feedback` コマンドでお送りください（完全匿名）。

---

**メンテナンス**: このドキュメントは各リリースごとに更新されます。
**最終更新**: 2025-11-09
