package hub

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

type MsgType string

const (
	MsgTypeCommon = "common"
	MsgTypeDirect = "direct"
)

type Msg struct {
	Body      string
	ChannelID string
	UserID    string
	Type      MsgType
}

type Interaction struct {
	I *discordgo.Interaction
	R *discordgo.InteractionResponse
}

type Bot struct {
	ctx             context.Context
	logger          *log.Logger
	errs            chan error
	cfg             *Config
	msgChan         chan *Msg
	interactionChan chan *Interaction
	store           Store
	commands        map[string]Cmd
}

func NewBot(ctx context.Context, logger *log.Logger, cfg *Config, state Store, cmds ...Cmd) *Bot {
	cmdMap := map[string]Cmd{}
	for _, cmd := range cmds {
		cmdMap[cmd.Meta().Name] = cmd
	}
	return &Bot{
		ctx:             ctx,
		cfg:             cfg,
		errs:            make(chan error, 1),
		logger:          logger,
		msgChan:         make(chan *Msg, 100),
		interactionChan: make(chan *Interaction, 100),
		store:           state,
		commands:        cmdMap,
	}
}

func (b *Bot) Start() error {
	dg, err := discordgo.New("Bot " + b.cfg.Token())
	if err != nil {
		return fmt.Errorf("fail creating Discord session %v", err)
	}
	defer func() {
		b.logger.Println("closing bot")
		err := dg.Close()
		if err != nil {
			b.logger.Printf("close err %v", err)
		}
	}()
	cmds := make([]*discordgo.ApplicationCommand, 0)
	for _, cmd := range b.commands {
		cmds = append(cmds, cmd.Meta())
	}
	created, err := dg.ApplicationCommandBulkOverwrite(b.cfg.AppID(), b.cfg.GuildID(), cmds)
	if err != nil {
		return err
	}
	for _, command := range created {
		b.logger.Printf("cmd '%s' created", command.Name)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		b.logger.Printf("Bot logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		cmdData := i.ApplicationCommandData()
		cmdName := cmdData.Name
		if cmd, ok := b.commands[cmdName]; ok {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				options := make(map[string]interface{})
				for _, option := range cmdData.Options {
					options[option.Name] = option.Value
				}
				cmd.Handle(b, options, s, i)
			case discordgo.InteractionApplicationCommandAutocomplete:
				cmd.Autocomplete(b, s, i)
			}
		}
	})

	dg.AddHandler(b.handleCreateMessage)
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		return fmt.Errorf("fail opening connection %v", err)
	}
	b.logger.Println("Bot is now running.  Press CTRL-C to exit.")

	go func() {
		for {
			select {
			case <-b.ctx.Done():
				return
			case msg := <-b.msgChan:
				err = b.handleOutgoingMessages(dg, msg)
				if err != nil {
					b.errs <- err
					return
				}
			case inter := <-b.interactionChan:
				err = dg.InteractionRespond(inter.I, inter.R)
				if err != nil {
					b.errs <- err
					return
				}
			}
		}
	}()

	select {
	case err = <-b.errs:
		if err != nil {
			b.logger.Printf("bot err %v", err)
			return err
		}
	case <-b.ctx.Done():
	}
	return nil
}

func (b *Bot) handleCreateMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "ping" {
		b.SendMessage(fmt.Sprintf("pong <@%s>", m.Author.ID), m.ChannelID)
	}
}

func (b *Bot) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse) {
	b.interactionChan <- &Interaction{
		I: interaction,
		R: resp,
	}
}

func (b *Bot) SendMessage(body string, channelIDs ...string) {
	for _, channelID := range channelIDs {
		b.msgChan <- &Msg{
			ChannelID: channelID,
			Body:      body,
			Type:      MsgTypeCommon,
		}
	}
}

func (b *Bot) SendMessageToDirect(body string, userID string) {
	b.msgChan <- &Msg{
		UserID: userID,
		Body:   body,
		Type:   MsgTypeDirect,
	}
}

func (b *Bot) handleOutgoingMessages(s *discordgo.Session, msg *Msg) error {
	if msg.Type == MsgTypeDirect {
		channel, err := s.UserChannelCreate(msg.UserID)
		if err != nil {
			return err
		}
		msg.ChannelID = channel.ID
	}
	b.logger.Printf("send (%s) to %v", msg.Body, msg.ChannelID)
	_, err := s.ChannelMessageSend(msg.ChannelID, msg.Body)
	if err != nil {
		return err
	}
	return nil
}
