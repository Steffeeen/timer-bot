package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/markusmobius/go-dateparser"
)

func handleTimer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Options[0].Name {
	case "create":
		handleTimerCreate(s, i)
	case "list":
		handleTimerList(s, i)
	case "delete":
		handleTimerDelete(s, i)
	case "edit":
		handleTimerEdit(s, i)
	case "snooze":
		handleTimerSnooze(s, i)
	}
}

func handleTimerCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	message := options[0].StringValue()
	timeStr := options[1].StringValue()

	date, err := parseDate(timeStr)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid date format",
			},
		})
		return
	}

	id, err := newTimerID()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error creating timer ID",
			},
		})
		return
	}

	timer, err := createTimer(id, message, i.Member.User.ID, i.ChannelID, date)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error creating timer",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Timer Created",
					Description: timer.Message,
					Color:       0x00ff00,
					Fields: []*discordgo.MessageEmbedField{
						{Name: "ID", Value: timer.ID, Inline: true},
						{Name: "Owner", Value: i.Member.User.Mention(), Inline: true},
						{Name: "Due", Value: fmt.Sprintf("<t:%d:R>", timer.Due.Unix()), Inline: true},
					},
				},
			},
		},
	})
}

func handleTimerList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	timers, err := getAllTimersForUser(i.Member.User.ID, true)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error getting timers",
			},
		})
		return
	}

	if len(timers) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have no active timers.",
			},
		})
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Active Timers",
		Color: 0x3c1984,
	}

	for _, timer := range timers {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  timer.ID,
			Value: fmt.Sprintf("%s - Due: <t:%d:R>", timer.Message, timer.SnoozedDue.Unix()),
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handleTimerDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	timerID := options[0].StringValue()

	timer, err := getTimerByID(timerID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid timer ID",
			},
		})
		return
	}

	if timer.User != i.Member.User.ID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not own this timer",
			},
		})
		return
	}

	err = deleteTimer(timerID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error deleting timer",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Timer Deleted",
					Description: timer.Message,
					Color:       0xff0000,
					Fields: []*discordgo.MessageEmbedField{
						{Name: "ID", Value: timer.ID, Inline: true},
					},
				},
			},
		},
	})
}

func handleTimerEdit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	timerID := options[0].StringValue()

	timer, err := getTimerByID(timerID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid timer ID",
			},
		})
		return
	}

	if timer.User != i.Member.User.ID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not own this timer",
			},
		})
		return
	}

	var newMessage *string
	var newTime *time.Time

	for _, opt := range options[1:] {
		switch opt.Name {
		case "message":
			val := opt.StringValue()
			newMessage = &val
		case "time":
			date, err := parseDate(opt.StringValue())
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid date format",
					},
				})
				return
			}
			newTime = &date
		}
	}

	if newMessage != nil {
		timer.Message = *newMessage
	}
	if newTime != nil {
		timer.Due = *newTime
		timer.SnoozedDue = *newTime
	}

	err = updateTimer(timer)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error updating timer",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Timer Edited",
					Description: timer.Message,
					Color:       0xffff00,
					Fields: []*discordgo.MessageEmbedField{
						{Name: "ID", Value: timer.ID, Inline: true},
						{Name: "Due", Value: fmt.Sprintf("<t:%d:R>", timer.Due.Unix()), Inline: true},
					},
				},
			},
		},
	})
}

func handleTimerSnooze(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options[0].Options
	timerID := options[0].StringValue()
	timeStr := options[1].StringValue()

	timer, err := getTimerByID(timerID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid timer ID",
			},
		})
		return
	}

	if timer.User != i.Member.User.ID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not own this timer",
			},
		})
		return
	}

	date, err := parseDate(timeStr)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid date format",
			},
		})
		return
	}

	err = snoozeTimer(timerID, date)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error snoozing timer",
			},
		})
		return
	}

	snoozedTimer, err := getTimerByID(timerID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error getting snoozed timer",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Timer Snoozed",
					Description: snoozedTimer.Message,
					Color:       0x00ffff,
					Fields: []*discordgo.MessageEmbedField{
						{Name: "ID", Value: snoozedTimer.ID, Inline: true},
						{Name: "New Due Date", Value: fmt.Sprintf("<t:%d:R>", snoozedTimer.SnoozedDue.Unix()), Inline: true},
					},
				},
			},
		},
	})
}
