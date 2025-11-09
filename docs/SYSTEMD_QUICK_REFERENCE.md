# systemd サービスファイル クイックリファレンス

## 📍 ファイルの場所

- **プロジェクト内**: `config/hxs-reservation-bot.service`
- **systemd ディレクトリ**: `/etc/systemd/system/hxs-reservation-bot.service`（インストール後）

## 🚀 クイックセットアップ

### 1. 自動セットアップ（推奨）

```bash
# プロジェクトルートで実行
./setup-systemd.sh
```

### 2. 手動セットアップ

```bash
# サービスファイルをコピー
sudo cp config/hxs-reservation-bot.service /etc/systemd/system/

# サービスファイルを編集
sudo nano /etc/systemd/system/hxs-reservation-bot.service

# systemd をリロード
sudo systemctl daemon-reload

# サービスを有効化・起動
sudo systemctl enable hxs-reservation-bot.service
sudo systemctl start hxs-reservation-bot.service
```

## ⚙️ 必須のカスタマイズ項目

サービスファイルをコピーした後、以下の項目を**必ず**編集してください：

### 1. User（実行ユーザー）

```ini
[Service]
User=hxs  # ← これを実際のユーザー名に変更
```

**例**:
```ini
User=dice
```

### 2. WorkingDirectory（プロジェクトディレクトリ）

```ini
WorkingDirectory=/home/hxs/booking.hxs  # ← 実際のパスに変更
```

**例**:
```ini
WorkingDirectory=/home/dice/programs/booking.hxs
```

### 3. ExecStart（実行ファイルのパス）

```ini
ExecStart=/home/hxs/booking.hxs/bin/hxs_reservation_system  # ← 実際のパスに変更
```

**例**:
```ini
ExecStart=/home/dice/programs/booking.hxs/bin/hxs_reservation_system
```

### 4. 環境変数の設定（どちらか選択）

#### 方法A: EnvironmentFile を使用

```ini
# コメントを解除して実際のパスを指定
EnvironmentFile=/home/dice/programs/booking.hxs/.env
```

#### 方法B: Environment で直接指定（推奨）

```ini
Environment="DISCORD_TOKEN=MTQy...実際のトークン"
Environment="GUILD_ID=1228693698632618117"
Environment="ENV=production"
```

## 📝 完全な設定例

実際の環境に合わせた設定例：

```ini
[Unit]
Description=HXS Reservation System Discord Bot
After=network.target

[Service]
# 実際のユーザー名
User=dice

# 実際のプロジェクトパス
WorkingDirectory=/home/dice/programs/booking.hxs

# 実際の実行ファイルパス
ExecStart=/home/dice/programs/booking.hxs/bin/hxs_reservation_system

# 自動再起動設定
Restart=always
RestartSec=10

# 環境変数（直接指定）
Environment="DISCORD_TOKEN=MTQy...実際のトークン"
Environment="GUILD_ID=1228693698632618117"
Environment="ENV=production"

# または .env ファイルから読み込み
# EnvironmentFile=/home/dice/programs/booking.hxs/.env

# ログ設定
StandardOutput=journal
StandardError=journal
SyslogIdentifier=hxs-bot

[Install]
WantedBy=multi-user.target
```

## 🔒 セキュリティ設定

サービスファイルに機密情報（トークン）を含む場合、パーミッションを制限：

```bash
sudo chmod 600 /etc/systemd/system/hxs-reservation-bot.service
```

## 🎯 セットアップ後の確認

```bash
# サービスの状態確認
sudo systemctl status hxs-reservation-bot.service

# ログ確認
sudo journalctl -u hxs-reservation-bot.service -f

# 自動起動が有効か確認
sudo systemctl is-enabled hxs-reservation-bot.service
```

## ⚡ よくあるコマンド

```bash
# 起動
sudo systemctl start hxs-reservation-bot.service

# 停止
sudo systemctl stop hxs-reservation-bot.service

# 再起動
sudo systemctl restart hxs-reservation-bot.service

# 状態確認
sudo systemctl status hxs-reservation-bot.service

# ログ表示（リアルタイム）
sudo journalctl -u hxs-reservation-bot.service -f

# ログ表示（最新100行）
sudo journalctl -u hxs-reservation-bot.service -n 100
```

## 🐛 トラブルシューティング

### サービスが起動しない

1. **設定を確認**:
   ```bash
   sudo systemctl status hxs-reservation-bot.service
   ```

2. **詳細ログを確認**:
   ```bash
   sudo journalctl -u hxs-reservation-bot.service -n 50
   ```

3. **よくある原因**:
   - User が存在しない
   - WorkingDirectory のパスが間違っている
   - ExecStart のファイルが存在しない
   - 環境変数が設定されていない
   - 実行権限がない

### パーミッションエラー

```bash
# ファイルの所有者を確認
ls -l /home/dice/programs/booking.hxs/bin/hxs_reservation_system

# 所有者を変更（必要な場合）
sudo chown dice:dice /home/dice/programs/booking.hxs/bin/hxs_reservation_system

# 実行権限を付与
chmod +x /home/dice/programs/booking.hxs/bin/hxs_reservation_system
```

### 環境変数が読み込まれない

EnvironmentFile を使用する場合：

```bash
# .env ファイルの存在確認
ls -l /home/dice/programs/booking.hxs/.env

# .env ファイルの内容確認
cat /home/dice/programs/booking.hxs/.env

# パーミッション確認
ls -l /home/dice/programs/booking.hxs/.env
```

または、Environment で直接指定することを推奨します。

## 📚 関連ドキュメント

- [詳細な systemd セットアップガイド](SYSTEMD_SETUP.md)
- [プロジェクト構造](PROJECT_STRUCTURE.md)
- [環境変数の設定](../config/.env.example)
