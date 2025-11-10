package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/logging"
	"github.com/dice/hxs_reservation_system/models"
	"github.com/dice/hxs_reservation_system/storage"
)

// HandleInteraction はDiscordのインタラクションを処理する
func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string) {
	commandName := i.ApplicationCommandData().Name

	// DMかどうかを判定
	isDM := i.GuildID == ""

	// チャンネルIDを取得
	channelID := i.ChannelID

	// ユーザー情報を取得（DMの場合とサーバーの場合で取得方法が異なる）
	var userID, username string
	if isDM {
		userID = i.User.ID
		username = i.User.Username
	} else {
		userID = i.Member.User.ID
		username = getDisplayName(i.Member)
	}

	// チャンネル制限チェック（DMは除く）
	if !isDM && allowedChannelID != "" && channelID != allowedChannelID {
		respondEphemeral(s, i, "❌ このコマンドは指定されたチャンネルでのみ使用できます。")
		logger.LogCommand(commandName, userID, username, channelID, false, "Not allowed channel", nil)
		return
	}

	// コマンドパラメータを取得
	parameters := make(map[string]interface{})
	for _, opt := range i.ApplicationCommandData().Options {
		parameters[opt.Name] = opt.Value
	}

	// コマンド実行開始をログに記録
	logger.LogCommand(commandName, userID, username, channelID, true, "", parameters)

	switch commandName {
	case "reserve":
		handleReserve(s, i, store, logger, allowedChannelID, isDM)
	case "cancel":
		handleCancel(s, i, store, logger, allowedChannelID, isDM)
	case "complete":
		handleComplete(s, i, store, logger, allowedChannelID, isDM)
	case "list":
		handleList(s, i, store, logger, isDM)
	case "my-reservations":
		handleMyReservations(s, i, store, logger, isDM)
	case "help":
		handleHelp(s, i, logger, isDM)
	case "feedback":
		handleFeedback(s, i, logger, isDM)
	}
}

