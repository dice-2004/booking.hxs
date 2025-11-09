# クイックスタートガイド

このガイドでは、最速でDiscord Botを起動する手順を説明します。

## 📋 前提条件

- Go 1.21以上がインストールされていること
- Discord Botトークンを取得済みであること

## 🚀 5分で起動

### 1. セットアップ（自動）

```bash
cd hxs_reservation_system
./setup.sh
```

これで以下が自動実行されます：
- ✅ Goバージョンチェック
- ✅ 依存関係のダウンロード
- ✅ .envファイルの作成
- ✅ ビルドテスト

### 2. Discord Botトークンを設定

```bash
vi .env
```

以下を編集：
```env
DISCORD_TOKEN=あなたのBotトークン
GUILD_ID=テスト用サーバーID（オプション）
```

### 3. 起動！

```bash
make run
```

以上です！🎉

## 💡 よく使うコマンド

```bash
make help          # コマンド一覧を表示
make run           # Botを起動
make build         # ビルド
make start         # ビルドして起動
make clean         # クリーンアップ
```

## 🔄 開発ワークフロー

### 初回セットアップ
```bash
./setup.sh
vi .env
make run
```

### コード編集後
```bash
make check         # フォーマット+静的解析
make run           # 実行
```

### 本番デプロイ前
```bash
./switch_env.sh production
make build
./bin/hxs_reservation_system
```

## 📝 Discord上でのコマンド

Botが起動したら、Discordで以下のコマンドが使えます：

- `/reserve` - 予約作成
- `/cancel` - 予約キャンセル
- `/complete` - 予約完了
- `/list` - 全予約表示
- `/my-reservations` - 自分の予約表示

## 🆘 困ったら

### Botが起動しない
```bash
# 依存関係を再インストール
make clean
make install
make run
```

### コマンドが表示されない
- Discord Developer PortalでBotの権限を確認
- `applications.commands`スコープが有効か確認
- サーバーにBotを再招待

### 依存関係のエラー
```bash
./manage_deps.sh clean
./manage_deps.sh install
```

## 📚 詳細情報

詳しい使い方は [README.md](README.md) を参照してください。

## 🎯 次のステップ

1. `.env.development`と`.env.production`を設定
2. `./switch_env.sh`で環境を切り替え
3. ホットリロードを有効化: `go install github.com/cosmtrek/air@latest && make dev`
4. コードをカスタマイズ

Happy Coding! 🚀
