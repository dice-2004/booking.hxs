package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

// handleReserve は予約作成コマンドを処理する
func handleReserve(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	// ユーザー情報を取得
	userID, username := getUserInfo(i, isDM)

	// 必須パラメータを取得
	date := optionMap["date"].StringValue()
	startTime := optionMap["start_time"].StringValue()

	// 日付を正規化（YYYY/M/D → YYYY/MM/DD）
	date = normalizeDate(date)

	// 時刻を正規化（H:MM → HH:MM）
	startTime = normalizeTime(startTime)

	// オプションパラメータを取得
	var endTime string
	if opt, ok := optionMap["end_time"]; ok {
		endTime = opt.StringValue()
		// 時刻を正規化（H:MM → HH:MM）
		endTime = normalizeTime(endTime)
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
	var reservationDate time.Time
	if parsedDate, err := time.Parse("2006-01-02", date); err != nil {
		if t2, err2 := time.Parse("2006/01/02", date); err2 == nil {
			// 正規化して保存用は YYYY-MM-DD に統一
			date = t2.Format("2006-01-02")
			reservationDate = t2
		} else {
			errorMsg := "日付の形式が正しくありません（YYYY-MM-DD または YYYY/MM/DD）"
			logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
			respondError(s, i, errorMsg)
			return
		}
	} else {
		reservationDate = parsedDate
	}

	var startTimeParsed time.Time
	if t, err := time.Parse("15:04", startTime); err != nil {
		errorMsg := "開始時間の形式が正しくありません（HH:MM形式で入力してください）"
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
		respondError(s, i, errorMsg)
		return
	} else {
		startTimeParsed = t
	}

	if _, err := time.Parse("15:04", endTime); err != nil {
		errorMsg := "終了時間の形式が正しくありません（HH:MM形式で入力してください）"
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
		respondError(s, i, errorMsg)
		return
	}

	// 終了時刻が開始時刻より前または同じ時刻でないかチェック
	if endTime <= startTime {
		errorMsg := fmt.Sprintf("❌ 終了時刻は開始時刻より後である必要があります\n\n"+
			"**開始時刻:** %s\n"+
			"**終了時刻:** %s\n\n"+
			"終了時刻を開始時刻より後の時刻に設定してください。",
			startTime,
			endTime,
		)
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, "End time before start time", parameters)
		respondEphemeral(s, i, errorMsg)
		return
	}

	// 過去日時のチェック
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	nowJST := time.Now().In(jst)

	// 予約日時を構築（日付 + 開始時刻）
	reservationDateTime := time.Date(
		reservationDate.Year(),
		reservationDate.Month(),
		reservationDate.Day(),
		startTimeParsed.Hour(),
		startTimeParsed.Minute(),
		0, 0, jst,
	)

	// 現在時刻より過去の場合はエラー
	if reservationDateTime.Before(nowJST) {
		errorMsg := fmt.Sprintf("❌ 過去の日時は予約できません\n\n"+
			"**指定された日時:** %s %s\n"+
			"**現在日時:** %s\n\n"+
			"現在時刻以降の日時を指定してください。",
			formatDate(date),
			startTime,
			nowJST.Format("2006-01-02 15:04"),
		)
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, "Past datetime", parameters)
		respondEphemeral(s, i, errorMsg)
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
		fields := []*discordgo.MessageEmbedField{
			{
				Name:   "📅 重複している予約",
				Value:  formatDate(overlappingReservation.Date),
				Inline: false,
			},
			{
				Name:   "👤 予約者",
				Value:  fmt.Sprintf("<@%s>", overlappingReservation.UserID),
				Inline: true,
			},
			{
				Name:   "🕐 時間",
				Value:  fmt.Sprintf("%s - %s", overlappingReservation.StartTime, overlappingReservation.EndTime),
				Inline: true,
			},
		}

		embed := &discordgo.MessageEmbed{
			Title:       "🔴 予約できませんでした",
			Description: "指定された時間は既に予約されています。",
			Fields:      fields,
			Color:       0xED4245, // Discord Red
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
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
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "予約ID",
			Value:  fmt.Sprintf("`%s`", reservation.ID),
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
	}
	if comment != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "💬 コメント",
			Value:  comment,
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🟢 予約が完了しました！",
		Description: "~※予約IDは取り消しや完了の際に必要です。大切に保管してください。\nお忘れの際には、`/my-reservations` コマンドで確認できます。~",
		Fields:      fields,
		Color:       0x57F287, // Discord Green
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "部室予約システム  |  reserve",
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	// チャンネルの全員に予約情報を通知（予約IDは含めない）
	publicEmbed := &discordgo.MessageEmbed{
		Title: "🟢 新しい予約が追加されました",
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
		Color:     0x57F287, // Discord Green
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "部室予約システム  |  reserve",
		},
	}
	if comment != "" {
		publicEmbed.Fields = append(publicEmbed.Fields, &discordgo.MessageEmbedField{
			Name:   "💬 コメント",
			Value:  comment,
			Inline: false,
		})
	}
	// DMから実行された場合も、指定チャンネルに通知
	s.ChannelMessageSendEmbed(allowedChannelID, publicEmbed)

	// Botステータスを更新
	if UpdateStatusCallback != nil {
		UpdateStatusCallback()
	}
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
	respondEmbed(s, i, "🔵 予約を完了にしました", fmt.Sprintf("予約ID: `%s`", reservationID), 0x5865F2, true)

	// チャンネルの全員に通知
	completeEmbed := &discordgo.MessageEmbed{
		Title: "🔵 予約が終わりました",
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
		Color:     0x5865F2, // Discord Blue
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "部室予約システム  |  complete",
		},
	}
	if comment != "" {
		completeEmbed.Fields = append(completeEmbed.Fields, &discordgo.MessageEmbedField{
			Name:   "💬 コメント",
			Value:  comment,
			Inline: false,
		})
	}
	// DMから実行された場合も、指定チャンネルに通知
	s.ChannelMessageSendEmbed(allowedChannelID, completeEmbed)

	// Botステータスを更新
	if UpdateStatusCallback != nil {
		UpdateStatusCallback()
	}
}

