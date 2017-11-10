/*
Package bot contains implementation of the Calories bot.
The bot's Message handlers and functions are described in this package.
*/
package bot

import (
	"log"
	"net/http"

	"github.com/bobheadxi/calories/facebook"
	"github.com/bobheadxi/calories/server"
)

// Bot : The Calories bot of the app.
type Bot struct {
	api    *facebook.API
	server *server.Server
}

// New : Sets up and returns a Bot
func New(api *facebook.API, sv *server.Server) *Bot {
	b := Bot{
		api:    api,
		server: sv,
	}
	b.api.MessageHandler = b.TestMessageReceivedAndReply
	return &b
}

// Run : Spins up the Calories bot
func (b *Bot) Run(port string) {
	http.HandleFunc("/webhook", b.api.Handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// TestMessageReceivedAndReply : Tests that bot receives messages and replies.
// DEPRECATE ASAP - replace with Bot handlers or something
func (b *Bot) TestMessageReceivedAndReply(event facebook.Event, sender facebook.Sender, msg facebook.ReceivedMessage) {
	b.api.SendTextMessage(sender.ID, "Hello!")
	response, err := b.server.SumCalories(sender.ID)
	if err != nil {
		log.Print("No calories for you" + err.Error())
	}
	b.api.SendTextMessage(sender.ID, "your total calories are "+string(response))
}
