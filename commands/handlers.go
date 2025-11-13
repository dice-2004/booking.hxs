package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/logging"
	"github.com/dice/hxs_reservation_system/storage"
)

var UpdateStatusCallback func()

func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string) {
	// コマンドインタラクションの処理
	commandName := i.ApplicationCommandData().Name
	isDM := i.GuildID == ""
	channelID := i.ChannelID

	var userID, username string
	if isDM {
		userID = i.User.ID
		username = i.User.Username
	} else {
		userID = i.Member.User.ID
		username = getDisplayName(i.Member)
	}

	if !isDM && allowedChannelID != "" && channelID != allowedChannelID {
		respondEphemeral(s, i, "This command can only be used in the allowed channel or DM.")
		logger.LogCommand(commandName, userID, username, channelID, false, "Not allowed channel", nil)
		return
	}

	parameters := make(map[string]interface{})
	for _, opt := range i.ApplicationCommandData().Options {
		parameters[opt.Name] = opt.Value
	}

	logger.LogCommand(commandName, userID, username, channelID, true, "", parameters)

	switch commandName {
	case "reserve":
		handleReserve(s, i, store, logger, allowedChannelID, isDM)
	case "cancel":
		handleCancel(s, i, store, logger, allowedChannelID, isDM)
	case "complete":
		handleComplete(s, i, store, logger, allowedChannelID, isDM)
	case "edit":
		handleEdit(s, i, store, logger, allowedChannelID, isDM)
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