// handleEdit は予約編集コマンドを処理する
func handleEdit(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	// ユーザー情報を取得
	userID, username := getUserInfo(i, isDM)

	// 予約IDを取得
	reservationID := optionMap["reservation_id"].StringValue()

	// 予約を取得
	reservation, err := store.GetReservation(reservationID)
	if err != nil {
		respondError(s, i, "指定された予約が見つかりません。")
		return
	}

	// 予約の所有者チェック
	if reservation.UserID != userID {
		respondError(s, i, "他のユーザーの予約は編集できません。")
		return
	}

	// ステータスチェック
	if reservation.Status != models.StatusPending {
		respondError(s, i, "完了またはキャンセルされた予約は編集できません。")
		return
	}

	// 変更前の情報を保持
	oldDate := reservation.Date
	oldStartTime := reservation.StartTime
	oldEndTime := reservation.EndTime
	oldComment := reservation.Comment

	// 新しい値を取得（指定されていない場合は現在の値を保持）
	newDate := oldDate
	newStartTime := oldStartTime
	newEndTime := oldEndTime
	newComment := oldComment

	hasChanges := false

	// 日付の変更
	if opt, ok := optionMap["date"]; ok {
		dateStr := opt.StringValue()
		// 日付を正規化
		dateStr = normalizeDate(dateStr)

		// 日付の形式を検証
		var parsedDate time.Time
		if t, err := time.Parse("2006-01-02", dateStr); err != nil {
			if t2, err2 := time.Parse("2006/01/02", dateStr); err2 == nil {
				dateStr = t2.Format("2006-01-02")
				parsedDate = t2
			} else {
				respondError(s, i, "日付の形式が正しくありません（YYYY-MM-DD または YYYY/MM/DD 形式で入力してください）")
				return
			}
		} else {
			parsedDate = t
		}

		// 過去の日付チェック
		jst := time.FixedZone("Asia/Tokyo", 9*60*60)
		now := time.Now().In(jst)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)
		if parsedDate.Before(today) {
			respondError(s, i, "過去の日付には変更できません。")
			return
		}

		newDate = dateStr
		hasChanges = true
	}

	// 開始時間の変更
	if opt, ok := optionMap["start_time"]; ok {
		timeStr := opt.StringValue()
		// 時刻を正規化
		timeStr = normalizeTime(timeStr)

		if _, err := time.Parse("15:04", timeStr); err != nil {
			respondError(s, i, "開始時間の形式が正しくありません（HH:MM形式で入力してください）")
			return
		}
		newStartTime = timeStr
		hasChanges = true
	}

	// 終了時間の変更
	if opt, ok := optionMap["end_time"]; ok {
		timeStr := opt.StringValue()
		// 時刻を正規化
		timeStr = normalizeTime(timeStr)

		if _, err := time.Parse("15:04", timeStr); err != nil {
			respondError(s, i, "終了時間の形式が正しくありません（HH:MM形式で入力してください）")
			return
		}
		newEndTime = timeStr
		hasChanges = true
	}

	// コメントの変更
	if opt, ok := optionMap["comment"]; ok {
		newComment = opt.StringValue()
		hasChanges = true
	}

	// 変更がない場合
	if !hasChanges {
		respondError(s, i, "変更する項目を少なくとも1つ指定してください。")
		return
	}

	// 時刻の整合性チェック
	if newEndTime <= newStartTime {
		respondError(s, i, "終了時間は開始時間より後である必要があります。")
		return
	}

	// 重複チェック用に一時的な予約オブジェクトを作成
	tempReservation := &models.Reservation{
		ID:        reservationID, // 自分の予約は除外するためにIDを設定
		UserID:    userID,
		Username:  username,
		Date:      newDate,
		StartTime: newStartTime,
		EndTime:   newEndTime,
		Comment:   newComment,
		Status:    models.StatusPending,
	}

	// 時間の重複をチェック（自分の予約以外と）
	overlappingReservation, err := store.CheckOverlap(tempReservation)
	if err != nil {
		respondError(s, i, "予約の重複チェックに失敗しました")
		logger.LogError("ERROR", "handleEdit", "Failed to check overlap", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	if overlappingReservation != nil {
		fields := []*discordgo.MessageEmbedField{
			{
				Name:   "📅 日付",
				Value:  strings.ReplaceAll(newDate, "-", "/"),
				Inline: false,
			},
			{
				Name:   "👤 予約者",
				Value:  fmt.Sprintf("<@%s>", overlappingReservation.UserID),
				Inline: true,
			},
			{
				Name:   "🕐 時間",
				Value:  fmt.Sprintf("%s - %s", overlappingReservation.StartTime, overlappingReservation.EndTime),
				Inline: true,
			},
		}

		embed := &discordgo.MessageEmbed{
			Title:       "🔴 予約を編集できませんでした",
			Description: "指定された時間は既に予約されています。",
			Fields:      fields,
			Color:       0xED4245, // Discord Red
			Timestamp:   time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "部室予約システム  |  edit",
			},
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 予約を更新
	reservation.Date = newDate
	reservation.StartTime = newStartTime
	reservation.EndTime = newEndTime
	reservation.Comment = newComment

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の更新に失敗しました。")
		logger.LogError("ERROR", "handleEdit", "Failed to save reservation", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	// 成功メッセージ
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "🆔 予約ID",
			Value:  reservation.ID,
			Inline: false,
		},
	}

	// 変更内容を表示
	if oldDate != newDate {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "📅 日付",
			Value:  fmt.Sprintf("%s → %s", strings.ReplaceAll(oldDate, "-", "/"), strings.ReplaceAll(newDate, "-", "/")),
			Inline: false,
		})
	}
	if oldStartTime != newStartTime || oldEndTime != newEndTime {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "🕐 時間",
			Value:  fmt.Sprintf("%s-%s → %s-%s", oldStartTime, oldEndTime, newStartTime, newEndTime),
			Inline: false,
		})
	}
	if oldComment != newComment {
		oldCommentDisplay := oldComment
		if oldCommentDisplay == "" {
			oldCommentDisplay = "（なし）"
		}
		newCommentDisplay := newComment
		if newCommentDisplay == "" {
			newCommentDisplay = "（なし）"
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "💬 コメント",
			Value:  fmt.Sprintf("%s → %s", oldCommentDisplay, newCommentDisplay),
			Inline: false,
		})
	}

	respondEmbedWithFields(s, i, "🟡 予約を編集しました", "", fields, 0xFEE75C, true)

	// 公開通知(変更がある場合)
	if !isDM {
		editEmbed := &discordgo.MessageEmbed{
			Title:       "🟡 予約が編集されました",
			Description: fmt.Sprintf("<@%s> さんが予約を編集しました", userID),
			Fields:      fields,
			Color:       0xFEE75C, // Discord Yellow
			Timestamp:   time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "部室予約システム  |  edit",
			},
		}
		s.ChannelMessageSendEmbed(allowedChannelID, editEmbed)
	} else if allowedChannelID != "" {
		// DMから実行された場合も、指定チャンネルに通知
		editEmbed := &discordgo.MessageEmbed{
			Title:       "🟡 予約が編集されました",
			Description: fmt.Sprintf("%s さんが予約を編集しました", username),
			Fields:      fields,
			Color:       0xFEE75C, // Discord Yellow
			Timestamp:   time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "部室予約システム  |  edit",
			},
		}
		s.ChannelMessageSendEmbed(allowedChannelID, editEmbed)
	}

	// Botステータスを更新
	if UpdateStatusCallback != nil {
		UpdateStatusCallback()
	}
}
