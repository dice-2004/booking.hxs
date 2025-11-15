package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

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
	respondEmbed(s, i, "🔴 予約を取り消しました", fmt.Sprintf("予約ID: `%s`", reservationID), 0xED4245, true)

	// チャンネルの全員に通知
	cancelEmbed := &discordgo.MessageEmbed{
		Title: "🔴 予約が取り消されました",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "👤 予約者",
				Value:  fmt.Sprintf("<@%s>", reservation.UserID),
				Inline: false,
			},
			{
				Name:   "📅 日付",
				Value:  formatDate(reservation.Date),
				Inline: true,
			},
			{
				Name:   "🕐 時間",
				Value:  fmt.Sprintf("%s - %s", reservation.StartTime, reservation.EndTime),
				Inline: true,
			},
		},
		Color:     0xED4245, // Discord Red
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "部室予約システム  |  cancel",
		},
	}
	if comment != "" {
		cancelEmbed.Fields = append(cancelEmbed.Fields, &discordgo.MessageEmbedField{
			Name:   "💬 コメント",
			Value:  comment,
			Inline: false,
		})
	}
	// DMから実行された場合も、指定チャンネルに通知
	s.ChannelMessageSendEmbed(allowedChannelID, cancelEmbed)

	// Botステータスを更新
	if UpdateStatusCallback != nil {
		UpdateStatusCallback()
	}
}
