package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//city model
type CityModel struct {
	CityName   string   `json:"cityname"`
	CityLatLng string   `json:"latlng"`
	Chats      []int64  `json:"chats"`
	Last_urls  []string `json:"last_urls"`
}

//auto-checker
type YChecker struct {
	interval     time.Duration
	quit         chan struct{}
	mongoSession *mgo.Session
	yFinder      *YoutubeFinder
	bot          *tgbotapi.BotAPI
}

//ctor
func NewYChecker(interval time.Duration, yFinder *YoutubeFinder, bot *tgbotapi.BotAPI) *YChecker {
	return &YChecker{interval, make(chan struct{}), nil, yFinder, bot}
}
func (yc *YChecker) Start() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	yc.mongoSession = session
	ticker := time.NewTicker(yc.interval)
	for {
		select {
		case <-ticker.C:
			yc.CheckAndSend()
		case <-yc.quit:
			ticker.Stop()
			defer yc.mongoSession.Close()
			return
		}
	}
}

func (yc *YChecker) CheckAndSend() {
	var cities []CityModel
	yc.mongoSession.DB(dbname).C(citiesCollection).Find(bson.M{}).All(&cities)
	for _, c := range cities {
		urls := yc.yFinder.searchListByLocation(c.CityLatLng, "")
		newUrls := yc.saveUrlsForCity(c, urls)
		for _, ci := range c.Chats {
			if len(newUrls) > 0 {
				infoMsg := tgbotapi.NewMessage(ci, fmt.Sprintf("%s [%v]", c.CityName, len(newUrls)))
				yc.bot.Send(infoMsg)
				for _, u := range newUrls {
					msg := tgbotapi.NewMessage(ci, u)
					yc.bot.Send(msg)
				}
			}
		}
	}
}

func (yc *YChecker) saveUrlsForCity(c CityModel, urls []string) []string {
	var newUrls []string

	if len(c.Last_urls) == 0 {
		newUrls = urls
	} else {
		for _, i1 := range urls {
			found := false
			for _, i2 := range c.Last_urls {
				if i1 == i2 {
					found = true
					break
				}
			}
			if found == false {
				newUrls = append(newUrls, i1)
			}

		}
	}

	delCount := len(c.Last_urls) - maxUrlCount - len(urls)
	if delCount > 0 {
		c.Last_urls = c.Last_urls[delCount:]
	}
	c.Last_urls = append(c.Last_urls, newUrls...)
	fmt.Println(c.CityName, len(c.Last_urls))
	lowerCity := strings.ToLower(c.CityName)
	selector := bson.M{"cityname": lowerCity}
	yc.mongoSession.DB(dbname).C(citiesCollection).UpdateAll(selector, bson.M{"$set": bson.M{"last_urls": c.Last_urls}})

	return newUrls
}

func (yc *YChecker) Stop() {
	close(yc.quit)
}

//requests city for user
func (yc *YChecker) RequestCheck(chatId int64, cityName string) {
	var cities []CityModel
	lowerCity := strings.ToLower(cityName)
	selector := bson.M{"cityname": lowerCity}
	err := yc.mongoSession.DB(dbname).C(citiesCollection).Find(selector).All(&cities)
	if err != nil {
		panic(err)
	}
	if len(cities) == 0 {
		res := geocode(cityName, developerKey)
		if len(res.Results) == 0 {
			return
		}
		location := res.Results[0].Geometry.Location
		newCity := CityModel{}
		newCity.CityName = lowerCity
		newCity.CityLatLng = fmt.Sprintf("%v,%v", location.Lat, location.Lng)
		newCity.Chats = make([]int64, 0)
		newCity.Chats = append(newCity.Chats, chatId)
		newCity.Last_urls = make([]string, 0)
		yc.mongoSession.DB(dbname).C(citiesCollection).Insert(newCity)
	} else {
		yc.mongoSession.DB(dbname).C(citiesCollection).UpdateAll(selector, bson.M{"$addToSet": bson.M{"chats": chatId}})
	}
}
