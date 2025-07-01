package commands

import (
	"fmt"
	"main/gcp/firestore/guild"
	"math/rand"
	"strings"
	"time"

	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/gotidy/ptr"
	"github.com/samber/lo"
)

var rollCommandName = "roll"
var lastRollCommandName = "lastroll"
var allRolledCommandName = "allrolled"
var scheduleRollCommandName = "scheduleroll"

var RollCommands = []discordgo.ApplicationCommand{
	{
		Name:        rollCommandName,
		Description: "Manually roll a number between 1 and 1000",
	},
	{
		Name:        lastRollCommandName,
		Description: "Get the last number that was rolled in this server",
	},
	{
		Name:        allRolledCommandName,
		Description: "Get a list of all numbers rolled in this server",
	},
	{
		Name:        scheduleRollCommandName,
		Description: "Schedule what date and time the number should be rolled (frequency once a week)",
		Options: []*discordgo.ApplicationCommandOption{

			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "day",
				Description: "What Day the roll will start on",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "What Time the roll will start on (24HH Time)",
				Required:    true,
			},
		},
	},
}

var loserGifs = []string{
	"https://i.imgur.com/ztfqRxX.gif",
}

var RollCommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	rollCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		//get guild data
		guildID := i.GuildID

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

		previouslyRolledNumbers := lo.Map(guildData.RolledNumbers, func(data guild.RolledNumber, _ int) int {
			return data.Number
		})

		allGuesses := make(map[int][]string)
		for userID, guess := range guildData.Guesses {
			allGuesses[guess] = append(allGuesses[guess], userID)
		}

		// Build a set for fast exclusion
		excluded := make(map[int]struct{}, len(previouslyRolledNumbers))
		for _, n := range previouslyRolledNumbers {
			excluded[n] = struct{}{}
		}

		// Build a slice of valid numbers
		validNumbers := make([]int, 0, 1000-len(excluded))
		for i := 1; i <= 1000; i++ {
			if _, found := excluded[i]; !found {
				validNumbers = append(validNumbers, i)
			}
		}
		randomNumber := validNumbers[rand.Intn(len(validNumbers))]

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Now Rolling!",
			},
		})
		// 1. Send initial drumroll message
		resp, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Now rolling a number, drumroll please! 🥁🥁🥁",
		})
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Something went wrong %s", err.Error()),
			})
			return
		}

		// 2. Wait a couple of seconds
		time.Sleep(5 * time.Second)

		// 3. Reveal the number and result
		winners := allGuesses[randomNumber]
		if len(winners) > 0 {
			embed := &discordgo.MessageEmbed{
				Image: &discordgo.MessageEmbedImage{
					URL: "https://i.imgur.com/XPoG75D.gif",
				},
			}
			s.FollowupMessageEdit(i.Interaction, resp.ID, &discordgo.WebhookEdit{
				Content: ptr.Of(fmt.Sprintf("🎉 The number is **%d**! HOLY SHIT YOU DID IT %s", randomNumber, strings.Join(winners, " and "))),
				Embeds:  &[]*discordgo.MessageEmbed{embed},
			})
		} else {
			gifUrl := loserGifs[rand.Intn(len(loserGifs))]
			embed := &discordgo.MessageEmbed{
				Image: &discordgo.MessageEmbedImage{
					URL: gifUrl,
				},
			}
			s.FollowupMessageEdit(i.Interaction, resp.ID, &discordgo.WebhookEdit{
				Content: ptr.Of(fmt.Sprintf("The number is **%d**! Better luck next time!", randomNumber)),
				Embeds:  &[]*discordgo.MessageEmbed{embed},
			})
		}

		// 4. Edit the original message

		guildData.AddRolledNumber(ctx, randomNumber)
	},
	lastRollCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		//get guild data
		guildID := i.GuildID

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

		lastRoll := guildData.LastNumberRolled
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf("The last number rolled was %d on %s", lastRoll.Number, lastRoll.DateRolled.Format("02-01-2006")),
			},
		})

	},
	allRolledCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		//get guild data
		guildID := i.GuildID

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

		numbers := lo.Map(guildData.RolledNumbers, func(data guild.RolledNumber, _ int) int { return data.Number })
		sort.Ints(numbers)
		numberStrings := lo.Map(numbers, func(n int, _ int) string { return fmt.Sprintf("%d", n) })

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Here are a list of all the rolls that have been made in this server!",
				Files: []*discordgo.File{
					{
						ContentType: "text/plain",
						Name:        "rolls.txt",
						Reader:      strings.NewReader(strings.Join(numberStrings, "\n")),
					},
				},
			},
		})
	},
	scheduleRollCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Not yet implemented :(",
			},
		})
	},
}
