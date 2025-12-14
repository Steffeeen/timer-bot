package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

type Timer struct {
	InternalID  int
	ID          string
	Message     string
	User        string
	Channel     string
	Created     time.Time
	Due         time.Time
	SnoozedDue  time.Time
	SnoozeCount int
	Shown       bool
}

func checkDueTimers(session *discordgo.Session) {
	timers, err := getDueTimers()
	if err != nil {
		fmt.Println("Error getting due timers:", err)
		return
	}

	for _, timer := range timers {
		showDueTimer(session, timer)
		err := markTimerAsShown(timer.ID)
		if err != nil {
			fmt.Println("Error marking timer as shown:", err)
		}
	}
}

func showDueTimer(session *discordgo.Session, timer *Timer) {
	user, err := session.User(timer.User)
	if err != nil {
		fmt.Println("Error getting user:", err)
		return
	}

	embed := createTimerEmbed(timer, user, TimerEmbedTypeDue)

	_, err = session.ChannelMessageSendEmbed(timer.Channel, embed)
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

type TimerEmbedType struct {
	Title                     string
	color                     int
	includeDurationForDue     bool
	includeDurationForCreated bool
}

var (
	TimerEmbedTypeCreation = TimerEmbedType{"Timer Created", 0x00ff00, true, false}
	TimerEmbedTypeDeletion = TimerEmbedType{"Timer Deleted", 0xff0000, false, true}
	TimerEmbedTypeSnooze   = TimerEmbedType{"Timer Snoozed", 0x00ffff, true, true}
	TimerEmbedTypeEdit     = TimerEmbedType{"Timer Edited", 0xffff00, true, true}
	TimerEmbedTypeDue      = TimerEmbedType{"Timer Due", 0x0000ff, false, true}
)

func createTimerEmbed(timer *Timer, owner *discordgo.User, embedType TimerEmbedType) *discordgo.MessageEmbed {
	due :=

	return &discordgo.MessageEmbed{
		Title:       "Timer Due",
		Description: timer.Message,
		Color:       0x0000ff,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Id",
				Value: timer.ID,
				Inline: true,
			},
			{
				Name: "Owner",
				Value: owner.Mention(),
				Inline: true,
			},
			{
				Name: "\u200b",
				Value: "\u200b",
				Inline: true,
			},
			{
				Name: "Due",

			},
		},
	}
}

func formatTime(timeToFormat time.Time, includeDuration bool) string {
	base := timeToFormat.In(time.Local).
		Format("02/01/2006, 15:04:05")

}
