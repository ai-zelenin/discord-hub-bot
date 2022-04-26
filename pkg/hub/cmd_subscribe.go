package hub

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v3"
)

type SubscribeCmd struct {
	store Store
	cfg   *Config
}

func NewSubscribeCmd(cfg *Config, store Store) *SubscribeCmd {
	return &SubscribeCmd{
		store: store,
		cfg:   cfg,
	}
}

func (c *SubscribeCmd) Meta() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "subscribe",
		Description: "subscribe for one of integration channel",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "source",
				Description:  "source name",
				Required:     true,
				Autocomplete: true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "notify-in-direct",
				Description: "if True bot send message directly to user",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "only-personal",
				Description: "if True bot send message only if notification match your(for example mentions you as assignee)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "custom-filter",
				Description: `example: MatchTokens(JsonPath("issue.fields.assignee.displayName"),"Name Surname")`,
				Required:    false,
			},
		},
	}
}

func (c *SubscribeCmd) Handle(b *Bot, options map[string]interface{}, _ *discordgo.Session, i *discordgo.InteractionCreate) {
	var user *discordgo.User
	switch {
	case i.Member != nil && i.Member.User != nil:
		user = i.Member.User
	case i.User != nil:
		user = i.User
	}
	if user == nil {
		b.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "cannot identify user",
			},
		})
	}
	var customFilter string
	v, ok := options["custom-filter"]
	if ok {
		customFilter = v.(string)
	}
	sub := &Subscription{
		ChannelID:    i.ChannelID,
		Source:       options["source"].(string),
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Direct:       options["notify-in-direct"].(bool),
		OnlyPersonal: options["only-personal"].(bool),
		CustomFilter: customFilter,
	}
	err := c.store.AddSubscription(sub)
	if err != nil {
		b.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error %v", err),
			},
		})
	} else {
		data, err := yaml.Marshal(sub)
		if err != nil {
			b.errs <- err
			return
		}
		b.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Notification source added \n %s", string(data)),
			},
		})
	}
}

func (c *SubscribeCmd) Autocomplete(b *Bot, _ *discordgo.Session, i *discordgo.InteractionCreate) {
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
