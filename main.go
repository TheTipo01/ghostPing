package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	Token          string
	DataSourceName string
	driverName     = "mysql"
	database       *sql.DB
)

//Structure for holding a discord ping
type GhostPing struct {
	IdMenzionato   string
	Timestamp      time.Time
	IdMenzionatore string
	IdServer       string
	IdCanale       string
	IdMessaggio    string
	Eliminato      bool
}

func init() {
	var err error
	//Token
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&DataSourceName, "d", "", "Data source name")
	flag.Parse()

	//Prepare the tables
	database, err = sql.Open(driverName, DataSourceName)
	if err != nil {
		log.Println("Error opening db connection,", err)
		return
	}

	execQuery(tblUsers, database)
	execQuery(tblServers, database)
	execQuery(tblChannels, database)
	execQuery(tblPings, database)
}

func main() {
	//Discord
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("Error creating Discord session,", err)
		return
	}

	//Handler for discord events
	dg.AddHandler(messageCreate)
	dg.AddHandler(messageUpdate)
	dg.AddHandler(messageDeleted)

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}

	err = dg.UpdateStatus(0, "ghostpin.ga")
	if err != nil {
		fmt.Println("Can't set status,", err)
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err = dg.Close()
	if err != nil {
		log.Println("Error closing discord session,", err)
	}

	err = database.Close()
	if err != nil {
		log.Println("Error closing database,", err)
	}
}

//Function for checking messages and taking action
func messageCheck(m *discordgo.Message, s *discordgo.Session) {

	//Check if someone has been mentioned
	if len(m.Mentions) != 0 {
		for _, menzioni := range m.Mentions {
			ping := GhostPing{menzioni.ID, time.Now(), m.Author.ID, m.GuildID, m.ChannelID, m.ID, false}
			insertion(&ping, s)
		}
	}

	//Check if everyone is mentioned
	if m.MentionEveryone {
		ping := GhostPing{"everyone", time.Now(), m.Author.ID, m.GuildID, m.ChannelID, m.ID, false}
		insertion(&ping, s)
	}
}

//Function called whenever a new message is created
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Message.Author.ID == s.State.User.ID {
		return
	}

	messageCheck(m.Message, s)
}

//Function called whenever a message is modified
func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {

	statement, err := database.Prepare("UPDATE pings SET deleted = 1 WHERE messageId = ?")

	if err != nil {
		log.Println("Error preparing query,", err)
	}

	_, err = statement.Exec(m.ID)

	if err != nil {
		log.Println("Error updating deleted message,", err)
	}

	messageCheck(m.Message, s)

}

//Function called whenever a message is deleted for updating the corresponding value in the database
func messageDeleted(_ *discordgo.Session, m *discordgo.MessageDelete) {

	statement, err := database.Prepare("UPDATE pings SET deleted = 1 WHERE messageId = ?")

	if err != nil {
		log.Println("Error preparing query,", err)
	}

	_, err = statement.Exec(m.ID)

	if err != nil {
		log.Println("Error updating deleted message,", err)
	}

}

//Function for handling database insertion of pings
func insertion(ping *GhostPing, s *discordgo.Session) {
	//Menzionatore
	statement, err := database.Prepare("INSERT INTO users (id, nickname) VALUES (?, ?)")
	if err != nil {
		log.Println("Error preparing query,", err)
	}
	member, _ := s.GuildMember(ping.IdServer, ping.IdMenzionatore)

	_, err = statement.Exec(ping.IdMenzionatore, member.User.Username)
	if err != nil {
		log.Println("Error inserting menzionatore into the database,", err)
	}

	//Menzionato
	statement, err = database.Prepare("INSERT INTO users (id, nickname) VALUES (?, ?)")
	if err != nil {
		log.Println("Error preparing query,", err)
	}
	member, _ = s.GuildMember(ping.IdServer, ping.IdMenzionato)

	_, err = statement.Exec(ping.IdMenzionato, member.User.Username)
	if err != nil {
		log.Println("Error inserting menzionato into the database,", err)
	}

	//Server
	statement, err = database.Prepare("INSERT INTO server (id, name) VALUES (?, ?)")
	if err != nil {
		log.Println("Error preparing query,", err)
	}
	server, _ := s.Guild(ping.IdServer)

	_, err = statement.Exec(ping.IdServer, server.Name)
	if err != nil {
		log.Println("Error inserting server into the database,", err)
	}

	//Channel
	statement, err = database.Prepare("INSERT INTO channels (id, name, serverId) VALUES (?, ?, ?)")
	if err != nil {
		log.Println("Error preparing query,", err)
	}
	channel, _ := s.Channel(ping.IdCanale)

	_, err = statement.Exec(ping.IdCanale, channel.Name, ping.IdServer)
	if err != nil {
		log.Println("Error inserting channel into the database,", err)
	}

	//Ping
	statement, err = database.Prepare("INSERT INTO pings (menzionatoreId, menzionatoId, channelId, serverId, timestamp, messageId) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("Error preparing query,", err)
	}

	_, err = statement.Exec(ping.IdMenzionatore, ping.IdMenzionato, ping.IdCanale, ping.IdServer, ping.Timestamp, ping.IdMessaggio)
	if err != nil {
		log.Println("Error inserting ping into the database,", err)
	}

}
