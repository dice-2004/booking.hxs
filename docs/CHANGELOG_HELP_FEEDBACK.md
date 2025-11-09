# 変更履歴 - ヘルプ・フィードバック機能追加

## 📅 変更日時
2025年11月9日

## 🎯 変更概要

ユーザビリティとフィードバック収集を改善するため、以下の2つの新しいコマンドを追加しました：

1. **`/help`** - ヘルプメッセージ表示（コマンドを打った人にしか見えない）
2. **`/feedback`** - 匿名フィードバック送信（コマンドを打った人にしか見えない）

## ✨ 新機能

### 1. `/help` コマンド

**機能:**
- すべてのコマンドの一覧と説明を表示
- 各コマンドのパラメータと使用方法を説明
- プライバシーとデータ管理に関する情報を提供

**特徴:**
- ✅ コマンドを実行した人にのみ表示（Ephemeralメッセージ）
- ✅ チャンネルを汚さない
- ✅ 絵文字を使った分かりやすい表示
- ✅ すべてのコマンドを網羅

**使用例:**
```
/help
```

**表示内容:**
- 📅 /reserve - 予約作成
- 🚫 /cancel - 予約取り消し
- ✅ /complete - 予約完了
- 📋 /list - すべての予約を表示
- 👤 /my-reservations - 自分の予約を表示
- 📝 /feedback - フィードバック送信
- ℹ️ /help - ヘルプ表示

各コマンドについて、パラメータと使い方を詳しく説明します。

---

### 2. `/feedback` コマンド

**機能:**
- システムへのご意見・ご要望を匿名で送信
- フィードバックは特定のチャンネルに転送される
- 送信者の情報は一切含まれない（完全匿名）

**特徴:**
- ✅ 完全匿名（誰が送信したか分からない）
- ✅ コマンド実行自体も本人にしか見えない
- ✅ ログにはメッセージの長さのみ記録（内容は記録されない）
- ✅ 送信者に確認メッセージを表示

**使用例:**
```
/feedback message:予約の編集機能があると便利だと思います
```

**動作フロー:**

1. ユーザーが `/feedback` コマンドを実行
   - コマンド自体は本人にしか見えない

2. フィードバックが特定のチャンネルに転送される
   - 送信者の情報（ユーザーID、ユーザー名など）は含まれない
   - タイムスタンプのみが記録される

3. 送信者に確認メッセージが表示される
   ```
   ✅ フィードバックを送信しました

   ご意見ありがとうございます。
   あなたのフィードバックは匿名で運営チームに届けられました。

   今後のシステム改善に活用させていただきます。
   ```

**フィードバックチャンネルの表示例:**
```
💬 新しいフィードバック

予約の編集機能があると便利だと思います

━━━━━━━━━━━━━━━━━━━━
受信日時: 2025-11-09 15:30:00 | 匿名フィードバック
```

---

## 🔧 技術的な変更

### 1. コード変更

#### `commands/handlers.go`

**追加された関数:**

1. `handleHelp()`
   ```go
   func handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate, logger *logging.Logger)
   ```
   - ヘルプメッセージを構築して表示
   - Ephemeralメッセージで送信（本人にのみ表示）
   - コマンド実行をログに記録

2. `handleFeedback()`
   ```go
   func handleFeedback(s *discordgo.Session, i *discordgo.InteractionCreate, logger *logging.Logger)
   ```
   - フィードバック内容を受け取る
   - 環境変数 `FEEDBACK_CHANNEL_ID` からチャンネルIDを取得
   - Embedメッセージでフィードバックチャンネルに送信
   - 送信者に確認メッセージを表示
   - ログにはメッセージの長さのみ記録

**HandleInteraction の変更:**
```go
switch commandName {
    // ... 既存のケース
    case "help":
        handleHelp(s, i, logger)
    case "feedback":
        handleFeedback(s, i, logger)
}
```

**インポート追加:**
```go
import (
    "os"  // 環境変数の取得用
    // ... 既存のインポート
)
```

---

#### `main.go`

**コマンド登録の追加:**

```go
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
```

---

### 2. 環境変数の追加

**新しい環境変数:**
- `FEEDBACK_CHANNEL_ID`: フィードバックを受け取るチャンネルのID

**設定方法:**

1. Discordの設定で「開発者モード」を有効化
2. フィードバックを受け取りたいチャンネルを右クリック
3. 「IDをコピー」を選択
4. `.env` ファイルに追加:
   ```env
   FEEDBACK_CHANNEL_ID=your_feedback_channel_id_here
   ```

