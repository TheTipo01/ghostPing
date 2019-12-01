package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	Token string
	mutex = &sync.Mutex{}
	dataSourceName = "./ghostpingers.db"
	driverName = "sqlite3"
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
	//Token
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	//Database
	database, _ := sql.Open(driverName, dataSourceName)
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS stronzi (MenzionatoId TEXT, oraora datetime, MenzionatoreId TEXT, ServerId TEXT, ChannellId TEXT, MessageId TEXT PRIMARY KEY, Eliminato INTEGER)")

	_, err := statement.Exec()
	if err != nil {
		log.Println("Error creating table,", err)
	}

	err = database.Close()

	if err != nil {
		log.Println("Error closing database,", err)
	}
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

	err = dg.Open()
	if err != nil {
		log.Println("Error opening connection,", err)
		return
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err = dg.Close()

	if err != nil {
		log.Println("Error closing discord session,", err)
	}
}

//Function for checking messages and taking action
func messageCheck(m *discordgo.Message) {

	//Check if someone has been mentioned
	if len(m.Mentions) != 0 {
		for _, menzioni := range m.Mentions {
			ping := GhostPing{menzioni.ID, time.Now(), m.Author.ID, m.GuildID, m.ChannelID, m.ID, false}
			insertion(&ping)
		}
	}

	//Check if everyone is mentioned
	if m.MentionEveryone {
		ping := GhostPing{"everyone", time.Now(), m.Author.ID, m.GuildID, m.ChannelID, m.ID, false}
		insertion(&ping)
	}
}

//Function called whenever a new message is created
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Message.Author.ID == s.State.User.ID {
		return
	}

	messageCheck(m.Message)
}

//Function called whenever a message is modified
func messageUpdate(_ *discordgo.Session, m *discordgo.MessageUpdate) {

	messageCheck(m.Message)
}

//Function called whenever a message is deleted for updating the corresponding value in the database
func messageDeleted(_ *discordgo.Session, m *discordgo.MessageDelete) {

	mutex.Lock()
	defer mutex.Unlock()

	database, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Println("Error opening database connection,", err)
	}

	statement, err := database.Prepare("UPDATE stronzi SET Eliminato = 1 WHERE MessageId = ?")

	if err != nil {
		log.Println("Error preparing query,", err)
	}

	res, err := statement.Exec(m.ID)

	if err != nil {
		log.Println("Error updating deleted message,", err)
	}

	err = database.Close()
	if err != nil {
		log.Println("Error closing database connection,", err)
	}

	if rows, _ := res.RowsAffected(); rows > 0 {
		html()
	}

}

//Function for handling database insertion of pings
func insertion(ping *GhostPing) {

	mutex.Lock()
	defer mutex.Unlock()

	database, err := sql.Open(driverName, dataSourceName)

	if err != nil {
		log.Println("Error opening database connection,", err)
	}

	statement, err := database.Prepare("INSERT INTO stronzi (MenzionatoId, oraora, MenzionatoreId, ServerId, ChannellId, MessageId, Eliminato) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("Error preparing query,", err)
	}

	_, err = statement.Exec(ping.IdMenzionato, ping.Timestamp, ping.IdMenzionatore, ping.IdServer, ping.IdCanale, ping.IdMessaggio, ping.Eliminato)
	if err != nil {
		log.Println("Error inserting into the database,", err)
	}

	err = database.Close()
	if err != nil {
		log.Println("Error closing database connection,", err)
	}

}

//Function for generating an HTML page to show a list of pings
func html() {

	//Variables
	var MenzionatoId, MenzionatoreId, ServerId, ChannellId, MessageId string
	var oraora time.Time
	var eliminato bool

	//Opening database
	database, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Println("Error opening database connection,", err)
	}

	//Creating session
	s, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("Error creating session,", err)
	}

	//Querying database
	rows, err := database.Query("SELECT * FROM stronzi WHERE Eliminato = 1 ORDER BY oraora DESC")
	if err != nil {
		log.Println("Error querying database,", err)
	}

	//Various string for formatting html in a tidy way
	altro := "<!DOCTYPE html><html lang=\"it\"><head><title>Ghostpingers</title><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><link rel=\"stylesheet\" href=\"css/bootstrap.min.css\"></head><body><div class=\"container\"><h2>Persone che pingano</h2><p>Lista delle persone che pingano altre persone:</p><table class=\"table table-hover\"><thead><tr><th>Username Menzionato</th><th>Ora e data</th><th>User Menzionatore</th><th>Server</th><th>Canale</th></tr></thead><tbody>"
	mezzo := "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>"

	for rows.Next() {
		err = rows.Scan(&MenzionatoId, &oraora, &MenzionatoreId, &ServerId, &ChannellId, &MessageId, &eliminato)
		if err != nil {
			log.Println("Error scannings rows from query,", err)
		}

		menzionato, _ := s.User(MenzionatoId)
		menzionatore, _ := s.User(MenzionatoreId)
		server, _ := s.Guild(ServerId)
		canale, _ := s.Channel(ChannellId)

		if MenzionatoId != "EVERYONE" {
			altro += fmt.Sprintf(mezzo, menzionato.Username, oraora.Format("02/01/2006 - 15:04:05"), menzionatore.Username, server.Name, canale.Name)
		} else {
			altro += fmt.Sprintf(mezzo, MenzionatoId, oraora.Format("02/01/2006 - 15:04:05"), menzionatore.Username, server.Name, canale.Name)
		}

	}

	f, err := os.Create("./index.html")
	if err != nil {
		log.Println("Error creating file,", err)
	}

	_, err = f.WriteString(altro + "</tbody></table></div></body></html>")
	if err != nil {
		log.Println("Error writing string to file,", err)
	}

	err = f.Close()
	if err != nil {
		log.Println("Error closing file,", err)
	}

	err = s.Close()
	if err != nil {
		log.Println("Error closing session,", err)
	}

	err = database.Close()
	if err != nil {
		log.Println("Error closing database connection,", err)
	}

}
