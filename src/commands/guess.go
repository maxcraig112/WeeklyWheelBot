package commands

import (
	"fmt"
	"main/gcp/firestore/guild"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gotidy/ptr"
	"github.com/samber/lo"
)

var currentGuessCommand = "currentguess"
var guessCommandName = "guess"
var GuessCommands = []discordgo.ApplicationCommand{
	{
		Name:        currentGuessCommand,
		Description: "What is the current guess of the user in this server (defaults to user making command)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "guesser",
				Description: "Who is making the guess (should be a discord mention), by default it is the user making the command",
				Required:    false,
			},
		},
	},
	{
		Name:        guessCommandName,
		Description: "Register your guess between 1 and 1000",
		Options: []*discordgo.ApplicationCommandOption{

			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "guess",
				Description: "Make a guess between 1 and 1000",
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
	},
}

var GuessCommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	currentGuessCommand: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guildID := i.GuildID
		guildStore := guild.NewGuildStore(Clients.Firestore)
		guildData, err := guildStore.CreateOrGetGuildDocument(ctx, guildID)

		options := i.ApplicationCommandData().Options

		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		var userID string

		// if the user provided a guesses as a mention then we will use that
		if option, ok := optionMap["guesser"]; ok {
			// validate that it is of the form <@userID>
			mention := option.StringValue()
			if strings.HasPrefix(mention, "<@") && strings.HasSuffix(mention, ">") {
				userID = mention
			} else {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid guesser value, it has to be a user mention",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
		} else {
			userID = string(getMentionFromUserID(i.Member.User.ID))
		}
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Something went wrong! %s", err.Error()),
				},
			})
			return
		}
		if currentGuess, ok := guildData.Guesses[userID]; ok {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("The current guess for %s is %d!", userID, currentGuess),
				},
			})
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%s hasn't guessed a number yet in this server, run /guess to make one", userID),
				},
			})
		}
	},
	guessCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guildID := i.GuildID
		guildStore := guild.NewGuildStore(Clients.Firestore)
		guildData, err := guildStore.CreateOrGetGuildDocument(ctx, guildID)
		// Access options in the order provided by the user.
		options := i.ApplicationCommandData().Options

		// Or convert the slice into a map
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		var userID string
		var guess int

		// if the user provided a guesses as a mention then we will use that
		if option, ok := optionMap["guesser"]; ok {
			// validate that it is of the form <@userID>
			mention := option.StringValue()
			if strings.HasPrefix(mention, "<@") && strings.HasSuffix(mention, ">") {
				userID = mention
			} else {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid guesser value, it has to be a user mention",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
		} else {
			userID = string(getMentionFromUserID(i.Member.User.ID))
		}

		if opt, ok := optionMap["guess"]; ok {
			guess = int(opt.IntValue())
			// validate that no once else has the same guess
			invertedGuesses := invertGuessMap(guildData.Guesses)
			if user, ok := invertedGuesses[guess]; ok {
				if user == userID {
					// you have already guessed this number
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("%s have already guessed this number!", userID),
						},
					})
					return
				}
				// someone has already guessed this number
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Sorry! %s cannot guess %d as it has already been guessed by %s", userID, guess, user),
					},
				})
				return
			}
			// validate that the guess hasn't already been rolled
			previousRolls := lo.Map(guildData.SpunNumbers, func(spunNumber guild.SpunNumber, _ int) int { return spunNumber.Number })
			if lo.Contains(previousRolls, guess) {
				// cannot guess this number as it has been previously rolled
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Sorry! %s cannot guess %d as it has already been rolled in this server", userID, guess),
					},
				})
				return
			}
		}

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
					Content: fmt.Sprintf("%s already have a guess (%v). Do you want to override it?", userID, previousGuess),
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
					Flags: discordgo.MessageFlagsEphemeral,
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
					"%s now has a guess of %d", userID, guess,
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
				Content: fmt.Sprintf("The guess for %s has been overridden to %d.", userID, guess),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
	"guess_override_no": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Your guess was not changed.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}

func getMentionFromUserID(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

func invertGuessMap(guesses map[string]int) map[int]string {
	inverted := make(map[int]string, len(guesses))
	for k, v := range guesses {
		inverted[v] = k
	}

	return inverted
}
