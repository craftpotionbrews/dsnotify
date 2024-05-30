package dsnotify

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

const (
	serverName = "DSNotify"
)

var (
	streamers = make(map[string]*Streamer)
)

type Streamer struct {
	User      string
	Nick      string
	Guild     string
	Channel   string
	Timestamp time.Time
}

func Listen() {
	client, err := newClient()
	if err != nil {
		log.Fatalf("[ERROR] Unable to obtain discord client: %v", err)
		return
	}

	client.AddHandler(ready)
	client.AddHandler(voice)

	// https://discord.com/developers/docs/topics/gateway#list-of-intents
	client.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	if err := client.Open(); err != nil {
		log.Fatalf("[ERROR] Unable to open websocket connection to discord: %v", err)
	}
	defer client.Close()

	log.Printf("[INFO] %s is now playing %s.  Press CTRL-C to exit.", serverName, readyMessage)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func newClient() (*discordgo.Session, error) {
	connection := ""

	if config.Auth.Token != "" {
		connection = "Bot " + config.Auth.Token
	}

	if config.Auth.Bearer != "" {
		connection = "Bearer " + config.Auth.Bearer
	}

	session, err := discordgo.New(connection)
	if err != nil {
		return &discordgo.Session{}, fmt.Errorf("unable to get new discordgo client: %v", err)
	}

	return session, nil
}

func debug(v *discordgo.VoiceStateUpdate) {
	if config.Guilds[v.GuildID].Debug {
		spew.Dump(v)
	}
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, readyMessage); err != nil {
		log.Printf("[ERROR] Failed to update ready message: %v", err)
	}
}

func voice(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// Ignore events unless config enabled
	if !config.Guilds[v.GuildID].Enabled {
		return
	}

	// Ignore events that aren't streams
	if !v.VoiceState.SelfStream {
		return
	}

	// When users force quit discord, a voice state update occurs with
	// no channelID, and therefore not enough details to process
	if v.ChannelID == "" {
		return
	}

	debug(v)
	process(s, v)
}

func process(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	expire()

	streamer := newStreamer(v)

	if _, ok := streamers[streamer.User]; ok {
		log.Printf("[INFO] Updating known user in cache: %s (aka: %s)", streamer.User, streamer.Nick)
		save(streamer)
	} else {
		log.Printf("[INFO] Adding unknown user to cache with notification: %s (aka: %s)", streamer.User, streamer.Nick)
		save(streamer)
		notify(streamer, s)
	}
}

func newStreamer(v *discordgo.VoiceStateUpdate) *Streamer {
	streamer := &Streamer{
		User:      v.UserID,
		Nick:      v.Member.User.Username,
		Guild:     v.GuildID,
		Channel:   v.ChannelID,
		Timestamp: time.Now().UTC(),
	}

	if v.Member.Nick != "" {
		streamer.Nick = v.Member.Nick
	}

	return streamer
}

func save(s *Streamer) {
	streamers[s.User] = s
}

func notify(st *Streamer, s *discordgo.Session) {
	notifyChannel := config.Guilds[st.Guild].NotifyChannel
	notifyRole := config.Guilds[st.Guild].NotifyRole
	streamChannel, err := s.Channel(st.Channel)

	// Construct the notification in two parts to prevent failure
	// if there are transient issues in channel resolution
	msg := fmt.Sprintf("<@&%s> %s is going live", notifyRole, st.Nick)

	if err == nil {
		msg += fmt.Sprintf(" in channel %s", streamChannel.Name)
	}

	go retryLinear(3, 2, func() error {
		_, err := s.ChannelMessageSend(notifyChannel, msg)
		return err
	})
}

func expire() {
	threshold := time.Now().UTC().Add(-1 * time.Hour)

	for uid, streamer := range streamers {
		if streamer.Timestamp.After(threshold) {
			continue
		}

		log.Printf("[INFO] Removing inactive user from cache: %s (aka: %s)", streamer.User, streamer.Nick)
		delete(streamers, uid)
	}
}

func retryLinear(times int, interval int, f func() error) error {
	var err error
	if err = f(); err == nil {
		return err
	}

	for i := 0; i < (times - 1); i++ {
		time.Sleep(time.Duration(interval) * time.Second)

		if err = f(); err == nil {
			return err
		}
	}

	return err
}
