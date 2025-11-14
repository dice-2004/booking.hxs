package commands

import (
	"fmt"
	"os"
	"time"

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

	var userID, username string
	if isDM {
		userID = i.User.ID
		username = i.User.Username
	} else {
		userID = i.Member.User.ID
		username = getDisplayName(i.Member)
	}

	respondEphemeral(s, i, helpMessage)

	// ログに記録
	logger.LogCommand("help", userID, username, i.ChannelID, true, "", nil)
}

// handleFeedback はフィードバックコマンドを処理する（匿名で特定チャンネルに転送）
func handleFeedback(s *discordgo.Session, i *discordgo.InteractionCreate, logger *logging.Logger, isDM bool) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondError(s, i, "フィードバック内容を入力してください")
		return
	}

	message := options[0].StringValue()
	if message == "" {
		respondError(s, i, "フィードバック内容を入力してください")
		return
	}

	// ユーザー情報を取得
	var userID, username string
	if isDM {
		userID = i.User.ID
		username = i.User.Username
	} else {
		userID = i.Member.User.ID
		username = getDisplayName(i.Member)
	}

	// 環境変数からフィードバックチャンネルIDを取得
	feedbackChannelID := os.Getenv("FEEDBACK_CHANNEL_ID")
	if feedbackChannelID == "" {
		respondError(s, i, "フィードバックチャンネルが設定されていません。管理者に連絡してください。")
		logger.LogCommand("feedback", userID, username, i.ChannelID, false, "FEEDBACK_CHANNEL_ID not set", map[string]interface{}{"message_length": len(message)})
		return
	}

	// タイムスタンプを生成
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// フィードバックチャンネルに匿名で転送
	feedbackEmbed := &discordgo.MessageEmbed{
		Title:       "💬 新しいフィードバック",
		Description: message,
		Color:       0x5865F2, // Discord Blurple
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("部室予約システム  |  feedback  |  受信日時: %s", timestamp),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err := s.ChannelMessageSendEmbed(feedbackChannelID, feedbackEmbed)
	if err != nil {
		respondError(s, i, "フィードバックの送信に失敗しました。管理者に連絡してください。")
		logger.LogCommand("feedback", userID, username, i.ChannelID, false, fmt.Sprintf("Failed to send feedback: %v", err), map[string]interface{}{"message_length": len(message)})
		return
	}

	// 送信者に確認メッセージを表示（自分だけに見える）
	respondEmbed(s, i, "✅ フィードバックを送信しました",
		"ご意見ありがとうございます。\nあなたのフィードバックは匿名で運営チームに届けられました。\n\n今後のシステム改善に活用させていただきます。",
		0x57F287, true)

	// ログに記録（メッセージの長さのみ記録、内容は記録しない）
	logger.LogCommand("feedback", userID, username, i.ChannelID, true, "", map[string]interface{}{"message_length": len(message)})
}
