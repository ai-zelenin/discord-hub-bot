package hub

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type UnsubscribeCmd struct {
	store Store
	cfg   *Config
}

func NewUnsubscribeCmd(cfg *Config, store Store) *UnsubscribeCmd {
	return &UnsubscribeCmd{
		store: store,
		cfg:   cfg,
	}
}

func (c *UnsubscribeCmd) Meta() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "unsubscribe",
		Description: "unsubscribe from integration channel",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "source",
				Description:  "source name",
				Required:     true,
				Autocomplete: true,
			},
		},
	}
}

func (c *UnsubscribeCmd) Handle(b *Bot, options map[string]interface{}, _ *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string
	switch {
	case i.Member != nil && i.Member.User != nil:
		userID = i.Member.User.ID
	case i.User != nil:
		userID = i.User.ID
	}
	if userID == "" {
		b.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "cannot identify user",
			},
		})
	}
	source := options["source"].(string)
	sub, err := c.store.RemoveSubscription(source, userID)
	if err != nil {
		b.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error %v", err),
			},
		})
	} else {
		b.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Notification source(%#v) removed", sub),
			},
		})
	}
}

func (c *UnsubscribeCmd) Autocomplete(b *Bot, _ *discordgo.Session, i *discordgo.InteractionCreate) {
	choises := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	for key, srcCfg := range c.cfg.Sources() {
		if srcCfg.ActionType == ActionTypeProxy {
			choises = append(choises, &discordgo.ApplicationCommandOptionChoice{
				Name:  key,
				Value: key,
			})
		}
	}
	b.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Content: "choose notification source",
			Choices: choises,
		},
	})
}
