package handlers

import (
	"log"

	"github.com/kgantsov/stockholm_commute_bot/pkg/client"
	"github.com/kgantsov/stockholm_commute_bot/pkg/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	tb "gopkg.in/tucnak/telebot.v2"
)

type App struct {
	Session *mgo.Session
	Bot     *tb.Bot
	Sl      *client.SLClient
}

func StartHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		app.Bot.Send(m.Sender, "Hi this is a StockholmCommuteBot")
	}
}

func HomeHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		c := app.Session.DB("commute_bot").C("users")

		var u models.User
		err := c.Find(bson.M{"id": m.Sender.ID}).One(&u)

		if err != nil {
			app.Bot.Send(m.Sender, "Please setup home and work locations", tb.ModeMarkdown)
		}

		trips := app.Sl.GetHomeTrips(u)

		for _, trip := range trips.Trip {
			app.Bot.Send(m.Sender, app.Sl.GetMessageForTrip(trip), tb.ModeMarkdown)
		}
	}
}

func WorkHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		c := app.Session.DB("commute_bot").C("users")

		var u models.User
		err := c.Find(bson.M{"id": m.Sender.ID}).One(&u)

		if err != nil {
			app.Bot.Send(m.Sender, "Please setup home and work locations", tb.ModeMarkdown)
		}

		trips := app.Sl.GetWorkTrips(u)

		for _, trip := range trips.Trip {
			app.Bot.Send(m.Sender, app.Sl.GetMessageForTrip(trip), tb.ModeMarkdown)
		}
	}
}

func SetHomeHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		c := app.Session.DB("commute_bot").C("users")

		lookup := app.Sl.GetStationsByName(m.Payload)

		if len(lookup.ResponseData) == 0 {
			app.Bot.Send(m.Sender, "No stations found")
			return
		}

		var replyKeys [][]tb.ReplyButton

		for _, station := range lookup.ResponseData {
			replyBtn := tb.ReplyButton{Text: station.Name}

			app.Bot.Handle(&replyBtn, func(st client.Station) func(m *tb.Message) {
				return func(m *tb.Message) {
					var user models.User
					err := c.Find(bson.M{"id": m.Sender.ID}).One(&user)

					if err == nil {
						err = c.Update(
							bson.M{"id": m.Sender.ID},
							&models.User{
								ID:       m.Sender.ID,
								Name:     m.Sender.FirstName,
								ChatID:   m.Chat.ID,
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
								ChatID:   m.Chat.ID,
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

		app.Bot.Send(m.Sender, "Choose location:", &tb.ReplyMarkup{ReplyKeyboard: replyKeys})
	}
}

func SetWorkHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		c := app.Session.DB("commute_bot").C("users")

		lookup := app.Sl.GetStationsByName(m.Payload)

		if len(lookup.ResponseData) == 0 {
			app.Bot.Send(m.Sender, "No stations found")
			return
		}

		var replyKeys [][]tb.ReplyButton

		for _, station := range lookup.ResponseData {
			replyBtn := tb.ReplyButton{Text: station.Name}

			app.Bot.Handle(&replyBtn, func(st client.Station) func(m *tb.Message) {
				return func(m *tb.Message) {
					var user models.User
					err := c.Find(bson.M{"id": m.Sender.ID}).One(&user)

					if err == nil {
						err = c.Update(
							bson.M{"id": m.Sender.ID},
							&models.User{
								ID:       m.Sender.ID,
								Name:     m.Sender.FirstName,
								ChatID:   m.Chat.ID,
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
								ChatID:   m.Chat.ID,
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

		app.Bot.Send(m.Sender, "Choose location:", &tb.ReplyMarkup{ReplyKeyboard: replyKeys})
	}
}
