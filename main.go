package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var command = &discordgo.ApplicationCommand{
	Name:        "criacanal",
	Description: "Cria um canal que será deletado em 30 segundos.",
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env")
	}

	logFile, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Failed to open bot.log:", err)
	}
	defer logFile.Close()

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN not set in .env")
	}

	guildID := os.Getenv("GUILD_ID")
	if guildID == "" {
		log.Fatal("GUILD_ID not set in .env")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Failed to create Discord session:", err)
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(onInteractionCreate)

	err = dg.Open()
	if err != nil {
		log.Fatal("Failed to connect to Discord:", err)
	}
	defer dg.Close()

	commands, _ := dg.ApplicationCommands(dg.State.User.ID, guildID)
	for _, cmd := range commands {
		_ = dg.ApplicationCommandDelete(dg.State.User.ID, guildID, cmd.ID)
		log.Printf("Deleted command: %s", cmd.Name)
	}

	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, guildID, command)
	if err != nil {
		log.Fatalf("Failed to register command %s: %v", command.Name, err)
	}
	log.Printf("Command registered: /%s", command.Name)

	log.Println("Bot is running. Press Ctrl+C to stop.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Bot shutting down...")
}

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if i.ApplicationCommandData().Name == "criacanal" {
		user := i.Member.User
		log.Printf("/criacanal used by %s (%s)", user.Username, user.ID)

		channel, err := s.GuildChannelCreate(i.GuildID, "canal-do-"+user.Username, discordgo.ChannelTypeGuildText)
		if err != nil {
			log.Printf("Failed to create channel: %v", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Canal não foi criado.",
				},
			})
			return
		}

		log.Printf("Channel #%s created for %s", channel.Name, user.Username)

		s.ChannelMessageSend(channel.ID, fmt.Sprintf("Canal criado para <@%s>! Ele será deletado em 30 segundos...", user.ID))

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Canal criado: <#%s>", channel.ID),
			},
		})

		go func(channelID string) {
			time.Sleep(30 * time.Second)

			_, err := s.ChannelDelete(channelID)
			if err != nil {
				log.Printf("Failed to delete channel %s: %v", channelID, err)
			} else {
				log.Printf("Channel deleted: %s", channelID)
			}
		}(channel.ID)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.Content == "tivo" || m.Content == "tvo" || m.Content == "Tivo" || m.Content == "Tvo" {
		log.Printf("Triggered response: %s sent %s", m.Author.Username, m.Content)
		s.ChannelMessageSend(m.ChannelID, "vai tomar no seu cu irmao")
	}
}
