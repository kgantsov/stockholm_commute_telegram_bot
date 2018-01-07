package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/kgantsov/stockholm_commute_bot/pkg/client"
	tb "gopkg.in/tucnak/telebot.v2"
)

var userTextMap map[int]client.UserPoints

func main() {
	userTextMap = make(map[int]client.UserPoints)
	var mutex = &sync.RWMutex{}

	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	sl := client.NewSLClient()

	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Sender, "Hi this is a StockholmCommuteBot")
	})

	b.Handle("/home", func(m *tb.Message) {
		mutex.RLock()
		u, ok := userTextMap[m.Sender.ID]
		mutex.RUnlock()

		if !ok {
			b.Send(m.Sender, "Please setup home and work locations", tb.ModeMarkdown)
		}

		trips := sl.GetHomeTrips(u)

		for _, trip := range trips.Trip {
			b.Send(m.Sender, sl.GetMessageForTrip(trip), tb.ModeMarkdown)
		}
	})

	b.Handle("/work", func(m *tb.Message) {
		mutex.RLock()
		u, ok := userTextMap[m.Sender.ID]
		mutex.RUnlock()

		if !ok {
			b.Send(m.Sender, "Please setup home and work locations", tb.ModeMarkdown)
		}

		trips := sl.GetWorkTrips(u)

		for _, trip := range trips.Trip {
			b.Send(m.Sender, sl.GetMessageForTrip(trip), tb.ModeMarkdown)
		}
	})

	b.Handle("/set_home", func(m *tb.Message) {
		lookup := sl.GetStationsByName(m.Payload)

		if len(lookup.ResponseData) == 0 {
			b.Send(m.Sender, "No stations found")
			return
		}

		var replyKeys [][]tb.ReplyButton

		for _, station := range lookup.ResponseData {
			replyBtn := tb.ReplyButton{Text: station.Name}

			b.Handle(&replyBtn, func(st client.Station) func(m *tb.Message) {
				return func(m *tb.Message) {
					mutex.Lock()
					if _, ok := userTextMap[m.Sender.ID]; !ok {
						userTextMap[m.Sender.ID] = client.UserPoints{}
					}

					userTextMap[m.Sender.ID] = client.UserPoints{
						HomeID:   st.SiteID,
						HomeName: st.Name,
						WorkID:   userTextMap[m.Sender.ID].WorkID,
						WorkName: userTextMap[m.Sender.ID].WorkName,
					}
					mutex.Unlock()
				}
			}(station))

			replyKeys = append(replyKeys, []tb.ReplyButton{replyBtn})
		}

		b.Send(m.Sender, "Choose location:", &tb.ReplyMarkup{ReplyKeyboard: replyKeys})
	})

	b.Handle("/set_work", func(m *tb.Message) {
		lookup := sl.GetStationsByName(m.Payload)

		if len(lookup.ResponseData) == 0 {
			b.Send(m.Sender, "No stations found")
			return
		}

		var replyKeys [][]tb.ReplyButton

		for _, station := range lookup.ResponseData {
			replyBtn := tb.ReplyButton{Text: station.Name}

			b.Handle(&replyBtn, func(st client.Station) func(m *tb.Message) {
				return func(m *tb.Message) {
					mutex.Lock()
					if _, ok := userTextMap[m.Sender.ID]; !ok {
						userTextMap[m.Sender.ID] = client.UserPoints{}
					}

					userTextMap[m.Sender.ID] = client.UserPoints{
						HomeID:   userTextMap[m.Sender.ID].HomeID,
						HomeName: userTextMap[m.Sender.ID].HomeName,
						WorkID:   st.SiteID,
						WorkName: st.Name,
					}
					mutex.Unlock()
				}
			}(station))

			replyKeys = append(replyKeys, []tb.ReplyButton{replyBtn})
		}

		b.Send(m.Sender, "Choose location:", &tb.ReplyMarkup{ReplyKeyboard: replyKeys})
	})

	b.Start()
}
