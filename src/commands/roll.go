package commands

import (
	"fmt"
	"main/gcp/firestore/guild"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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
				// Note: this isn't documented, but you can use that if you want to.
				// This flag just allows you to create messages visible only for the caller of the command
				// (user who triggered the command)
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
		var resultMsg string
		if len(winners) > 0 {
			resultMsg = fmt.Sprintf("🎉 The number is **%d**! Congratulations to: %s", randomNumber, strings.Join(winners, ", "))
		} else {
			resultMsg = fmt.Sprintf("The number is **%d**! Better luck next time!", randomNumber)
		}

		// 4. Edit the original message
		s.FollowupMessageEdit(i.Interaction, resp.ID, &discordgo.WebhookEdit{
			Content: &resultMsg,
		})

		guildData.AddRolledNumber(ctx, randomNumber)
	},
	lastRollCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	},
	allRolledCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	},
	scheduleRollCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	},
}
