package commands

import (
	"bufio"
	"fmt"
	"main/gcp/firestore/guild"
	"net/http"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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

var uploadPastRollsName = "uploadrolls"

var UploadCommands = []discordgo.ApplicationCommand{
	{
		Name:        uploadPastRollsName,
		Description: "Upload a txt file of all previous rolls (overrides previous rolls)",
		Options: []*discordgo.ApplicationCommandOption{

			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "file",
				Description: "txt file with a roll number on each line",
				Required:    true,
			},
		},
	},
}

var UploadCommandHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	uploadPastRollsName: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guildID := i.GuildID

		guildStore := guild.NewGuildStore(Clients.Firestore)
		guildData, err := guildStore.CreateOrGetGuildDocument(ctx, guildID)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Something went wrong! %s", err.Error()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		options := i.ApplicationCommandData().Options
		if len(options) == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "No file provided.",
				},
			})
			return
		}

		attachmentID := options[0].Value.(string)
		attachment, ok := i.ApplicationCommandData().Resolved.Attachments[attachmentID]
		if !ok {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Could not resolve attachment.",
				},
			})
			return
		}
		attachmentUrl := attachment.URL

		// Check if file is .txt (ignore query parameters)
		baseUrl := attachmentUrl
		if idx := strings.Index(baseUrl, "?"); idx != -1 {
			baseUrl = baseUrl[:idx]
		}
		if len(baseUrl) < 4 || baseUrl[len(baseUrl)-4:] != ".txt" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "File must be a .txt file.",
				},
			})
			return
		}

		// Download the file
		resp, err := http.Get(attachmentUrl)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Could not download the file.",
				},
			})
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var rolls []int
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if n, err := strconv.Atoi(line); err == nil {
				rolls = append(rolls, n)
			}
		}

		guildData.BulkAddRolledNumbers(ctx, rolls)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully uploaded %d rolls from file.", len(rolls)),
			},
		})
	},
}
