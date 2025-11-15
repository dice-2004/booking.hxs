package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
)

// handleHelp はヘルプコマンドを処理する（コマンドを打った人にしか見えない）
func handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate, logger *logging.Logger, isDM bool) {
	helpMessage := "# 📖部室予約システム - ヘルプ\n" +
		"## 利用可能なコマンド:\n" +
		"**/reserve**\n" +
		"> 部室の予約を作成します\n" +
		"> - `date`: 予約日（YYYY-MM-DD または YYYY/MM/DD、例: 2025-10-15）\n" +
		"> - `start_time`: 開始時間（HH:MM形式、例: 14:00）\n" +
		"> - `end_time`: 終了時間（HH:MM形式、例: 15:00）※省略時は開始時刻+1時間\n" +
		"> - `comment`: コメント（任意）\n\n" +
		"**/edit**\n" +
		"> 予約を編集します\n" +
		"> - `reservation_id`: 予約ID\n" +
		"> - `date`: 予約日（任意）\n" +
		"> - `start_time`: 開始時間（任意）\n" +
		"> - `end_time`: 終了時間（任意）\n" +
		"> - `comment`: コメント（任意）\n\n" +
		"**/cancel**\n" +
		"> 予約を取り消します\n" +
		"> - `reservation_id`: 予約ID\n" +
		"> - `comment`: コメント（任意）\n\n" +
		"**/complete**\n" +
		"> 予約を完了にします\n" +
		"> - `reservation_id`: 予約ID\n" +
		"> - `comment`: コメント（任意）\n\n" +
		"**/list**\n" +
		"> すべての予約を表示します（自分だけに表示されます）\n\n" +
		"**/my-reservations**\n" +
		"> 自分の予約を表示します（自分だけに表示されます）\n\n" +
		"**/feedback**\n" +
		"> システムへのご意見・ご要望を匿名で送信します\n" +
		"> - `message`: フィードバック内容\n\n" +
		"**/help**\n" +
		"> このヘルプメッセージを表示します\n\n" +
		"## プライバシー:\n" +
		"- /list、/my-reservations、/help、/feedback は自分だけに表示されます\n" +
		"- 予約作成時、予約IDは予約者だけに通知されます\n" +
		"- フィードバックは完全に匿名で送信されます\n\n" +
		"## データ管理:\n" +
		"- 完了・キャンセル済みの予約は30日後に自動削除されます\n" +
		"- 期限切れの予約は毎日午前3時に自動完了されます\n\n" +
		"## 利用可能チャンネル:\n" +
		"- https://discord.com/channels/1090816023965479035/1375843736864559195で利用が可能です\n" +
		"- または、認証済みの場合のみDMでも利用可能です\n\n" +
		"## 認証方法:\n" +
		"[こちら](https://discord.com/oauth2/authorize?client_id=1425303718882185237)から認証を行ってください\n" +
		"## サポート:\n" +
		"- 問題が発生した場合は、フィードバックまでご連絡ください\n"

	userID, username := getUserInfo(i, isDM)

	respondEphemeral(s, i, helpMessage)

	// ログに記録
	logger.LogCommand("help", userID, username, i.ChannelID, true, "", nil)
}
