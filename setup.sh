#!/bin/bash

# 色の定義
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Go Discord Bot 開発環境セットアップ ===${NC}\n"

# 1. Goのバージョンチェック
echo -e "${YELLOW}1. Goのバージョンを確認中...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}エラー: Goがインストールされていません${NC}"
    echo "https://golang.org/dl/ からGoをインストールしてください"
    exit 1
fi

GO_VERSION=$(go version)
echo -e "${GREEN}✓ ${GO_VERSION}${NC}\n"

# 2. プロジェクトディレクトリの確認
echo -e "${YELLOW}2. プロジェクトディレクトリを確認中...${NC}"
if [ ! -f "go.mod" ]; then
    echo -e "${RED}エラー: go.modファイルが見つかりません${NC}"
    echo "プロジェクトのルートディレクトリで実行してください"
    exit 1
fi
echo -e "${GREEN}✓ プロジェクトディレクトリを確認しました${NC}\n"

# 3. 依存関係のダウンロード
echo -e "${YELLOW}3. 依存関係をダウンロード中...${NC}"
go mod download
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 依存関係のダウンロード完了${NC}\n"
else
    echo -e "${RED}エラー: 依存関係のダウンロードに失敗しました${NC}"
    exit 1
fi

# 4. go.modの整理
echo -e "${YELLOW}4. go.modを整理中...${NC}"
go mod tidy
echo -e "${GREEN}✓ go.mod整理完了${NC}\n"

# 5. .envファイルの作成
echo -e "${YELLOW}5. 環境変数ファイルをチェック中...${NC}"
if [ ! -f ".env" ]; then
    if [ -f "config/.env.example" ]; then
        cp config/.env.example .env
        echo -e "${GREEN}✓ .envファイルを作成しました${NC}"
        echo -e "${YELLOW}⚠️  .envファイルを編集して、以下を設定してください:${NC}"
        echo "   - DISCORD_TOKEN: BotのトークンDiscord Developer Portalで取得）"
        echo "   - GUILD_ID: テスト用サーバーのID（オプション）"
        echo "   - FEEDBACK_CHANNEL_ID: フィードバック受信チャンネルのID（オプション、/feedbackコマンド用）"
    else
        echo -e "${RED}エラー: config/.env.exampleファイルが見つかりません${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ .envファイルは既に存在します${NC}"
fi
echo ""

# 6. ビルドディレクトリの作成
echo -e "${YELLOW}6. ビルドディレクトリを作成中...${NC}"
mkdir -p bin
echo -e "${GREEN}✓ binディレクトリを作成しました${NC}\n"

# 7. 依存関係の検証
echo -e "${YELLOW}7. 依存関係を検証中...${NC}"
go mod verify
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 依存関係の検証完了${NC}\n"
else
    echo -e "${RED}エラー: 依存関係の検証に失敗しました${NC}"
    exit 1
fi

# 8. ビルドテスト
echo -e "${YELLOW}8. ビルドテストを実行中...${NC}"
go build -o bin/booking.hxs main.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ ビルドテスト成功${NC}\n"
    rm -f bin/booking.hxs
else
    echo -e "${RED}エラー: ビルドに失敗しました${NC}"
    exit 1
fi

# 完了メッセージ
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ セットアップが完了しました！${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${YELLOW}次のステップ:${NC}"
echo "1. .env ファイルを編集してDiscord Botトークンを設定"
echo "   vi .env"
echo ""
echo "2. アプリケーションを実行"
echo "   make run        # 開発モードで実行"
echo "   make build      # ビルド"
echo "   make start      # ビルドしてから実行"
echo ""
echo "3. 利用可能なコマンドを確認"
echo "   make help"
echo ""
echo -e "${GREEN}開発を楽しんでください！${NC}"
