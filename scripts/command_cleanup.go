package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// このスクリプトはDiscordのアプリコマンドを強制的に再登録します
// DMでコマンドが表示されない、または古い状態の場合に使用してください

func main() {
	// .envファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Discordトークンを取得
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is not set in environment variables")
	}

	guildID := os.Getenv("GUILD_ID")

	// Discord セッションを作成
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	// Discordに接続
	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer dg.Close()

	log.Println("Connected to Discord")

	// 既存のコマンドを全て削除
	log.Println("Deleting existing commands...")

	// グローバルコマンドを削除
	globalCommands, err := dg.ApplicationCommands(dg.State.User.ID, "")
	if err != nil {
		log.Printf("Failed to get global commands: %v", err)
	} else {
		for _, cmd := range globalCommands {
			err := dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID)
			if err != nil {
				log.Printf("Failed to delete global command '%s': %v", cmd.Name, err)
			} else {
				log.Printf("Deleted global command: %s", cmd.Name)
			}
		}
	}

	// ギルド固有のコマンドを削除（GUILD_IDが設定されている場合）
	if guildID != "" {
		guildCommands, err := dg.ApplicationCommands(dg.State.User.ID, guildID)
		if err != nil {
			log.Printf("Failed to get guild commands: %v", err)
		} else {
			for _, cmd := range guildCommands {
				err := dg.ApplicationCommandDelete(dg.State.User.ID, guildID, cmd.ID)
				if err != nil {
					log.Printf("Failed to delete guild command '%s': %v", cmd.Name, err)
				} else {
					log.Printf("Deleted guild command: %s", cmd.Name)
				}
			}
		}
	}

	log.Println("All existing commands deleted")
	log.Println("")
	log.Println("Registering new commands...")

	// 新しいコマンドを登録
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "reserve",
			Description: "部室の予約を作成します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "date",
					Description: "予約日（YYYY-MM-DD または YYYY/MM/DD、例: 2025-10-15 または 2025/10/15）",
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

	// コマンドをグローバルに登録（DMで使用可能にする）
	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("❌ Failed to create global command '%s': %v", cmd.Name, err)
		} else {
			log.Printf("✅ Registered global command: %s", cmd.Name)
		}
	}

	// ギルド固有のコマンドも登録（GUILD_IDが設定されている場合）
	if guildID != "" {
		log.Println("")
		log.Println("Registering guild-specific commands...")
		for _, cmd := range commands {
			_, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, cmd)
			if err != nil {
				log.Printf("❌ Failed to create guild command '%s': %v", cmd.Name, err)
			} else {
				log.Printf("✅ Registered guild command: %s", cmd.Name)
			}
		}
	}

	log.Println("")
	log.Println("✅ Command refresh completed!")
	log.Println("")
	log.Println("注意:")
	log.Println("- グローバルコマンドの反映には最大1時間かかる場合があります")
	log.Println("- ギルド固有のコマンドは即座に反映されます")
	log.Println("- Discordクライアントを再起動すると反映が早まる場合があります")
}
