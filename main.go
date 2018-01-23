package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Syfaro/telegram-bot-api"
)

const maxUrlCount = 8
const checkInterval = 1 * time.Minute

var developerKey = os.Getenv("DEVKEY")
var botApiKey = os.Getenv("BOTAPIKEY")
var dbname = "yra"
var citiesCollection = "cities"

func main() {

	if developerKey == "" || botApiKey == "" {
		panic("dev key not found")
	}

	yFinder := New(developerKey)

	bot, err := tgbotapi.NewBotAPI(botApiKey)
	if err != nil {
		fmt.Println(err)
		log.Panic(err)
	}
	yc := NewYChecker(checkInterval, yFinder, bot)
	go yc.Start()

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	var ucfg = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := bot.GetUpdatesChan(ucfg)

	for {
		select {
		case update := <-updates:
			ChatID := update.Message.Chat.ID
			yc.RequestCheck(ChatID, update.Message.Text)

			var replies []string
			geoResult := geocode(update.Message.Text, developerKey)
			if len(geoResult.Results) == 0 {
				replies = append(replies, "there is no city found.")
			} else {
				location := geoResult.Results[0].Geometry.Location
				replies = yFinder.searchListByLocation(fmt.Sprintf("%v,%v", location.Lat, location.Lng), "")
			}

			for _, r := range replies {
				msg := tgbotapi.NewMessage(ChatID, r)
				bot.Send(msg)
			}
		}
	}
}
