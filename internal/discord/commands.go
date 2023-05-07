package discord

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type DiscordCommand struct {
	log         *zap.Logger
	commandList []*discordgo.ApplicationCommand
	callbackMap map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

	// VoiceManager
	vm *VoiceManager
	// Command specific data
	legoYodaBuffer [][]byte
}

const (
	helloWorldCommandName   = "test-command"
	playLegoYodaCommandName = "play-yoda"
)

var (
	helloWorldCommand = discordgo.ApplicationCommand{
		Name:        helloWorldCommandName,
		Description: "Shagohad says hello",
	}

	playLegoYodaCommand = discordgo.ApplicationCommand{
		Name:        playLegoYodaCommandName,
		Description: "Play a fun noise from a cute critter",
	}

	commandList = []*discordgo.ApplicationCommand{
		&helloWorldCommand,
		&playLegoYodaCommand,
	}

	LegoYodaMp3Path = "files/legoYoda.mp3"
)

func NewDiscordCommand(log *zap.Logger) *DiscordCommand {
	dc := DiscordCommand{
		log: log,
		commandList: []*discordgo.ApplicationCommand{
			&helloWorldCommand,
			&playLegoYodaCommand,
		},
		callbackMap: make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)),
	}

	dc.populateCallbackMap()

	dc.vm = NewVoiceManager(log)

	// Prepare legoYoda ahead of time
	dc.vm.ConvertMp3ToBuffer(LegoYodaMp3Path, &dc.legoYodaBuffer)

	return &dc
}

func (d *DiscordCommand) populateCallbackMap() {
	d.callbackMap[helloWorldCommandName] = d.helloWorldCallback
	d.callbackMap[playLegoYodaCommandName] = d.playLegoYodaCallback
}
func (d *DiscordCommand) Log() *zap.Logger {
	return d.log
}

func (d *DiscordCommand) generalCommandCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if callback, ok := d.callbackMap[i.ApplicationCommandData().Name]; ok {
		callback(s, i)
	}
}

func (d *DiscordCommand) AddGeneralHandler(s *discordgo.Session) {
	s.AddHandler(d.generalCommandCallback)
}

func (d *DiscordCommand) CommandList() []*discordgo.ApplicationCommand {
	return d.commandList
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

func (d *DiscordCommand) handlePlayingAudio(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	var id string
	if i.Interaction.Member != nil {
		id = i.Interaction.Member.User.ID
	} else {
		// TODO: Throw a fit if user tries to use this from DM
		id = i.Interaction.User.ID
	}
	if id == s.State.User.ID {
		return errors.New("recursion prevention")
	}

	// Find the channel that the message came from.
	c, err := s.State.Channel(i.Interaction.ChannelID)
	if err != nil {
		// Could not find channel.
		d.Log().Error("Could not find Channel")
		return errors.New("resource not found")
	}

	// Find the guild for that channel.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find guild.
		d.Log().Error("Could not find guild")
		return errors.New("resource not found")
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range g.VoiceStates {
		if vs.UserID == id {

			err = d.vm.PlayAudio(s, g.ID, vs.ChannelID, &d.legoYodaBuffer)
			if err != nil {
				fmt.Println("Error playing sound:", err)
			}

			return nil
		}
	}

	return errors.New("resource not found")
}

func (d *DiscordCommand) playLegoYodaCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	d.Log().Info("In playLegoYoda callback")
	var content string
	err := d.handlePlayingAudio(s, i)
	if err != nil {
		content = "Error playing Lego Yoda!"
	} else {
		content = "Crush my cock with a rock I must"
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}
