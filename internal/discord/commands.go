package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type DiscordCommand struct {
	log         *zap.Logger
	commandList []*discordgo.ApplicationCommand
	callbackMap map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

const (
	helloWorldCommandName = "test-command"
)

var (
	helloWorldCommand = discordgo.ApplicationCommand{
		Name:        helloWorldCommandName,
		Description: "Shagohad says hello",
	}

	commandList = []*discordgo.ApplicationCommand{
		&helloWorldCommand,
	}
)

func NewDiscordCommand(log *zap.Logger) *DiscordCommand {
	dc := DiscordCommand{
		log: log,
		commandList: []*discordgo.ApplicationCommand{
			&helloWorldCommand,
		},
		callbackMap: make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)),
	}

	dc.populateCallbackMap()

	return &dc
}

func (d *DiscordCommand) populateCallbackMap() {
	d.callbackMap[helloWorldCommandName] = d.helloWorldCallback
}
func (d *DiscordCommand) Log() *zap.Logger {
	return d.log
}

func (d *DiscordCommand) generalCommandCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if callback, ok := d.callbackMap[i.ApplicationCommandData().Name]; ok {
		callback(s, i)
	}
}

func (d *DiscordCommand) helloWorldCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	d.Log().Info("In hello world Callback")
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Now you owe me 2000$ß",
		},
	})
}

func (d *DiscordCommand) AddGeneralHandler(s *discordgo.Session) {
	s.AddHandler(d.generalCommandCallback)
}

func (d *DiscordCommand) CommandList() []*discordgo.ApplicationCommand {
	return d.commandList
}
