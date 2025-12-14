package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	token         = os.Getenv("TOKEN")
	applicationID = os.Getenv("APPLICATION_ID")
)

func main() {
	if token == "" || applicationID == "" {
		fmt.Println("No token or application ID provided")
		return
	}

	err := initDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	session, err := setupDiscordSession()
	if err != nil {
		fmt.Println("Error setting up Discord session:", err)
		return
	}

	registeredCommands, err := registerCommands(session)
	if err != nil {
		fmt.Println("Error registering commands:", err)
	}

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			checkDueTimers(session)
		}
	}()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	unregisterAllCommands(session, registeredCommands)
}

func setupDiscordSession() (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is up!")
	})
	session.AddHandler(interactionCreate)

	err = session.Open()
	if err != nil {
		return nil, err
	}
	defer func(session *discordgo.Session) {
		err = session.Close()
	}(session)

	return session, err
}

func registerCommands(session *discordgo.Session) ([]*discordgo.ApplicationCommand, error) {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, cmd := range commands {
		registered, err := session.ApplicationCommandCreate(applicationID, "", cmd)
		if err != nil {
			return nil, err
		}
		registeredCommands[i] = registered
	}

	return registeredCommands, nil
}

func unregisterAllCommands(session *discordgo.Session, registeredCommands []*discordgo.ApplicationCommand) {
	fmt.Println("Unregistering all commands...")
	for _, cmd := range registeredCommands {
		err := session.ApplicationCommandDelete(applicationID, "", cmd.ID)
		if err != nil {
			fmt.Println("Error deleting command:", err)
		}
	}
	fmt.Println("Unregistered all commands, shutting down...")
}
