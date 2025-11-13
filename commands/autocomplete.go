package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/models"
	"github.com/dice/hxs_reservation_system/storage"
)

// HandleAutocomplete はオートコンプリートのリクエストを処理する
func HandleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage) {
	data := i.ApplicationCommandData()

	// 現在フォーカスされているオプションを取得
	var focusedOption *discordgo.ApplicationCommandInteractionDataOption
	for _, opt := range data.Options {
		if opt.Focused {
			focusedOption = opt
			break
		}
	}

	if focusedOption == nil {
		return
	}

	var choices []*discordgo.ApplicationCommandOptionChoice

	// コマンド名を取得
	commandName := data.Name

	switch focusedOption.Name {
	case "date":
		choices = getDateSuggestions(focusedOption.StringValue())
	case "start_time":
		choices = getTimeSuggestions(focusedOption.StringValue(), "")
	case "end_time":
		// end_timeの場合、start_timeを取得して考慮する
		var startTime string
		for _, opt := range data.Options {
			if opt.Name == "start_time" {
				startTime = opt.StringValue()
				break
			}
		}
		choices = getTimeSuggestions(focusedOption.StringValue(), startTime)
	case "reservation_id":
		// ユーザーIDを取得
		var userID string
		if i.Member != nil {
			userID = i.Member.User.ID
		} else if i.User != nil {
			userID = i.User.ID
		}

		// コマンドに応じて候補を生成
		if commandName == "cancel" || commandName == "complete" || commandName == "edit" {
			choices = getReservationSuggestions(store, userID, "pending", focusedOption.StringValue())
		}
	}

	// 最大25個まで
	if len(choices) > 25 {
		choices = choices[:25]
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})

	if err != nil {
		// Autocompleteのエラーはログのみ
		fmt.Printf("Failed to respond to autocomplete: %v\n", err)
	}
}

// getDateSuggestions は日付の候補を生成する
func getDateSuggestions(input string) []*discordgo.ApplicationCommandOptionChoice {
	now := time.Now()
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	nowJST := now.In(jst)

	// 入力が空の場合
	if input == "" {
		suggestions := []*discordgo.ApplicationCommandOptionChoice{
			{Name: "今日", Value: nowJST.Format("2006/01/02")},
			{Name: "明日", Value: nowJST.AddDate(0, 0, 1).Format("2006/01/02")},
			{Name: "明後日", Value: nowJST.AddDate(0, 0, 2).Format("2006/01/02")},
		}

		// 3日後から30日後まで
		for i := 3; i <= 30; i++ {
			if i%7 == 0 && i <= 28 {
				week := i / 7
				suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
					Name:  fmt.Sprintf("%d週間後", week),
					Value: nowJST.AddDate(0, 0, i).Format("2006/01/02"),
				})
			} else {
				suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
					Name:  nowJST.AddDate(0, 0, i).Format("2006/01/02"),
					Value: nowJST.AddDate(0, 0, i).Format("2006/01/02"),
				})
			}
		}
		return suggestions
	}

	// 年の候補を生成（"25"、"26"などの入力に対して"2025"、"2026"を提案）
	if len(input) == 2 {
		if yearNum, err := strconv.Atoi(input); err == nil {
			currentYear := nowJST.Year()
			currentCentury := (currentYear / 100) * 100
			fullYear := currentCentury + yearNum

			if fullYear < currentYear-10 {
				fullYear += 100
			}

			suggestions := []*discordgo.ApplicationCommandOptionChoice{}
			for month := 1; month <= 12 && len(suggestions) < 25; month++ {
				for day := 1; day <= 7 && len(suggestions) < 25; day++ {
					dateStr := fmt.Sprintf("%d/%02d/%02d", fullYear, month, day)
					suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
						Name:  dateStr,
						Value: dateStr,
					})
				}
			}

			if len(suggestions) > 0 {
				return suggestions
			}
		}
	}

	// 月の候補を生成
	if len(input) <= 2 {
		if monthNum, err := strconv.Atoi(input); err == nil && monthNum >= 1 && monthNum <= 12 {
			year := nowJST.Year()
			suggestions := []*discordgo.ApplicationCommandOptionChoice{}
			for yearOffset := 0; yearOffset <= 1 && len(suggestions) < 25; yearOffset++ {
				targetYear := year + yearOffset
				daysInMonth := time.Date(targetYear, time.Month(monthNum+1), 0, 0, 0, 0, 0, jst).Day()

				for day := 1; day <= daysInMonth && len(suggestions) < 25; day++ {
					dateStr := fmt.Sprintf("%d/%02d/%02d", targetYear, monthNum, day)
					suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
						Name:  dateStr,
						Value: dateStr,
					})
				}
			}

			if len(suggestions) > 0 {
				return suggestions
			}
		}
	}

	// 日の候補を生成
	if len(input) <= 2 {
		if dayNum, err := strconv.Atoi(input); err == nil && dayNum >= 1 && dayNum <= 31 {
			year := nowJST.Year()
			month := int(nowJST.Month())
			suggestions := []*discordgo.ApplicationCommandOptionChoice{}

			for monthOffset := 0; monthOffset <= 3 && len(suggestions) < 25; monthOffset++ {
				targetMonth := month + monthOffset
				targetYear := year

				if targetMonth > 12 {
					targetYear++
					targetMonth -= 12
				}

				daysInMonth := time.Date(targetYear, time.Month(targetMonth+1), 0, 0, 0, 0, 0, jst).Day()

				if dayNum <= daysInMonth {
					dateStr := fmt.Sprintf("%d/%02d/%02d", targetYear, targetMonth, dayNum)
					suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
						Name:  dateStr,
						Value: dateStr,
					})
				}
			}

			if len(suggestions) > 0 {
				return suggestions
			}
		}
	}

	// 通常のフィルタリング処理
	allSuggestions := []*discordgo.ApplicationCommandOptionChoice{
		{Name: "今日", Value: nowJST.Format("2006/01/02")},
		{Name: "明日", Value: nowJST.AddDate(0, 0, 1).Format("2006/01/02")},
		{Name: "明後日", Value: nowJST.AddDate(0, 0, 2).Format("2006/01/02")},
	}

	for i := 3; i <= 30; i++ {
		if i%7 == 0 && i <= 28 {
			week := i / 7
			allSuggestions = append(allSuggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  fmt.Sprintf("%d週間後", week),
				Value: nowJST.AddDate(0, 0, i).Format("2006/01/02"),
			})
		} else {
			allSuggestions = append(allSuggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  nowJST.AddDate(0, 0, i).Format("2006/01/02"),
				Value: nowJST.AddDate(0, 0, i).Format("2006/01/02"),
			})
		}
	}

	// 入力でフィルタリング
	suggestions := []*discordgo.ApplicationCommandOptionChoice{}
	for _, choice := range allSuggestions {
		if strings.Contains(choice.Value.(string), input) || strings.Contains(choice.Name, input) {
			suggestions = append(suggestions, choice)
			if len(suggestions) >= 25 {
				break
			}
		}
	}

	if len(suggestions) > 0 {
		return suggestions
	}

	return allSuggestions
}

