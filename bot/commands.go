package bot

import (
	"github.com/bwmarrin/discordgo"
)

var integrationTypes = []discordgo.ApplicationIntegrationType{
	discordgo.ApplicationIntegrationGuildInstall,
	discordgo.ApplicationIntegrationUserInstall,
}

var contexts = []discordgo.InteractionContextType{
	discordgo.InteractionContextGuild,
	discordgo.InteractionContextBotDM,
	discordgo.InteractionContextPrivateChannel,
}

var defaultMemberPermissions int64 = discordgo.PermissionAdministrator
var dmPermission = false

type command struct {
	data    *discordgo.ApplicationCommand
	execute func(*discordgo.Session, *discordgo.InteractionCreate, *Bot)
}

var commands = map[string]command{
	"chat":            chat,
	"clear":           clear,
	"set-bot-channel": setBotChannel,
	"ping":            ping,
	"history":         history,
}
