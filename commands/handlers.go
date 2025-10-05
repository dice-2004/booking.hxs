package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/models"
	"github.com/dice/hxs_reservation_system/storage"
)

// HandleInteraction はDiscordのインタラクションを処理する
func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage) {
	switch i.ApplicationCommandData().Name {
	case "reserve":
		handleReserve(s, i, store)
	case "cancel":
		handleCancel(s, i, store)
	case "complete":
		handleComplete(s, i, store)
	case "list":
		handleList(s, i, store)
	case "my-reservations":
		handleMyReservations(s, i, store)
	}
}

// handleReserve は予約作成コマンドを処理する
func handleReserve(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
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

	// 日付と時間の形式を検証
	if _, err := time.Parse("2006-01-02", date); err != nil {
		respondError(s, i, "日付の形式が正しくありません（YYYY-MM-DD形式で入力してください）")
		return
	}

	if _, err := time.Parse("15:04", startTime); err != nil {
		respondError(s, i, "開始時間の形式が正しくありません（HH:MM形式で入力してください）")
		return
	}

	if _, err := time.Parse("15:04", endTime); err != nil {
		respondError(s, i, "終了時間の形式が正しくありません（HH:MM形式で入力してください）")
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
		UserID:    i.Member.User.ID,
		Username:  getDisplayName(i.Member),
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
		Comment:   comment,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ChannelID: i.ChannelID,
	}

	// 時間の重複をチェック
	overlappingReservation, err := store.CheckOverlap(reservation)
	if err != nil {
		respondError(s, i, "予約の重複チェックに失敗しました")
		return
	}

	if overlappingReservation != nil {
		msg := fmt.Sprintf("❌ **予約できませんでした**\n\n"+
			"指定された時間は既に予約されています。\n\n"+
			"**重複している予約:**\n"+
			"予約者: <@%s>\n"+
			"日時: %s %s - %s",
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
		return
	}

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		return
	}

	// 予約者にはIDを含めたメッセージを送信（Ephemeral）
	ephemeralMsg := fmt.Sprintf("✅ **予約が完了しました！**\n\n"+
		"**予約ID:** `%s`\n"+
		"日時: %s %s - %s\n"+
		"%s\n\n"+
		"※予約IDは取り消しや完了の際に必要です。大切に保管してください。",
		reservation.ID,
		formatDate(reservation.Date),
		reservation.StartTime,
		reservation.EndTime,
		formatComment(comment),
	)
	respondEphemeral(s, i, ephemeralMsg)

	// チャンネルの全員に予約情報を通知（予約IDは含めない）
	publicMsg := fmt.Sprintf("📅 **新しい予約が追加されました**\n\n"+
		"予約者: <@%s>\n"+
		"日時: %s %s - %s\n"+
		"%s",
		reservation.UserID,
		formatDate(reservation.Date),
		reservation.StartTime,
		reservation.EndTime,
		formatComment(comment),
	)
	s.ChannelMessageSend(i.ChannelID, publicMsg)
}

// handleCancel は予約キャンセルコマンドを処理する
func handleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage) {
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
		return
	}

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		return
	}

	// 応答
	respondEphemeral(s, i, "✅ 予約を取り消しました")

	// チャンネルの全員に通知
	msg := fmt.Sprintf("🚫 **予約が取り消されました**\n\n"+
		"予約者: <@%s>\n"+
		"日時: %s %s - %s\n"+
		"%s",
		reservation.UserID,
		formatDate(reservation.Date),
		reservation.StartTime,
		reservation.EndTime,
		formatComment(comment),
	)
	s.ChannelMessageSend(i.ChannelID, msg)
}

// handleComplete は予約完了コマンドを処理する
func handleComplete(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage) {
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
		return
	}

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		return
	}

	// 応答
	respondEphemeral(s, i, "✅ 予約を完了にしました")

	// チャンネルの全員に通知
	msg := fmt.Sprintf("✨ **予約が完了しました**\n\n"+
		"予約者: <@%s>\n"+
		"日時: %s %s - %s\n"+
		"%s",
		reservation.UserID,
		formatDate(reservation.Date),
		reservation.StartTime,
		reservation.EndTime,
		formatComment(comment),
	)
	s.ChannelMessageSend(i.ChannelID, msg)
}

// handleList はすべての予約一覧を表示する
func handleList(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage) {
	reservations := store.GetAllReservations()

	if len(reservations) == 0 {
		respondEphemeral(s, i, "現在、予約はありません。")
		return
	}

	// 日時でソート
	sort.Slice(reservations, func(a, b int) bool {
		dateTimeA := reservations[a].Date + " " + reservations[a].StartTime
		dateTimeB := reservations[b].Date + " " + reservations[b].StartTime
		return dateTimeA < dateTimeB
	})

	// メッセージを構築
	var sb strings.Builder
	sb.WriteString("📋 **すべての予約一覧**\n\n")

 	for _, r := range reservations {
 		statusEmoji := getStatusEmoji(r.Status)
 		sb.WriteString(fmt.Sprintf("%s **%s %s - %s**\n", statusEmoji, formatDate(r.Date), r.StartTime, r.EndTime))
 		sb.WriteString(fmt.Sprintf("   予約者: <@%s>\n", r.UserID))
 		sb.WriteString(fmt.Sprintf("   予約ID: `%s`\n", r.ID))
 		if r.Comment != "" {
 			sb.WriteString(fmt.Sprintf("   コメント: %s\n", r.Comment))
 		}
 		sb.WriteString(fmt.Sprintf("   状態: %s\n\n", getStatusText(r.Status)))
 	}

	respondEphemeral(s, i, sb.String())
}

// handleMyReservations は自分の予約一覧を表示する
func handleMyReservations(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage) {
	userID := i.Member.User.ID
	reservations := store.GetUserReservations(userID)

	if len(reservations) == 0 {
		respondEphemeral(s, i, "あなたの予約はありません。")
		return
	}

	// 日時でソート
	sort.Slice(reservations, func(a, b int) bool {
		dateTimeA := reservations[a].Date + " " + reservations[a].StartTime
		dateTimeB := reservations[b].Date + " " + reservations[b].StartTime
		return dateTimeA < dateTimeB
	})

	// メッセージを構築
	var sb strings.Builder
	sb.WriteString("📋 **あなたの予約一覧**\n\n")

 	for _, r := range reservations {
 		statusEmoji := getStatusEmoji(r.Status)
 		sb.WriteString(fmt.Sprintf("%s **%s %s - %s**\n", statusEmoji, formatDate(r.Date), r.StartTime, r.EndTime))
 		sb.WriteString(fmt.Sprintf("   予約者: <@%s>\n", r.UserID))
 		sb.WriteString(fmt.Sprintf("   予約ID: `%s`\n", r.ID))
 		if r.Comment != "" {
 			sb.WriteString(fmt.Sprintf("   コメント: %s\n", r.Comment))
 		}
 		sb.WriteString(fmt.Sprintf("   状態: %s\n\n", getStatusText(r.Status)))
 	}

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
	// YYYY-MM-DD を YYYY/MM/DD に変換
	return strings.ReplaceAll(date, "-", "/")
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
