package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/commands"
	"github.com/dice/hxs_reservation_system/storage"
	"github.com/joho/godotenv"
)

var (
	store   *storage.Storage
	guildID string
)

func init() {
	// .envファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	guildID = os.Getenv("GUILD_ID")
}

func main() {
	// Storageの初期化
	store = storage.NewStorage()
	if err := store.Load(); err != nil {
		log.Fatalf("Failed to load reservations: %v", err)
	}
	log.Println("Reservations loaded successfully")

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

	// コマンドハンドラーを設定
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commands.HandleInteraction(s, i, store)
	})

	// 必要なIntentを設定
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	// Discordに接続
	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer dg.Close()

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
				log.Printf("Failed to save reservations: %v", err)
			} else {
				log.Println("Reservations saved successfully")
			}
		}
	}()

	// シグナルを待つ
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// 終了時にデータを保存
	log.Println("Saving reservations before exit...")
	if err := store.Save(); err != nil {
		log.Printf("Failed to save reservations: %v", err)
	} else {
		log.Println("Reservations saved successfully")
	}
}

func registerCommands(s *discordgo.Session) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "reserve",
			Description: "面接の予約を作成します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "date",
					Description: "予約日（YYYY-MM-DD形式、例: 2025-10-15）",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "start_time",
					Description: "開始時間（HH:MM形式、例: 14:00）",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "end_time",
					Description: "終了時間（HH:MM形式、例: 15:00）※省略時は開始時刻+1時間",
					Required:    false,
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