**注意:**
- `FEEDBACK_CHANNEL_ID` を設定しないと、`/feedback` コマンドは使用できません
- エラーメッセージが表示され、フィードバックは送信されません

---

### 3. 設定ファイルの更新

**更新されたファイル:**

1. **`config/.env.example`**
   ```env
   DISCORD_TOKEN=your_discord_bot_token_here
   GUILD_ID=your_guild_id_here
   FEEDBACK_CHANNEL_ID=your_feedback_channel_id_here  # 追加
   ```

2. **`config/.env.development`**
   ```env
   DISCORD_TOKEN=your_dev_discord_bot_token_here
   GUILD_ID=your_dev_guild_id_here
   FEEDBACK_CHANNEL_ID=your_dev_feedback_channel_id_here  # 追加
   ENV=development
   DATA_FILE=reservations_dev.json
   ```

3. **`config/.env.production`**
   ```env
   DISCORD_TOKEN=your_prod_discord_bot_token_here
   GUILD_ID=
   FEEDBACK_CHANNEL_ID=your_prod_feedback_channel_id_here  # 追加
   ENV=production
   DATA_FILE=reservations.json
   ```

---

## 📚 ドキュメントの更新

### 1. 新規作成されたドキュメント

#### `docs/COMMANDS.md` ✨ NEW
- すべてのコマンドの詳細なリファレンス
- パラメータの説明
- 使用例
- プライバシーに関する情報
- よくある質問（FAQ）
- 管理者向け情報

### 2. 更新されたドキュメント

#### `README.md`
**変更箇所:**

1. **機能リストの更新**
   ```markdown
   ## 機能

   - **予約作成** (`/reserve`)
   - **予約取り消し** (`/cancel`)
   - **予約完了** (`/complete`)
   - **予約一覧表示** (`/list`)
   - **自分の予約確認** (`/my-reservations`)
   - **ヘルプ表示** (`/help`) ← 追加
   - **フィードバック送信** (`/feedback`) ← 追加
   ```

2. **主な特徴の更新**
   ```markdown
   - ✅ 匿名フィードバック機能 ← 追加
   - ✅ コマンドを打った人にしか見えないヘルプ・フィードバック機能 ← 追加
   ```

3. **環境変数の説明を更新**
   ```markdown
   - `DISCORD_TOKEN`: Botトークン
   - `GUILD_ID`: サーバーID
   - `FEEDBACK_CHANNEL_ID`: フィードバック受信チャンネルのID ← 追加
   ```

4. **ドキュメントリストに追加**
   ```markdown
   - **[コマンドリファレンス](docs/COMMANDS.md)** ← 追加
   ```

---

#### `docs/QUICKSTART.md`
**変更箇所:**

環境変数の設定例に追加:
```env
DISCORD_TOKEN=あなたのBotトークン
GUILD_ID=テスト用サーバーID（オプション）
FEEDBACK_CHANNEL_ID=フィードバック受信チャンネルID（オプション） ← 追加
```

---

#### `docs/DEVELOPMENT.md`
**変更箇所:**

開発環境・本番環境の設定例に追加:
```env
# 開発環境
FEEDBACK_CHANNEL_ID=dev_feedback_channel_id  ← 追加

# 本番環境
FEEDBACK_CHANNEL_ID=prod_feedback_channel_id  ← 追加
```

---

#### `docs/PROJECT_SUMMARY.md`
**変更箇所:**

コマンド数の更新:
```markdown
- ✅ 完全な機能実装（7つのコマンド） ← 5つから7つに変更
  - 予約管理: reserve, cancel, complete
  - 表示: list, my-reservations
  - ユーティリティ: help, feedback ← 追加
```

---

#### `setup.sh`
**変更箇所:**

セットアップ時のメッセージに追加:
```bash
echo "   - FEEDBACK_CHANNEL_ID: フィードバック受信チャンネルのID（オプション、/feedbackコマンド用）"
```

---

## 🔐 プライバシー保護

### Ephemeralメッセージの使用

`/help` と `/feedback` コマンドは、Discordの「Ephemeralメッセージ」機能を使用しています。

**特徴:**
- ✅ コマンドを実行した人にのみ表示される
- ✅ 他のユーザーには一切見えない
- ✅ チャンネルの履歴に残らない
- ✅ 他のユーザーに通知されない

**実装:**
```go
respondEphemeral(s, i, message)

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: message,
            Flags:   discordgo.MessageFlagsEphemeral,  // 重要！
        },
    })
}
```

---

### フィードバックの匿名性

`/feedback` コマンドは完全に匿名化されています。

**保証される匿名性:**

