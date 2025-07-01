package commands

import (
	"fmt"
	"main/gcp/firestore/guild"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gotidy/ptr"
)

var GuessCommandName = "guess"
var GuessCommand = discordgo.ApplicationCommand{
	Name:        GuessCommandName,
	Description: "Register your guess between 1 and 1000",
	Options: []*discordgo.ApplicationCommandOption{

		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "guess",
			Description: "Guess between 1 and 1000",
			MinValue:    ptr.Of(float64(1)),
			MaxValue:    1000,
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "guesser",
			Description: "Who is making the guess (should be a discord mention), by default it is the user making the command",
			Required:    false,
		},
	},
}

var GuessCommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	GuessCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guildID := i.GuildID
		// Access options in the order provided by the user.
		options := i.ApplicationCommandData().Options

		// Or convert the slice into a map
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		var userID string
		var guess int
		// This example stores the provided arguments in an []interface{}
		// which will be used to format the bots response
		margs := make([]interface{}, 0, len(options))
		msgformat := "You just made a guess, the values you entered are "

		// if the user provided a guesses as a mention then we will use that
		if option, ok := optionMap["guesser"]; ok {
			userID = string(option.StringValue())
		} else {
			userID = string(fmt.Sprintf("<@%s>", i.Member.User.ID))
		}
		msgformat += fmt.Sprintf("\nGuesserID: %s", userID)

		if opt, ok := optionMap["guess"]; ok {
			guess = int(opt.IntValue())
			msgformat += fmt.Sprintf("\nGuess: %d", guess)
		}

		guildStore := guild.NewGuildStore(Clients.Firestore)
		guildData, err := guildStore.CreateOrGetGuildDocument(ctx, guildID)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Something went wrong! %s", err.Error()),
				},
			})
			return
		}
		if previousGuess, ok := guildData.Guesses[string(userID)]; ok {
			// Send a message with yes/no buttons asking if they want to override their guess
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("You already have a guess (%v). Do you want to override it?", previousGuess),
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Yes",
									Style:    discordgo.PrimaryButton,
									CustomID: fmt.Sprintf("guess_override_yes|%s|%d|%s", guildID, guess, userID),
								},
								discordgo.Button{
									Label:    "No",
									Style:    discordgo.SecondaryButton,
									CustomID: fmt.Sprintf("guess_override_no|%s|%d|%s", guildID, guess, userID),
								},
							},
						},
					},
				},
			})
			return
		}

		err = guildData.SetGuess(ctx, userID, guess)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Something went wrong! %s", err.Error()),
				},
			})
			return
		}

		// we have to add this to the guild guess
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			// Ignore type for now, they will be discussed in "responses"
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf(
					msgformat,
					margs...,
				),
			},
		})
	},
}

var GuessMessageHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"guess_override_yes": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		parts := strings.Split(i.MessageComponentData().CustomID, "|")
		if len(parts) < 4 {
			// handle error
			return
		}
		guildID := parts[1]
		g, _ := strconv.Atoi(parts[2])
		guess := int(g)
		userID := string(parts[3])
		// Update the guess in Firestore for userID in guildID
		// Respond to the user
		guildStore := guild.NewGuildStore(Clients.Firestore)
		guildData, err := guildStore.CreateOrGetGuildDocument(ctx, guildID)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Something went wrong! %s", err.Error()),
				},
			})
			return
		}

		err = guildData.SetGuess(ctx, userID, guess)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Something went wrong! %s", err.Error()),
				},
			})
			return
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Guess for %s has been overridden to %d.", userID, guess),
			},
		})
	},
	"guess_override_no": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Your guess was not changed.",
			},
		})
	},
}
