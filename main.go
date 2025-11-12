package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/commands"
	"github.com/dice/hxs_reservation_system/logging"
	"github.com/dice/hxs_reservation_system/storage"
	"github.com/joho/godotenv"
)

var (
	store            *storage.Storage
	logger           *logging.Logger
	guildID          string
	allowedChannelID string
	// 同一Interactionの重複処理を防止
	processedInteractions sync.Map
)

func init() {
	// .envファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	guildID = os.Getenv("GUILD_ID")
	allowedChannelID = os.Getenv("ALLOWED_CHANNEL_ID")
}

func main() {
	// Storageの初期化
	store = storage.NewStorage()
	if err := store.Load(); err != nil {
		log.Fatalf("Failed to load reservations: %v", err)
	}
	log.Println("Reservations loaded successfully")

	// Loggerの初期化
	logger = logging.NewLogger("./logs")
	log.Println("Logger initialized successfully")

	// Discordトークンを取得
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is not set in environment variables")
	}

	// Discord セッションを作成
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	// コマンドハンドラーを設定（重複ガード付き）
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Interaction ID で一度きりにする
		if _, loaded := processedInteractions.LoadOrStore(i.ID, struct{}{}); loaded {
			return
		}
		
		// Autocomplete処理
		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			commands.HandleAutocomplete(s, i)
			return
		}
		
		// 通常のコマンド処理
		commands.HandleInteraction(s, i, store, logger, allowedChannelID)
	})

	// 必要なIntentを設定
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	// Discordに接続
	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer dg.Close()

	// Botのステータスを設定
	updateBotStatus(dg, store)

	// コマンドハンドラーでステータス更新できるようにコールバックを設定
	commands.UpdateStatusCallback = func() {
		updateBotStatus(dg, store)
	}

	log.Println("Bot is now running. Press CTRL+C to exit.")

	// コマンドを登録
	if err := registerCommands(dg); err != nil {
		log.Fatalf("Failed to register commands: %v", err)
	}

	// 定期的にデータを保存
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := store.Save(); err != nil {
				log.Printf("❌ Failed to save reservations: %v", err)
				logger.LogError("ERROR", "main.periodicSave", "Failed to save reservations", err, nil)
			} else {
				log.Println("💾 Reservations saved successfully")
			}
			// ステータスも更新
			updateBotStatus(dg, store)
		}
	}()

	// 定期的に古いログをクリーンアップ（1日1回）
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			logger.CleanupOldLogs()
		}
	}()

	// 定期的に期限切れ予約を自動完了（毎日午前3時）
	go func() {
		for {
			now := time.Now()
			// 次の午前3時を計算
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if !now.Before(next) {
				// 今日の3時を過ぎている場合は明日の3時
				next = next.Add(24 * time.Hour)
			}

			// 起動直後の場合は即座に実行、それ以外は次の3時まで待機
			if now.Hour() == 0 && now.Minute() < 5 {
				// 起動直後（深夜0時台の最初の5分間）なら即座に実行
				log.Println("Startup: Running initial cleanup tasks...")
			} else {
				// 次の実行時刻まで待機
				duration := time.Until(next)
				log.Printf("Next auto-complete scheduled at: %s (in %v)", next.Format("2006-01-02 15:04:05"), duration)
				time.Sleep(duration)
			}

			// 終了時刻が過ぎたpending予約を自動完了
			completedCount, err := store.AutoCompleteExpiredReservations()
			if err != nil {
				log.Printf("❌ Failed to auto-complete expired reservations: %v", err)
				logger.LogError("ERROR", "main.autoComplete", "Failed to auto-complete expired reservations", err, nil)
			} else if completedCount > 0 {
				log.Printf("✅ Auto-completed %d expired reservation(s) and saved", completedCount)
			} else {
				log.Println("✓ Auto-complete check completed: no expired reservations found")
			}
		}
	}() // 定期的に古い予約データをクリーンアップ（毎日午前3時10分）
	go func() {
		for {
			now := time.Now()
			// 次の午前3時10分を計算
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 10, 0, 0, now.Location())
			if !now.Before(next) {
				// 今日の3時10分を過ぎている場合は明日の3時10分
				next = next.Add(24 * time.Hour)
			}

			// 起動直後の場合は即座に実行、それ以外は次の3時10分まで待機
			if now.Hour() == 0 && now.Minute() < 5 {
				// 起動直後（深夜0時台の最初の5分間）なら即座に実行
				log.Println("Startup: Running initial cleanup tasks...")
			} else {
				// 次の実行時刻まで待機
				duration := time.Until(next)
				log.Printf("Next cleanup scheduled at: %s (in %v)", next.Format("2006-01-02 15:04:05"), duration)
				time.Sleep(duration)
			}

			// 古い完了済み・キャンセル済み予約を削除（30日以上前）
			deletedCount, err := store.CleanupOldReservations(30)
			if err != nil {
				log.Printf("❌ Failed to cleanup old reservations: %v", err)
				logger.LogError("ERROR", "main.cleanup", "Failed to cleanup old reservations", err, map[string]interface{}{
					"retention_days": 30,
				})
			} else if deletedCount > 0 {
				log.Printf("🗑️  Cleaned up %d old reservation(s) and saved", deletedCount)
			} else {
				log.Println("✓ Cleanup check completed: no old reservations to remove")
			}
		}
	}() // シグナルを待つ
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// 終了時にデータを保存
	log.Println("💾 Saving reservations before exit...")
	if err := store.Save(); err != nil {
		log.Printf("❌ Failed to save reservations: %v", err)
		logger.LogError("ERROR", "main.shutdown", "Failed to save reservations on shutdown", err, nil)
	} else {
		log.Println("✅ Reservations saved successfully")
	}

	// 統計情報を表示
	stats := logger.GetStats()
	log.Printf("=== コマンド統計 ===")
	log.Printf("総コマンド数: %d", stats.TotalCommands)
	log.Printf("コマンド別統計:")
	for cmd, count := range stats.CommandCounts {
		log.Printf("  %s: %d回", cmd, count)
	}
	log.Printf("ユーザー別統計:")
	for userID, count := range stats.UserCounts {
		log.Printf("  %s: %d回", userID, count)
	}
	log.Printf("最終更新: %s", stats.LastUpdated.Format("2006-01-02 15:04:05"))
}

