package bot

import (
	"github.com/bwmarrin/discordgo"
)

type command struct {
	data    *discordgo.ApplicationCommand
	execute func(*discordgo.Session, *discordgo.InteractionCreate, *Bot)
}

var commands = map[string]command{
	"clear":           clear,
	"set-bot-channel": setBotChannel,
	"ping":            ping,
}
