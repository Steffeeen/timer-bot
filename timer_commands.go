package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func respondWithLog(session *discordgo.Session, interaction *discordgo.Interaction, response *discordgo.InteractionResponse, context string) {
	err := session.InteractionRespond(interaction, response)
	if err != nil {
		fmt.Println("Error responding to interaction in", context+":", err)
	}
}

func respondWithError(session *discordgo.Session, interaction *discordgo.Interaction, errorStr string, context string, err error) {
	respondWithLog(session, interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: errorStr,
		},
	}, context)
	if err != nil {
		fmt.Println("Error in", context+":", err)
	}
}

func handleTimer(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.ApplicationCommandData().Options[0].Name {
	case "create":
		handleTimerCreate(session, interaction)
	case "list":
		handleTimerList(session, interaction)
	case "delete":
		handleTimerDelete(session, interaction)
	case "edit":
		handleTimerEdit(session, interaction)
	case "snooze":
		handleTimerSnooze(session, interaction)
	}
}

func handleTimerCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	message := options[0].StringValue()
	timeStr := options[1].StringValue()

	date, err := parseTime(timeStr)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Invalid date format", "handleTimerCreate() error case parsing time", err)
		return
	}

	id, err := newTimerID()
	if err != nil {
		respondWithError(session, interaction.Interaction, "Error creating timer ID", "handleTimerCreate() error case in creating timer id", err)
		return
	}

	user := getUserFromInteraction(interaction)
	timer, err := createTimer(id, message, user.ID, interaction.ChannelID, date)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Error creating timer", "handleTimerCreate() error case in creating timer", err)
		return
	}

	respondWithLog(session, interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				createTimerEmbed(timer, user, TimerEmbedTypeCreation),
			},
		},
	}, "handleTimerCreate() success case")
}

func handleTimerList(session *discordgo.Session, i *discordgo.InteractionCreate) {
	timers, err := getAllTimersForUser(i.Member.User.ID, true)
	if err != nil {
		respondWithError(session, i.Interaction, "Error getting timers", "handleTimerList() getting timers", err)
		return
	}

	if len(timers) == 0 {
		respondWithLog(session, i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have no active timers.",
			},
		}, "handleTimerList() no timers")
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

	respondWithLog(session, i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}, "handleTimerList() success case")
}

func handleTimerDelete(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	timerID := options[0].StringValue()

	timer, err := getTimerByID(timerID)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Invalid timer ID", "handleTimerDelete() invalid timer id", err)
		return
	}

	user := getUserFromInteraction(interaction)

	if timer.User != user.ID {
		respondWithError(session, interaction.Interaction, "You do not own this timer", "handleTimerDelete() not owner", err)
		return
	}

	err = deleteTimer(timerID)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Error deleting timer", "handleTimerDelete() error deleting timer", err)
		return
	}

	respondWithLog(session, interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				createTimerEmbed(timer, user, TimerEmbedTypeDeletion),
			},
		},
	}, "handleTimerDelete() success case")
}

func handleTimerEdit(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	timerID := options[0].StringValue()

	timer, err := getTimerByID(timerID)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Invalid timer ID", "handleTimerEdit() invalid timer id", err)
		return
	}

	user := getUserFromInteraction(interaction)

	if timer.User != user.ID {
		respondWithError(session, interaction.Interaction, "You do not own this timer", "handleTimerEdit() not owner", err)
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
			date, err := parseTime(opt.StringValue())
			if err != nil {
				respondWithError(session, interaction.Interaction, "Invalid date format", "handleTimerEdit() invalid date format", err)
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
		respondWithError(session, interaction.Interaction, "Error updating timer", "handleTimerEdit() error updating timer", err)
		return
	}

	respondWithLog(session, interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				createTimerEmbed(timer, getUserFromInteraction(interaction), TimerEmbedTypeEdit),
			},
		},
	}, "handleTimerEdit() success case")
}

func handleTimerSnooze(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	timerID := options[0].StringValue()
	timeStr := options[1].StringValue()

	timer, err := getTimerByID(timerID)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Invalid timer ID", "handleTimerSnooze() invalid timer id", err)
		return
	}

	user := getUserFromInteraction(interaction)

	if timer.User != user.ID {
		respondWithError(session, interaction.Interaction, "You do not own this timer", "handleTimerSnooze() not owner", err)
		return
	}

	date, err := parseTime(timeStr)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Invalid date format", "handleTimerSnooze() invalid date format", err)
		return
	}

	err = snoozeTimer(timerID, date)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Error snoozing timer", "handleTimerSnooze() error snoozing timer", err)
		return
	}

	snoozedTimer, err := getTimerByID(timerID)
	if err != nil {
		respondWithError(session, interaction.Interaction, "Error getting snoozed timer", "handleTimerSnooze() error getting snoozed timer", err)
		return
	}

	respondWithLog(session, interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				createTimerEmbed(snoozedTimer, getUserFromInteraction(interaction), TimerEmbedTypeSnooze),
			},
		},
	}, "handleTimerSnooze() success case")
}

func getUserFromInteraction(interaction *discordgo.InteractionCreate) *discordgo.User {
	if interaction.Member != nil {
		return interaction.Member.User
	}
	return interaction.User
}