// handleReserve は予約作成コマンドを処理する
func handleReserve(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
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

	// 必須パラメータを取得
	date := optionMap["date"].StringValue()
	startTime := optionMap["start_time"].StringValue()

	// オプションパラメータを取得
	var endTime string
	if opt, ok := optionMap["end_time"]; ok {
		endTime = opt.StringValue()
	} else {
		// 終了時間が指定されていない場合は開始時刻+1時間
		start, err := time.Parse("15:04", startTime)
		if err != nil {
			respondError(s, i, "開始時間の形式が正しくありません（HH:MM形式で入力してください）")
			return
		}
		endTime = start.Add(1 * time.Hour).Format("15:04")
	}

	comment := ""
	if opt, ok := optionMap["comment"]; ok {
		comment = opt.StringValue()
	}

	// ログ用パラメータを構築
	parameters := map[string]interface{}{
		"date":       date,
		"start_time": startTime,
		"end_time":   endTime,
	}
	if comment != "" {
		parameters["comment"] = comment
	}

	// 日付と時間の形式を検証（YYYY-MM-DD または YYYY/MM/DD を許可）
	if _, err := time.Parse("2006-01-02", date); err != nil {
		if t2, err2 := time.Parse("2006/01/02", date); err2 == nil {
			// 正規化して保存用は YYYY-MM-DD に統一
			date = t2.Format("2006-01-02")
		} else {
			errorMsg := "日付の形式が正しくありません（YYYY-MM-DD または YYYY/MM/DD）"
			logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
			respondError(s, i, errorMsg)
			return
		}
	}

	if _, err := time.Parse("15:04", startTime); err != nil {
		errorMsg := "開始時間の形式が正しくありません（HH:MM形式で入力してください）"
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
		respondError(s, i, errorMsg)
		return
	}

	if _, err := time.Parse("15:04", endTime); err != nil {
		errorMsg := "終了時間の形式が正しくありません（HH:MM形式で入力してください）"
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
		respondError(s, i, errorMsg)
		return
	}

	// 予約IDを生成
	reservationID, err := models.GenerateReservationID()
	if err != nil {
		respondError(s, i, "予約IDの生成に失敗しました")
		return
	}

	// 予約を作成
	reservation := &models.Reservation{
		ID:        reservationID,
		UserID:    userID,
		Username:  username,
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
		Comment:   comment,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ChannelID: allowedChannelID, // 公開メッセージの送信先は常に指定チャンネル
	}

	// 時間の重複をチェック
	overlappingReservation, err := store.CheckOverlap(reservation)
	if err != nil {
		respondError(s, i, "予約の重複チェックに失敗しました")
		logger.LogError("ERROR", "handlers.handleReserve", "Failed to check overlap", err, map[string]interface{}{
			"user_id": userID,
			"date":    date,
		})
		return
	}

	if overlappingReservation != nil {
		msg := fmt.Sprintf("❌ **予約できませんでした**\n\n"+
			"指定された時間は既に予約されています。\n\n"+
			"**重複している予約:**\n"+
			"👤   <@%s>\n"+
			"📅   %s %s - %s",
			overlappingReservation.UserID,
			formatDate(overlappingReservation.Date),
			overlappingReservation.StartTime,
			overlappingReservation.EndTime,
		)
		respondEphemeral(s, i, msg)
		return
	}

	// 予約を保存
	if err := store.AddReservation(reservation); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		logger.LogError("ERROR", "handlers.handleReserve", "Failed to add reservation", err, map[string]interface{}{
			"user_id":        userID,
			"reservation_id": reservation.ID,
		})
		return
	}

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		logger.LogError("ERROR", "handlers.handleReserve", "Failed to save reservations", err, map[string]interface{}{
			"user_id":        userID,
			"reservation_id": reservation.ID,
		})
		return
	}

	// 予約者にはIDを含めたメッセージを送信（Ephemeral）
	ephemeralMsg := fmt.Sprintf("✅ **予約が完了しました！**\n\n"+
		"**予約ID:** `%s`\n"+
		"📅   %s %s - %s\n"+
		"%s\n\n"+
		"※予約IDは取り消しや完了の際に必要です。大切に保管してください。\nお忘れの際には、`/my-reservations` コマンドで確認できます。",
		reservation.ID,
		formatDate(reservation.Date),
		reservation.StartTime,
		reservation.EndTime,
		formatComment(comment),
	)
	respondEphemeral(s, i, ephemeralMsg)

	// チャンネルの全員に予約情報を通知（予約IDは含めない）
	publicMsg := fmt.Sprintf("🟡 **新しい予約が追加されました**\n\n"+
		"👤   <@%s>\n"+
		"📅   %s %s - %s\n"+
		"%s",
		reservation.UserID,
		formatDate(reservation.Date),
		reservation.StartTime,
		reservation.EndTime,
		formatComment(comment),
	)
	// DMから実行された場合も、指定チャンネルに通知
	s.ChannelMessageSend(allowedChannelID, publicMsg)
}

