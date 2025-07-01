package commands

import (
	"context"
	"log"
	"main/gcp"
	"maps"

	"github.com/bwmarrin/discordgo"
)

var Clients *gcp.Clients
var ctx = context.Background()
var err error

var Commands []discordgo.ApplicationCommand
var CommandHandler map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
var MessageHandler map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

func init() {
	Clients, err = gcp.InitialiseClients(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize clients: %v", err)
	}

	Commands = []discordgo.ApplicationCommand{}
	Commands = append(Commands, GuessCommands...)
	Commands = append(Commands, SpinCommands...)
	Commands = append(Commands, UploadCommands...)

	CommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
	maps.Copy(CommandHandler, GuessCommandHandler)
	maps.Copy(CommandHandler, SpinCommandHandler)
	maps.Copy(CommandHandler, UploadCommandHandler)

	MessageHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
	maps.Copy(MessageHandler, GuessMessageHandler)
}
