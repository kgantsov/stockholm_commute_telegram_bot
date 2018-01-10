package main

import (
	"log"
	"os"
	"time"

	"github.com/kgantsov/stockholm_commute_bot/pkg/client"
	"github.com/kgantsov/stockholm_commute_bot/pkg/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	session, err := mgo.Dial(os.Getenv("MONGODB_URLS"))
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB("commute_bot").C("users")

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
		var u models.User
		err := c.Find(bson.M{"id": m.Sender.ID}).One(&u)

		if err != nil {
			b.Send(m.Sender, "Please setup home and work locations", tb.ModeMarkdown)
		}

		trips := sl.GetHomeTrips(u)

		for _, trip := range trips.Trip {
			b.Send(m.Sender, sl.GetMessageForTrip(trip), tb.ModeMarkdown)
		}
	})

	b.Handle("/work", func(m *tb.Message) {
		var u models.User
		err := c.Find(bson.M{"id": m.Sender.ID}).One(&u)

		if err != nil {
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
					var user models.User
					err := c.Find(bson.M{"id": m.Sender.ID}).One(&user)

					if err == nil {
						err = c.Update(
							bson.M{"id": m.Sender.ID},
							&models.User{
								ID:       m.Sender.ID,
								Name:     m.Sender.FirstName,
								HomeID:   st.SiteID,
								HomeName: st.Name,
								WorkName: user.WorkName,
								WorkID:   user.WorkID,
							},
						)
						if err != nil {
							log.Fatal(err)
						}
					} else {
						err = c.Insert(
							&models.User{
								ID:       m.Sender.ID,
								Name:     m.Sender.FirstName,
								HomeID:   st.SiteID,
								HomeName: st.Name,
								WorkName: "",
								WorkID:   "",
							},
						)
						if err != nil {
							log.Fatal(err)
						}
					}

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
					var user models.User
					err := c.Find(bson.M{"id": m.Sender.ID}).One(&user)

					if err == nil {
						err = c.Update(
							bson.M{"id": m.Sender.ID},
							&models.User{
								ID:       m.Sender.ID,
								Name:     m.Sender.FirstName,
								HomeID:   user.HomeID,
								HomeName: user.HomeName,
								WorkID:   st.SiteID,
								WorkName: st.Name,
							},
						)
						if err != nil {
							log.Fatal(err)
						}
					} else {
						err = c.Insert(
							&models.User{
								ID:       m.Sender.ID,
								Name:     m.Sender.FirstName,
								HomeID:   "",
								HomeName: "",
								WorkID:   st.SiteID,
								WorkName: st.Name,
							},
						)
						if err != nil {
							log.Fatal(err)
						}
					}
				}
			}(station))

			replyKeys = append(replyKeys, []tb.ReplyButton{replyBtn})
		}

		b.Send(m.Sender, "Choose location:", &tb.ReplyMarkup{ReplyKeyboard: replyKeys})
	})

	b.Start()
}