// handleCancel は予約キャンセルコマンドを処理する
func handleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	reservationID := optionMap["reservation_id"].StringValue()

	comment := ""
	if opt, ok := optionMap["comment"]; ok {
		comment = opt.StringValue()
	}

	// 予約を取得
	reservation, err := store.GetReservation(reservationID)
	if err != nil {
		respondError(s, i, "予約が見つかりませんでした。予約IDを確認してください。")
		return
	}

	// 予約をキャンセル済みに更新
	reservation.Status = models.StatusCancelled
	reservation.UpdatedAt = time.Now()

	if err := store.UpdateReservation(reservation); err != nil {
		respondError(s, i, "予約の更新に失敗しました")
		logger.LogError("ERROR", "handlers.handleCancel", "Failed to update reservation", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		logger.LogError("ERROR", "handlers.handleCancel", "Failed to save reservations", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	// 応答
	respondEphemeral(s, i, "✅ 予約を取り消しました")

	// チャンネルの全員に通知
	msg := fmt.Sprintf("🔴 **予約が取り消されました**\n\n"+
		"👤   <@%s>\n"+
		"📅   %s %s - %s\n"+
		"%s",
		reservation.UserID,
		formatDate(reservation.Date),
		reservation.StartTime,
		reservation.EndTime,
		formatComment(comment),
	)
	// DMから実行された場合も、指定チャンネルに通知
	s.ChannelMessageSend(allowedChannelID, msg)
}

// handleComplete は予約完了コマンドを処理する
func handleComplete(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	reservationID := optionMap["reservation_id"].StringValue()

	comment := ""
	if opt, ok := optionMap["comment"]; ok {
		comment = opt.StringValue()
	}

	// 予約を取得
	reservation, err := store.GetReservation(reservationID)
	if err != nil {
		respondError(s, i, "予約が見つかりませんでした。予約IDを確認してください。")
		return
	}

	// 予約を完了に更新
	reservation.Status = models.StatusCompleted
	reservation.UpdatedAt = time.Now()

	if err := store.UpdateReservation(reservation); err != nil {
		respondError(s, i, "予約の更新に失敗しました")
		logger.LogError("ERROR", "handlers.handleComplete", "Failed to update reservation", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		logger.LogError("ERROR", "handlers.handleComplete", "Failed to save reservations", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	// 応答
	respondEphemeral(s, i, "✅ 予約を完了にしました")

	// チャンネルの全員に通知
	msg := fmt.Sprintf("🔴 **予約が終わりました**\n\n"+
		"👤   <@%s>\n"+
		"📅   %s %s - %s\n"+
		"%s",
		reservation.UserID,
		formatDate(reservation.Date),
		reservation.StartTime,
		reservation.EndTime,
		formatComment(comment),
	)
	// DMから実行された場合も、指定チャンネルに通知
	s.ChannelMessageSend(allowedChannelID, msg)
}

// handleList はすべての予約一覧を表示する
func handleList(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, isDM bool) {

	allReservations := store.GetAllReservations()
	// 完了・キャンセル済みを除外
	reservations := make([]*models.Reservation, 0)
	for _, r := range allReservations {
		if r.Status != models.StatusCompleted && r.Status != models.StatusCancelled {
			reservations = append(reservations, r)
		}
	}

	if len(reservations) == 0 {
		respondEphemeral(s, i, "現在、予約はありません。")
		return
	}

	// 日時でソート
	sort.Slice(reservations, func(a, b int) bool {
		tA, errA := reservations[a].GetStartDateTime()
		tB, errB := reservations[b].GetStartDateTime()
		if errA != nil || errB != nil {
			// エラー時は元の順序
			return a < b
		}
		return tA.Before(tB)
	})

	// メッセージを構築
	var sb strings.Builder
	sb.WriteString("🔵 **すべての予約一覧**\n\n")
	for _, r := range reservations {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%s    **%s %s - %s**\n", getStatusEmoji(r.Status), formatDate(r.Date), r.StartTime, r.EndTime))
		sb.WriteString(fmt.Sprintf("👤   <@%s>\n", r.UserID))
		if r.Comment != "" {
			sb.WriteString(fmt.Sprintf("💬   %s\n", r.Comment))
		} else {
			sb.WriteString("💬   ----\n")
		}
	}
	// sb.WriteString("────────────────────────────\n")

	respondEphemeral(s, i, sb.String())
}

// handleMyReservations は自分の予約一覧を表示する
func handleMyReservations(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, isDM bool) {
	var userID string
	if isDM {
		userID = i.User.ID
	} else {
		userID = i.Member.User.ID
	}

	allReservations := store.GetUserReservations(userID)
	// 完了・キャンセル済みを除外
	reservations := make([]*models.Reservation, 0)
	for _, r := range allReservations {
		if r.Status != models.StatusCompleted && r.Status != models.StatusCancelled {
			reservations = append(reservations, r)
		}
	}

	if len(reservations) == 0 {
		respondEphemeral(s, i, "あなたの予約はありません。")
		return
	}

	// 日時でソート
	sort.Slice(reservations, func(a, b int) bool {
		tA, errA := reservations[a].GetStartDateTime()
		tB, errB := reservations[b].GetStartDateTime()
		if errA != nil || errB != nil {
			// エラー時は元の順序
			return a < b
		}
		return tA.Before(tB)
	})

	// メッセージを構築
	var sb strings.Builder
	sb.WriteString("🔵 **あなたの予約一覧**\n\n")
	for _, r := range reservations {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%s    **%s %s - %s**\n", getStatusEmoji(r.Status), formatDate(r.Date), r.StartTime, r.EndTime))
		sb.WriteString(fmt.Sprintf("👤   <@%s>\n", r.UserID))
		sb.WriteString(fmt.Sprintf("🆔    `%s`\n", r.ID))
		if r.Comment != "" {
			sb.WriteString(fmt.Sprintf("💬   %s\n", r.Comment))
		} else {
			sb.WriteString("💬   ----\n")
		}
	}
	// sb.WriteString("────────────────────────────\n")

	respondEphemeral(s, i, sb.String())
}

// ヘルパー関数

func getDisplayName(member *discordgo.Member) string {
	if member.Nick != "" {
		return member.Nick
	}
	return member.User.Username
}

func formatComment(comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("コメント: %s", comment)
}

func formatDate(date string) string {
	// YYYY-MM-DD を YYYY/MM/DD に変換し、一桁の場合はゼロ埋め
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return date
	}
	year := parts[0]
	month := fmt.Sprintf("%02s", parts[1])
	day := fmt.Sprintf("%02s", parts[2])
	return fmt.Sprintf("%s/%s/%s", year, month, day)
}

func getStatusEmoji(status models.ReservationStatus) string {
	switch status {
	case models.StatusPending:
		return "📅"
	case models.StatusCompleted:
		return "✅"
	case models.StatusCancelled:
		return "🚫"
	default:
		return "❓"
	}
}

func getStatusText(status models.ReservationStatus) string {
	switch status {
	case models.StatusPending:
		return "予約中"
	case models.StatusCompleted:
		return "完了"
	case models.StatusCancelled:
		return "キャンセル済み"
	default:
		return "不明"
	}
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleHelp はヘルプコマンドを処理する（コマンドを打った人にしか見えない）
func handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate, logger *logging.Logger, isDM bool) {
	helpMessage := "📖 **面接予約システム - ヘルプ**\n\n" +
		"### 利用可能なコマンド:" +
		"**/reserve**\n" +
		"> 部室の予約を作成します\n" +
		"> • `date`: 予約日（YYYY-MM-DD または YYYY/MM/DD、例: 2025-10-15）\n" +
		"> • `start_time`: 開始時間（HH:MM形式、例: 14:00）\n" +
		"> • `end_time`: 終了時間（HH:MM形式、例: 15:00）※省略時は開始時刻+1時間\n" +
		"> • `comment`: コメント（任意）\n\n" +
		"**/cancel**\n" +
		"> 予約を取り消します\n" +
		"> • `reservation_id`: 予約ID\n" +
		"> • `comment`: コメント（任意）\n\n" +
		"**/complete**\n" +
		"> 予約を完了にします\n" +
		"> • `reservation_id`: 予約ID\n" +
		"> • `comment`: コメント（任意）\n\n" +
		"**/list**\n" +
		"> すべての予約を表示します（自分だけに表示されます）\n\n" +
		"**/my-reservations**\n" +
		"> 自分の予約を表示します（自分だけに表示されます）\n\n" +
		"**/feedback**\n" +
		"> システムへのご意見・ご要望を匿名で送信します\n" +
		"> • `message`: フィードバック内容\n\n" +
		"**/help**\n" +
		"> このヘルプメッセージを表示します\n\n" +
		"### プライバシー:" +
		"> • /list、/my-reservations、/help、/feedback は自分だけに表示されます\n" +
		"> • 予約作成時、予約IDは予約者だけに通知されます\n" +
		"> • フィードバックは完全に匿名で送信されます\n\n" +
		"### データ管理:" +
		"> • 完了・キャンセル済みの予約は30日後に自動削除されます\n" +
		"> • 期限切れの予約は毎日午前3時に自動完了されます\n\n"+
		"### 利用可能チャンネル:" +
		"> • https://discord.com/channels/1090816023965479035/1375843736864559195で利用が可能です\n" +
		"> • または、認証済みの場合のみDMでも利用可能です\n\n"

	respondEphemeral(s, i, helpMessage)

	// ログに記録
	logger.LogCommand("help", i.Member.User.ID, getDisplayName(i.Member), i.ChannelID, true, "", nil)
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
			Text: fmt.Sprintf("受信日時: %s | 匿名フィードバック", timestamp),
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
	confirmMessage := `✅ **フィードバックを送信しました**

ご意見ありがとうございます。
あなたのフィードバックは匿名で運営チームに届けられました。

今後のシステム改善に活用させていただきます。`

	respondEphemeral(s, i, confirmMessage)

	// ログに記録（メッセージの長さのみ記録、内容は記録しない）
	logger.LogCommand("feedback", userID, username, i.ChannelID, true, "", map[string]interface{}{"message_length": len(message)})
}