// updateBotStatus はBotのステータスを更新する
func updateBotStatus(s *discordgo.Session, store *storage.Storage) {
	allReservations := store.GetAllReservations()
	pendingCount := 0
	for _, r := range allReservations {
		if r.Status == "pending" {
			pendingCount++
		}
	}

	var status string
	if pendingCount == 0 {
		status = "面接予約管理 | /help"
	} else {
		status = fmt.Sprintf("%d件の予約管理中 | /help", pendingCount)
	}

	if err := s.UpdateGameStatus(0, status); err != nil {
		log.Printf("Failed to update status: %v", err)
	}
}

func registerCommands(s *discordgo.Session) error {
	// 既存のコマンドを削除（重複を防ぐため）
	log.Println("Removing existing commands...")

	// グローバルコマンドを削除（もし存在すれば）
	globalCommands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		log.Printf("Failed to fetch existing global commands: %v", err)
	} else {
		for _, cmd := range globalCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
			if err != nil {
				log.Printf("Failed to delete global command %s: %v", cmd.Name, err)
			} else {
				log.Printf("Deleted existing global command: %s", cmd.Name)
			}
		}
	}

	// ギルド専用コマンドを削除（GUILD_IDが設定されている場合）
	if guildID != "" {
		guildCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
		if err != nil {
			log.Printf("Failed to fetch existing guild commands: %v", err)
		} else {
			for _, cmd := range guildCommands {
				err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
				if err != nil {
					log.Printf("Failed to delete guild command %s: %v", cmd.Name, err)
				} else {
					log.Printf("Deleted existing guild command: %s", cmd.Name)
				}
			}
		}
	}

	log.Println("Registering new commands...")

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "reserve",
			Description: "部室の予約を作成します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "date",
					Description:  "予約日（YYYY-MM-DD または YYYY/MM/DD、例: 2025-10-15 または 2025/10/15）",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "start_time",
					Description:  "開始時間（HH:MM形式、例: 14:00）",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "end_time",
					Description:  "終了時間（HH:MM形式、例: 15:00）※省略時は開始時刻+1時間",
					Required:     false,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comment",
					Description: "コメント（任意）",
					Required:    false,
				},
			},
		},
		{
			Name:        "cancel",
			Description: "予約を取り消します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "reservation_id",
					Description: "予約ID",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comment",
					Description: "コメント（任意）",
					Required:    false,
				},
			},
		},
		{
			Name:        "complete",
			Description: "予約を完了にします",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "reservation_id",
					Description: "予約ID",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comment",
					Description: "コメント（任意）",
					Required:    false,
				},
			},
		},
		{
			Name:        "list",
			Description: "すべての予約を表示します（自分だけに表示されます）",
		},
		{
			Name:        "my-reservations",
			Description: "自分の予約を表示します（自分だけに表示されます）",
		},
		{
			Name:        "help",
			Description: "ヘルプメッセージを表示します（自分だけに表示されます）",
		},
		{
			Name:        "feedback",
			Description: "システムへのご意見・ご要望を匿名で送信します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "フィードバック内容",
					Required:    true,
				},
			},
		},
	}

	for _, cmd := range commands {
		var err error
		if guildID != "" {
			_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		} else {
			_, err = s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		}
		if err != nil {
			return fmt.Errorf("cannot create '%s' command: %v", cmd.Name, err)
		}
		log.Printf("Registered command: %s", cmd.Name)
	}

	return nil
}