// getTimeSuggestions は時刻の候補を生成する
func getTimeSuggestions(input string, startTime string) []*discordgo.ApplicationCommandOptionChoice {
	suggestions := []*discordgo.ApplicationCommandOptionChoice{
		{Name: "09:00", Value: "09:00"},
		{Name: "09:30", Value: "09:30"},
		{Name: "10:00", Value: "10:00"},
		{Name: "10:30", Value: "10:30"},
		{Name: "11:00", Value: "11:00"},
		{Name: "11:30", Value: "11:30"},
		{Name: "12:00", Value: "12:00"},
		{Name: "12:30", Value: "12:30"},
		{Name: "13:00", Value: "13:00"},
		{Name: "13:30", Value: "13:30"},
		{Name: "14:00", Value: "14:00"},
		{Name: "14:30", Value: "14:30"},
		{Name: "15:00", Value: "15:00"},
		{Name: "15:30", Value: "15:30"},
		{Name: "16:00", Value: "16:00"},
		{Name: "16:30", Value: "16:30"},
		{Name: "17:00", Value: "17:00"},
		{Name: "17:30", Value: "17:30"},
		{Name: "18:00", Value: "18:00"},
		{Name: "18:30", Value: "18:30"},
		{Name: "19:00", Value: "19:00"},
		{Name: "19:30", Value: "19:30"},
		{Name: "20:00", Value: "20:00"},
		{Name: "20:30", Value: "20:30"},
		{Name: "21:00", Value: "21:00"},
	}

	// end_timeの場合、start_timeより後の時刻のみフィルタリング
	if startTime != "" {
		var filtered []*discordgo.ApplicationCommandOptionChoice
		for _, choice := range suggestions {
			if choice.Value.(string) > startTime {
				filtered = append(filtered, choice)
			}
		}
		suggestions = filtered
	}

	// 入力がある場合、さらにフィルタリング
	if input != "" {
		var filtered []*discordgo.ApplicationCommandOptionChoice
		for _, choice := range suggestions {
			if strings.HasPrefix(choice.Value.(string), input) {
				filtered = append(filtered, choice)
			}
		}
		if len(filtered) > 0 {
			return filtered
		}
	}

	return suggestions
}

// getReservationSuggestions はユーザーの予約候補を生成する
func getReservationSuggestions(store *storage.Storage, userID string, status string, input string) []*discordgo.ApplicationCommandOptionChoice {
	suggestions := []*discordgo.ApplicationCommandOptionChoice{}
	reservations := store.GetUserReservations(userID)

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst)

	var filteredReservations []*models.Reservation
	for _, r := range reservations {
		if r.Status == models.ReservationStatus(status) {
			reservationDate, err := time.Parse("2006-01-02", r.Date)
			if err != nil {
				continue
			}

			if !reservationDate.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)) {
				filteredReservations = append(filteredReservations, r)
			}
		}
	}

	sort.Slice(filteredReservations, func(i, j int) bool {
		if filteredReservations[i].Date != filteredReservations[j].Date {
			return filteredReservations[i].Date < filteredReservations[j].Date
		}
		return filteredReservations[i].StartTime < filteredReservations[j].StartTime
	})

	for _, r := range filteredReservations {
		displayDate := strings.ReplaceAll(r.Date, "-", "/")
		name := fmt.Sprintf("%s %s-%s", displayDate, r.StartTime, r.EndTime)
		if r.Comment != "" {
			comment := r.Comment
			if len(comment) > 20 {
				comment = comment[:20] + "..."
			}
			name = fmt.Sprintf("%s (%s)", name, comment)
		}

		if input == "" || strings.Contains(r.ID, input) || strings.Contains(name, input) {
			suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  name,
				Value: r.ID,
			})
		}

		if len(suggestions) >= 25 {
			break
		}
	}

	return suggestions
}
