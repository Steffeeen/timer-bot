package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/markusmobius/go-dateparser"
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

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "until":
			handleUntil(s, i)
		case "timer":
			handleTimer(s, i)
		}
	}
}

func handleUntil(s *discordgo.Session, i *discordgo.InteractionCreate) {
	dateStr := i.ApplicationCommandData().Options[0].StringValue()
	date, err := parseDate(dateStr)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid date format",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Time until",
					Description: fmt.Sprintf("Time until <t:%d:F> is <t:%d:R>", date.Unix(), date.Unix()),
					Color:       0x00ff00,
				},
			},
		},
	})
}