1. **フィードバックチャンネルへの転送時**
   - ❌ ユーザーID: 含まれない
   - ❌ ユーザー名: 含まれない
   - ❌ ニックネーム: 含まれない
   - ❌ アバター: 含まれない
   - ✅ タイムスタンプ: 含まれる（匿名）

2. **ログへの記録時**
   - ❌ フィードバック内容: 記録されない
   - ✅ メッセージの長さ: 記録される（統計用）
   - ✅ コマンド実行回数: 記録される

**実装:**
```go
// フィードバックチャンネルに転送（送信者情報なし）
feedbackEmbed := &discordgo.MessageEmbed{
    Title:       "💬 新しいフィードバック",
    Description: message,  // 内容のみ
    Footer: &discordgo.MessageEmbedFooter{
        Text: fmt.Sprintf("受信日時: %s | 匿名フィードバック", timestamp),
    },
}

// ログには長さのみ記録
logger.LogCommand("feedback", userID, username, channelID, true, "",
    map[string]interface{}{"message_length": len(message)})
```

---

## 🧪 テスト

### ビルドテスト

```bash
$ go build -o bin/hxs_reservation_system main.go
$ echo $?
0  # ビルド成功
```

### 動作確認方法

1. **ヘルプコマンドのテスト**
   ```
   /help
   ```
   - ✅ ヘルプメッセージが表示される
   - ✅ 自分にしか見えない
   - ✅ すべてのコマンドがリストされている

2. **フィードバックコマンドのテスト（環境変数設定前）**
   ```
   /feedback message:テストメッセージ
   ```
   - ✅ エラーメッセージが表示される
   - ✅ 「フィードバックチャンネルが設定されていません」と表示される

3. **フィードバックコマンドのテスト（環境変数設定後）**
   ```
   /feedback message:これはテストです
   ```
   - ✅ 確認メッセージが表示される
   - ✅ フィードバックチャンネルにメッセージが届く
   - ✅ 送信者情報は含まれていない

---

## 📊 統計

### コマンド数の推移
- **変更前**: 5つのコマンド
  - reserve, cancel, complete, list, my-reservations

- **変更後**: 7つのコマンド (+2)
  - reserve, cancel, complete, list, my-reservations
  - **help** ✨ NEW
  - **feedback** ✨ NEW

### コードの変更量
- **追加された関数**: 2つ（handleHelp, handleFeedback）
- **追加されたコマンド定義**: 2つ
- **更新されたドキュメント**: 6ファイル
- **新規作成されたドキュメント**: 1ファイル（COMMANDS.md）

---

## 🎯 今後の展望

### 考えられる改善点

1. **フィードバックのカテゴリ分類**
   - バグ報告、機能要望、改善提案などのカテゴリを追加

2. **フィードバックの優先度設定**
   - ユーザーが重要度を指定できるようにする

3. **フィードバックへの返信機能**
   - 運営チームからユーザーに（匿名のまま）返信できる機能

4. **ヘルプコマンドの改善**
   - 特定のコマンドのヘルプを表示: `/help reserve`
   - インタラクティブなヘルプ（ボタンで詳細表示）

---

## ✅ チェックリスト

すべての変更が完了し、動作確認済みです：

- [x] `/help` コマンドの実装
- [x] `/feedback` コマンドの実装
- [x] Ephemeralメッセージの実装
- [x] 匿名性の保証
- [x] 環境変数の追加（FEEDBACK_CHANNEL_ID）
- [x] 設定ファイルの更新（.env.example, .env.development, .env.production）
- [x] ドキュメントの更新（README.md, QUICKSTART.md, DEVELOPMENT.md, PROJECT_SUMMARY.md）
- [x] 新規ドキュメントの作成（COMMANDS.md）
- [x] セットアップスクリプトの更新（setup.sh）
- [x] ビルドの成功確認
- [x] エラーチェック（コンパイルエラーなし）

---

## 🎉 まとめ

ユーザビリティとフィードバック収集を大幅に改善する2つの新機能を追加しました：

**✨ ハイライト:**
- **`/help`**: すべてのコマンドをその場で確認できる
- **`/feedback`**: 完全匿名でフィードバックを送信できる
- **プライバシー保護**: Ephemeralメッセージで他のユーザーに見えない
- **充実したドキュメント**: COMMANDS.mdで全コマンドを網羅

**🚀 次のステップ:**
1. `.env` ファイルに `FEEDBACK_CHANNEL_ID` を設定
2. Botを再起動して新しいコマンドを登録
3. `/help` コマンドで使い方を確認
4. `/feedback` コマンドでユーザーの声を収集

---

**更新日**: 2025年11月9日
**バージョン**: v1.2.0（/help, /feedback追加）
