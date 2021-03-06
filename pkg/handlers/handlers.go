package handlers

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

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

// StartHandler is a handler function that returns list of all command bots
func StartHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		app.Bot.Send(
			m.Sender,
			"Hi, this is a @StockholmCommuteBot. It will provide you with information about"+
				"SL shedules in Stockholm. There is a list of commands available in this bot: \n\n"+
				"/start - Start communication with the bot\n"+
				"/home - Get schedule from work to home\n"+
				"/work - Get schedule from home to work\n"+
				"/set_home - Set home location\n"+
				"/set_home_reminder - Set home reminder\n"+
				"/set_work - Set work location\n"+
				"/set_work_reminder - Set work reminder\n",
		)
	}
}

// HomeHandler is a handler function that sends information about trips to a home destination
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

// WorkHandler is a handler function that sends information about trips to a work destination
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

// SetHomeHandler is a handler function that receives home location and saves it to a DB
func SetHomeHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		log.Debug(fmt.Sprintf("SetHomeHandler triggered <%s>", m.Payload))

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
								HomeTime: user.HomeTime,
								HomeName: st.Name,
								WorkName: user.WorkName,
								WorkID:   user.WorkID,
								WorkTime: user.WorkTime,
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
								HomeTime: "",
								WorkName: "",
								WorkID:   "",
								WorkTime: "",
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

// SetHomeReminderHandler is a handler function that receives home reminder and saves it to a DB
func SetHomeReminderHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		log.Debug(fmt.Sprintf("SetHomeReminderHandler triggered <%s>", m.Payload))

		c := app.Session.DB("commute_bot").C("users")

		const longForm = "Monday, 02-Jan-06 15:04:05 -0700"
		t, err := time.Parse(longForm, fmt.Sprintf("Monday, 12-Jan-16 %s:25 +0200", m.Payload))

		if err != nil {
			app.Bot.Send(m.Sender, "Time should be send in a format: 16:35")
			return
		}

		var user models.User
		err = c.Find(bson.M{"id": m.Sender.ID}).One(&user)

		if err == nil {
			err = c.Update(
				bson.M{"id": m.Sender.ID},
				&models.User{
					ID:       m.Sender.ID,
					Name:     m.Sender.FirstName,
					ChatID:   m.Chat.ID,
					HomeID:   user.HomeID,
					HomeTime: t.UTC().Format(time.Kitchen),
					HomeName: user.HomeName,
					WorkName: user.WorkName,
					WorkID:   user.WorkID,
					WorkTime: user.WorkTime,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			app.Bot.Send(m.Sender, "You need to set up home location first using /set_home command")
		}
	}
}

// SetWorkHandler is a handler function that receives work location and saves it to a DB
func SetWorkHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		log.Debug(fmt.Sprintf("SetWorkHandler triggered <%s>", m.Payload))

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
								HomeTime: user.HomeTime,
								WorkID:   st.SiteID,
								WorkName: st.Name,
								WorkTime: user.WorkTime,
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
								HomeTime: "",
								WorkID:   st.SiteID,
								WorkName: st.Name,
								WorkTime: "",
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

// SetWorkReminderHandler is a handler function that receives work reminder and saves it to a DB
func SetWorkReminderHandler(app *App) func(m *tb.Message) {
	return func(m *tb.Message) {
		log.Debug(fmt.Sprintf("SetWorkReminderHandler triggered <%s>", m.Payload))

		c := app.Session.DB("commute_bot").C("users")

		const longForm = "Monday, 02-Jan-06 15:04:05 -0700"
		t, err := time.Parse(longForm, fmt.Sprintf("Monday, 12-Jan-16 %s:25 +0200", m.Payload))

		if err != nil {
			app.Bot.Send(m.Sender, "Time should be send in a format: 16:35")
			return
		}

		var user models.User
		err = c.Find(bson.M{"id": m.Sender.ID}).One(&user)

		if err == nil {
			err = c.Update(
				bson.M{"id": m.Sender.ID},
				&models.User{
					ID:       m.Sender.ID,
					Name:     m.Sender.FirstName,
					ChatID:   m.Chat.ID,
					HomeID:   user.HomeID,
					HomeTime: user.HomeTime,
					HomeName: user.HomeName,
					WorkName: user.WorkName,
					WorkID:   user.WorkID,
					WorkTime: t.UTC().Format(time.Kitchen),
				},
			)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			app.Bot.Send(m.Sender, "You need to set up work location first using /set_work command")
		}
	}
}
