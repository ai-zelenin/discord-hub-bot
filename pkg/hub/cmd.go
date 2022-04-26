package hub

import (
	"github.com/bwmarrin/discordgo"
)

type Cmd interface {
	Meta() *discordgo.ApplicationCommand
	Handle(b *Bot, options map[string]interface{}, s *discordgo.Session, i *discordgo.InteractionCreate)
	Autocomplete(b *Bot, s *discordgo.Session, i *discordgo.InteractionCreate)
}
