package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/wamuusubi/shagohad/internal/discord"
	"go.uber.org/zap"
)

var (
	botToken = flag.String("token", "", "Bot access token")
)

type discordBot struct {
	discordCommand     *discord.DiscordCommand
	session            *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
	l                  *zap.Logger
	stop               chan os.Signal
}

func (d *discordBot) Session() *discordgo.Session {
	return d.session
}

func (d *discordBot) Log() *zap.Logger {
	return d.l
}

func (d *discordBot) Stop() chan os.Signal {
	return d.stop
}

func init() { flag.Parse() }

func NewDiscordBot(botToken *string, l *zap.Logger) (*discordBot, error) {

	bot := discordBot{}

	guildId := ""

	// Set up Session
	session, err := discordgo.New("Bot " + *botToken)

	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
		return nil, err
	}

	session.LogLevel = discordgo.LogInformational
	bot.l = l
	bot.session = session
	bot.discordCommand = discord.NewDiscordCommand(l)
	bot.registeredCommands = make([]*discordgo.ApplicationCommand, len(bot.discordCommand.CommandList()))

	err = session.Open()
	if err != nil {
		bot.l.Error("Cannot open the session", zap.Error(err))
		return nil, err
	}
	// Set up commands
	for i, command := range bot.discordCommand.CommandList() {
		cmd, err := bot.Session().ApplicationCommandCreate(bot.Session().State.User.ID, guildId, command)
		if err != nil {
			bot.Log().Error("Failed to create command", zap.String("commandName", command.Name))
			return nil, err
		}
		bot.registeredCommands[i] = cmd
	}

	bot.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Set up callbacks
	bot.discordCommand.AddGeneralHandler(bot.session)

	bot.stop = make(chan os.Signal, 1)

	return &bot, nil
}

func loggerHook(msgL, caller int, format string, a ...interface{}) {

}

// This is the main discord bot file, will likely be split up as time goes ons
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// TODO: Add bot token enforcement
	if botToken == nil || *botToken == "" {
		panic("Bot token required")
	}

	logger.Info("", zap.String("botToken", *botToken))

	bot, err := NewDiscordBot(botToken, logger)

	if err != nil {
		panic(err)
	}

	defer bot.Session().Close()

	signal.Notify(bot.Stop(), os.Interrupt)
	fmt.Printf("Press Ctrl+C to exit")
	<-bot.Stop()

	fmt.Printf("Shutting down")
}
