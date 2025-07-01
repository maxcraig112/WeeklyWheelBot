package main

import (
	"context"
	"log"
	"main/commands"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var s *discordgo.Session

var ctx = context.Background()
var GuildID string
var BotToken string
var RemoveCommands bool
var err error

func init() {
	_ = godotenv.Load()

	GuildID = os.Getenv("GUILDID")
	BotToken = os.Getenv("BOTTOKEN")
	RemoveCommands = strings.ToLower(os.Getenv("REMOVECOMMANDS")) == "true"

	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (

// {
// 	Name:        "Bulk Register",
// 	Description: "Provide a list of users and guesses and bulk register them to the system",
// 	Options: []*discordgo.ApplicationCommandOption{
// 		{
// 			Type:        discordgo.ApplicationCommandOptionAttachment,
// 			Name:        ".txt file",
// 			Description: "txt file with each line 'DISCORDID GUESS' ",
// 			Required:    true,
// 		},
// 	},
// },
// {
// 	Name:        "Upload Past Rolls",
// 	Description: "Provide a list of all numbers which have been previously rolled so they are uploaded",
// 	Options: []*discordgo.ApplicationCommandOption{
// 		{
// 			Type:        discordgo.ApplicationCommandOptionAttachment,
// 			Name:        ".txt file",
// 			Description: "txt file with each line being a pass roll",
// 			Required:    true,
// 		},
// 	},
// },

)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		if h, ok := commands.CommandHandler[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionMessageComponent {
			return
		}
		action := strings.Split(i.MessageComponentData().CustomID, "|")[0]
		if h, ok := commands.MessageHandler[action]; ok {
			h(s, i)
		}

	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands.Commands))
	for i, v := range commands.Commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, &v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// this will make sure to close the clients when the application ends

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if RemoveCommands {
		log.Println("Removing commands...")
		// // We need to fetch the commands, since deleting requires the command ID.
		// // We are doing this from the returned commands on line 375, because using
		// // this will delete all the commands, which might not be desirable, so we
		// // are deleting only the commands that we added.
		// registeredCommands, err := s.ApplicationCommands(s.State.User.ID, *GuildID)
		// if err != nil {
		// 	log.Fatalf("Could not fetch registered commands: %v", err)
		// }

		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
