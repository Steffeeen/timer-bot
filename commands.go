package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "until",
		Description: "Calculate the time until a date",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "date",
				Description: "The date to calculate the time until",
				Required:    true,
			},
		},
	},
	{
		Name:        "timer",
		Description: "Manage timers",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "create",
				Description: "Create a new timer",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "message",
						Description: "The message to display when the timer is up",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "time",
						Description: "The time to set the timer for",
						Required:    true,
					},
				},
			},
			{
				Name:        "list",
				Description: "List all timers",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "delete",
				Description: "Delete a timer",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "id",
						Description: "The ID of the timer to delete",
						Required:    true,
					},
				},
			},
			{
				Name:        "edit",
				Description: "Edit a timer",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "id",
						Description: "The ID of the timer to edit",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "message",
						Description: "The new message for the timer",
						Required:    false,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "time",
						Description: "The new time for the timer",
						Required:    false,
					},
				},
			},
			{
				Name:        "snooze",
				Description: "Snooze a timer",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "id",
						Description: "The ID of the timer to snooze",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "time",
						Description: "The amount of time to snooze the timer for",
						Required:    true,
					},
				},
			},
		},
	},
}

func interactionCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type == discordgo.InteractionApplicationCommand {
		switch interaction.ApplicationCommandData().Name {
		case "until":
			handleUntil(session, interaction)
		case "timer":
			handleTimer(session, interaction)
		}
	}
}

func handleUntil(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	dateStr := interaction.ApplicationCommandData().Options[0].StringValue()
	date, err := parseTime(dateStr)
	if err != nil {
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid date format",
			},
		})

		if err != nil {
			fmt.Println("Error sending message in date parsing failed in handleUntil():", err)
		}

		return
	}

	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Time until",
					Description: fmt.Sprintf("Time until <t:%d:F>: %s", date.Unix(), humanize.Time(date)),
					Color:       0x00ff00,
				},
			},
		},
	})

	if err != nil {
		fmt.Println("Error sending message in handleUntil():", err)
	}
}
