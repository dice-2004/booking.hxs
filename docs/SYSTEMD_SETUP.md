# systemdサービスとしてのセットアップガイド

HXS予約システムをsystemdサービスとして登録し、サーバー起動時に自動起動させる方法を説明します。

## 📋 前提条件

- Goがインストールされていること
- プロジェクトがビルドされていること（`bin/hxs_reservation_system`が存在）
- `.env`ファイルに必要な環境変数が設定されていること

## 🚀 セットアップ手順

### 1. バイナリのビルド

まず、実行可能なバイナリをビルドします：

```bash
cd /home/hxs/booking.hxs
make build
# または
go build -o bin/hxs_reservation_system main.go
```

### 2. サービスファイルのコピー

プロジェクトに含まれる`hxs-reservation-bot.service`をsystemdディレクトリにコピーします：

```bash
sudo cp hxs-reservation-bot.service /etc/systemd/system/
```

### 3. 環境変数の設定（重要）

systemdサービスは通常の`.env`ファイルを自動では読み込みません。以下のいずれかの方法で環境変数を設定してください：

#### 方法A: サービスファイルに直接記述（推奨）

```bash
sudo nano /etc/systemd/system/hxs-reservation-bot.service
```

`[Service]`セクションに以下を追加：

```ini
[Service]
Environment="DISCORD_TOKEN=your_actual_token_here"
Environment="GUILD_ID=your_guild_id_here"
Environment="ENV=production"
```

#### 方法B: 環境ファイルを使用

`.env`ファイルを安全な場所に配置し、サービスファイルで参照：

```bash
sudo nano /etc/systemd/system/hxs-reservation-bot.service
```

`EnvironmentFile`の行のコメントを解除：

```ini
[Service]
EnvironmentFile=/home/hxs/booking.hxs/.env
```

### 4. systemdの設定を反映

```bash
sudo systemctl daemon-reload
```

### 5. サービスの有効化（自動起動設定）

サーバー起動時に自動的にボットを起動するように設定：

```bash
sudo systemctl enable hxs-reservation-bot.service
```

### 6. サービスの起動

```bash
sudo systemctl start hxs-reservation-bot.service
```

### 7. 起動確認

```bash
sudo systemctl status hxs-reservation-bot.service
```

正常に起動していれば、`Active: active (running)`と表示されます。

## 🔧 メンテナンスコマンド

### ボットを停止する

```bash
sudo systemctl stop hxs-reservation-bot.service
```

### ボットを再起動する

コードを更新した後など：

```bash
# コードの更新（例）
cd /home/hxs/booking.hxs
git pull

# 再ビルド
make build

# サービスの再起動
sudo systemctl restart hxs-reservation-bot.service
```

### ボットの状態を確認する

```bash
sudo systemctl status hxs-reservation-bot.service
```

終了するには `Ctrl+C` を押します。

### ログをリアルタイムで確認する

```bash
sudo journalctl -u hxs-reservation-bot.service -f
```

終了するには `Ctrl+C` を押します。

### 過去のログを確認する

```bash
# 最新100行
sudo journalctl -u hxs-reservation-bot.service -n 100

# 今日のログ
sudo journalctl -u hxs-reservation-bot.service --since today

# 特定の日時以降のログ
sudo journalctl -u hxs-reservation-bot.service --since "2025-11-09 10:00:00"
```

### 自動起動を無効化する

```bash
sudo systemctl disable hxs-reservation-bot.service
```

### サービスを完全に削除する

```bash
# サービスを停止
sudo systemctl stop hxs-reservation-bot.service

# 自動起動を無効化
sudo systemctl disable hxs-reservation-bot.service

# サービスファイルを削除
sudo rm /etc/systemd/system/hxs-reservation-bot.service

# systemdの設定を再読み込み
sudo systemctl daemon-reload
```

## 📊 便利なコマンド集

### ボットが正常に動作しているか確認

```bash
sudo systemctl is-active hxs-reservation-bot.service
```

### 最後のエラーログだけを表示

```bash
sudo journalctl -u hxs-reservation-bot.service -p err
```

### ログをファイルに保存

```bash
sudo journalctl -u hxs-reservation-bot.service > bot_logs.txt
```

## 🔍 トラブルシューティング

### ボットが起動しない場合

1. **ステータスを確認**
   ```bash
   sudo systemctl status hxs-reservation-bot.service
   ```

2. **詳細なログを確認**
   ```bash
   sudo journalctl -u hxs-reservation-bot.service -n 50 --no-pager
   ```

3. **環境変数が正しく設定されているか確認**
   - `.env`ファイルが存在し、正しい値が設定されているか
   - サービスファイルで`EnvironmentFile`が正しく指定されているか

4. **バイナリが存在し、実行可能か確認**
   ```bash
   ls -l /home/hxs/booking.hxs/bin/hxs_reservation_system
   ```

5. **権限の確認**
   ```bash
   # ファイルの所有者を確認
   ls -l /home/hxs/booking.hxs/bin/hxs_reservation_system

   # 必要に応じて実行権限を付与
   chmod +x /home/hxs/booking.hxs/bin/hxs_reservation_system
   ```

### パーミッションエラーが出る場合

サービスファイルで指定した`User`が、ワーキングディレクトリとバイナリにアクセスできることを確認してください：

```bash
# 所有者を確認
ls -la /home/hxs/booking.hxs/

# 必要に応じて所有者を変更
sudo chown -R hxs:hxs /home/hxs/booking.hxs/
```

## 📝 ワンライナーまとめ

### 開発中の更新フロー

```bash
cd /home/hxs/booking.hxs && git pull && make build && sudo systemctl restart hxs-reservation-bot.service && sudo systemctl status hxs-reservation-bot.service
```

### クイック状態確認

```bash
sudo systemctl status hxs-reservation-bot.service && sudo journalctl -u hxs-reservation-bot.service -n 20 --no-pager
```

## 🔐 セキュリティのヒント

1. **環境変数を直接サービスファイルに書く場合**
   - サービスファイルのパーミッションを確認：
     ```bash
     sudo chmod 600 /etc/systemd/system/hxs-reservation-bot.service
     ```

2. **.envファイルを使用する場合**
   - `.env`ファイルのパーミッションを制限：
     ```bash
     chmod 600 /home/hxs/booking.hxs/.env
     ```

3. **gitに機密情報をコミットしない**
   - `.gitignore`に`.env`が含まれていることを確認

## 🎯 推奨される運用フロー

1. **初回セットアップ時**
   ```bash
   make build
   sudo cp hxs-reservation-bot.service /etc/systemd/system/
   sudo nano /etc/systemd/system/hxs-reservation-bot.service  # 環境変数を設定
   sudo systemctl daemon-reload
   sudo systemctl enable hxs-reservation-bot.service
   sudo systemctl start hxs-reservation-bot.service
   sudo systemctl status hxs-reservation-bot.service
   ```

2. **コード更新時**
   ```bash
   git pull
   make build
   sudo systemctl restart hxs-reservation-bot.service
   ```

3. **定期的な確認**
   ```bash
   sudo systemctl status hxs-reservation-bot.service
   sudo journalctl -u hxs-reservation-bot.service --since "1 hour ago"
   ```

---

これで、HXS予約システムが本格的なサービスとして稼働します！🚀
