#!/bin/bash
#
# HXS予約システム systemdサービスセットアップスクリプト
#

set -e

# 色の定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# プロジェクトのルートディレクトリ
PROJECT_DIR="/home/hxs/booking.hxs"
SERVICE_NAME="hxs-reservation-bot.service"
SERVICE_FILE="${PROJECT_DIR}/config/${SERVICE_NAME}"
SYSTEMD_DIR="/etc/systemd/system"

echo -e "${BLUE}==================================${NC}"
echo -e "${BLUE}HXS予約システム systemdセットアップ${NC}"
echo -e "${BLUE}==================================${NC}"
echo ""

# rootチェック
if [ "$EUID" -eq 0 ]; then
    echo -e "${RED}エラー: このスクリプトはroot権限で実行しないでください${NC}"
    echo -e "${YELLOW}代わりに通常ユーザーで実行してください（必要に応じてsudoを求めます）${NC}"
    exit 1
fi

# 1. プロジェクトディレクトリの確認
echo -e "${YELLOW}[1/7] プロジェクトディレクトリの確認...${NC}"
if [ ! -d "$PROJECT_DIR" ]; then
    echo -e "${RED}エラー: プロジェクトディレクトリが見つかりません: ${PROJECT_DIR}${NC}"
    exit 1
fi
cd "$PROJECT_DIR"
echo -e "${GREEN}✓ プロジェクトディレクトリ確認完了${NC}"
echo ""

# 2. サービスファイルの確認
echo -e "${YELLOW}[2/7] サービスファイルの確認...${NC}"
if [ ! -f "$SERVICE_FILE" ]; then
    echo -e "${RED}エラー: サービスファイルが見つかりません: ${SERVICE_FILE}${NC}"
    exit 1
fi
echo -e "${GREEN}✓ サービスファイル確認完了${NC}"
echo ""

# 3. バイナリのビルド
echo -e "${YELLOW}[3/7] バイナリのビルド...${NC}"
if [ -f "Makefile" ]; then
    make build
else
    go build -o bin/hxs_reservation_system main.go
fi

if [ ! -f "bin/hxs_reservation_system" ]; then
    echo -e "${RED}エラー: バイナリのビルドに失敗しました${NC}"
    exit 1
fi
chmod +x bin/hxs_reservation_system
echo -e "${GREEN}✓ ビルド完了${NC}"
echo ""

# 4. 環境変数の確認
echo -e "${YELLOW}[4/7] 環境変数の確認...${NC}"
if [ ! -f ".env" ]; then
    echo -e "${RED}警告: .envファイルが見つかりません${NC}"
    echo -e "${YELLOW}環境変数は手動でサービスファイルに設定する必要があります${NC}"
    read -p "続行しますか？ (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✓ .envファイル確認完了${NC}"
fi
echo ""

# 5. サービスファイルのコピー
echo -e "${YELLOW}[5/7] サービスファイルのコピー...${NC}"
if [ -f "${SYSTEMD_DIR}/${SERVICE_NAME}" ]; then
    echo -e "${YELLOW}サービスファイルは既に存在します${NC}"
    read -p "上書きしますか？ (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo cp "$SERVICE_FILE" "$SYSTEMD_DIR/"
        echo -e "${GREEN}✓ サービスファイルを上書きしました${NC}"
    else
        echo -e "${YELLOW}スキップしました${NC}"
    fi
else
    sudo cp "$SERVICE_FILE" "$SYSTEMD_DIR/"
    echo -e "${GREEN}✓ サービスファイルをコピーしました${NC}"
fi
echo ""

# 6. systemdの再読み込み
echo -e "${YELLOW}[6/7] systemdの設定を再読み込み...${NC}"
sudo systemctl daemon-reload
echo -e "${GREEN}✓ 再読み込み完了${NC}"
echo ""

# 7. サービスの有効化と起動
echo -e "${YELLOW}[7/7] サービスの設定...${NC}"
echo ""
echo -e "${BLUE}次のステップを選択してください:${NC}"
echo "1) サービスを有効化して起動する（推奨）"
echo "2) サービスを有効化のみ（後で手動起動）"
echo "3) 何もしない（手動で設定）"
read -p "選択 (1-3): " -n 1 -r
echo
echo ""

case $REPLY in
    1)
        echo -e "${YELLOW}サービスを有効化して起動します...${NC}"
        sudo systemctl enable "$SERVICE_NAME"
        sudo systemctl start "$SERVICE_NAME"
        sleep 2
        sudo systemctl status "$SERVICE_NAME" --no-pager
        echo -e "${GREEN}✓ サービスが起動しました${NC}"
        ;;
    2)
        echo -e "${YELLOW}サービスを有効化します...${NC}"
        sudo systemctl enable "$SERVICE_NAME"
        echo -e "${GREEN}✓ サービスを有効化しました${NC}"
        echo -e "${YELLOW}起動するには: sudo systemctl start ${SERVICE_NAME}${NC}"
        ;;
    3)
        echo -e "${YELLOW}手動設定モードです${NC}"
        echo -e "有効化: ${BLUE}sudo systemctl enable ${SERVICE_NAME}${NC}"
        echo -e "起動: ${BLUE}sudo systemctl start ${SERVICE_NAME}${NC}"
        ;;
    *)
        echo -e "${RED}無効な選択です${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${BLUE}==================================${NC}"
echo -e "${GREEN}セットアップ完了！${NC}"
echo -e "${BLUE}==================================${NC}"
echo ""
echo -e "${YELLOW}便利なコマンド:${NC}"
echo -e "  状態確認: ${BLUE}sudo systemctl status ${SERVICE_NAME}${NC}"
echo -e "  ログ確認: ${BLUE}sudo journalctl -u ${SERVICE_NAME} -f${NC}"
echo -e "  再起動:   ${BLUE}sudo systemctl restart ${SERVICE_NAME}${NC}"
echo -e "  停止:     ${BLUE}sudo systemctl stop ${SERVICE_NAME}${NC}"
echo ""
echo -e "${YELLOW}詳しくは SYSTEMD_SETUP.md をご覧ください${NC}"
echo ""
