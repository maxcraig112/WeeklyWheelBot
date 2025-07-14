package commands

import (
	"fmt"
	"main/gcp/firestore/guild"
	"main/gif"
	"math/rand"
	"strings"
	"time"

	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/gotidy/ptr"
	"github.com/samber/lo"
)

var spinCommandName = "spin"
var lastSpinCommandName = "lastspin"
var allSpunCommandName = "allspins"
var scheduleSpinCommandName = "scheduledspin"

var SpinCommands = []discordgo.ApplicationCommand{
	{
		Name:        spinCommandName,
		Description: "Manually spin a number between 1 and 1000",
	},
	{
		Name:        lastSpinCommandName,
		Description: "Get the last number that was spun in this server",
	},
	{
		Name:        allSpunCommandName,
		Description: "Get a list of all the numbers that have been spun in this server",
	},
	{
		Name:        scheduleSpinCommandName,
		Description: "Schedule what date and time the number should be spun (frequency once a week)",
		Options: []*discordgo.ApplicationCommandOption{

			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "day",
				Description: "What Day the spin will start on",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "What Time the spin will start on (24HH Time)",
				Required:    true,
			},
		},
	},
}

var SpinCommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	spinCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Now Spining!",
			},
		})
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

		previouslySpunNumbers := lo.Map(guildData.SpunNumbers, func(data guild.SpunNumber, _ int) int {
			return data.Number
		})

		allGuesses := make(map[int][]string)
		for userID, guess := range guildData.Guesses {
			allGuesses[guess] = append(allGuesses[guess], userID)
		}

		// Build a set for fast exclusion
		excluded := make(map[int]struct{}, len(previouslySpunNumbers))
		for _, n := range previouslySpunNumbers {
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

		// Generate the spinning wheel GIF with the spined number
		gifFile, err := gif.CreateSpinningWheelGIF("", 12, fmt.Sprintf("%d", randomNumber))
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Something went wrong generating the GIF! %s", err.Error()),
				},
			})
			return
		}
		defer func() {
			gifFile.Close()
		}()

		// 1. Send initial drumroll message with GIF embed
		gifFileName := "spinning_wheel.gif"
		resp, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Now spining a number, drumroll please! 🥁🥁🥁",
			Files: []*discordgo.File{{
				Name:        gifFileName,
				ContentType: "image/gif",
				Reader:      gifFile,
			}},
			Embeds: []*discordgo.MessageEmbed{{
				Image: &discordgo.MessageEmbedImage{
					URL: "attachment://" + gifFileName,
				},
			}},
		})
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Something went wrong %s", err.Error()),
			})
			return
		}

		// 2. Wait a couple of seconds
		time.Sleep(7 * time.Second)

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

			s.FollowupMessageEdit(i.Interaction, resp.ID, &discordgo.WebhookEdit{
				Content: ptr.Of(fmt.Sprintf("The number is **%d**! Better luck next time!", randomNumber)),
			})
		}

		// 4. Edit the original message

		guildData.AddSpunNumber(ctx, randomNumber)
	},
	lastSpinCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

		lastSpin := guildData.LastNumberSpun
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("The last number spined was %d on %s", lastSpin.Number, lastSpin.DateSpun.Format("02-01-2006")),
			},
		})

	},
	allSpunCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

		numbers := lo.Map(guildData.SpunNumbers, func(data guild.SpunNumber, _ int) int { return data.Number })
		sort.Ints(numbers)
		numberStrings := lo.Map(numbers, func(n int, _ int) string { return fmt.Sprintf("%d", n) })

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Here are a list of all the spins that have been made in this server!",
				Files: []*discordgo.File{
					{
						ContentType: "text/plain",
						Name:        "spins.txt",
						Reader:      strings.NewReader(strings.Join(numberStrings, "\n")),
					},
				},
			},
		})
	},
	scheduleSpinCommandName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Not yet implemented :(",
			},
		})
	},
}
