package main

import (
	"log"
	"os"
	"time"

	"github.com/kgantsov/stockholm_commute_bot/pkg/client"
	"github.com/kgantsov/stockholm_commute_bot/pkg/handlers"
	"github.com/kgantsov/stockholm_commute_bot/pkg/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	tb "gopkg.in/tucnak/telebot.v2"
)

func worker(app *handlers.App) {
	c := app.Session.DB("commute_bot").C("users")

	ticker := time.NewTicker(time.Minute * 1)

	for t := range ticker.C {
		var users []models.User
		c.Find(bson.M{"home_time": t.Format(time.Kitchen)}).All(&users)

		for _, user := range users {
			trips := app.Sl.GetHomeTrips(user)

			for _, trip := range trips.Trip {
				app.Bot.Send(
					&tb.Chat{ID: user.ChatID}, app.Sl.GetMessageForTrip(trip), tb.ModeMarkdown,
				)
			}
		}

		c.Find(bson.M{"work_time": t.Format(time.Kitchen)}).All(&users)

		for _, user := range users {
			trips := app.Sl.GetWorkTrips(user)

			for _, trip := range trips.Trip {
				app.Bot.Send(
					&tb.Chat{ID: user.ChatID}, app.Sl.GetMessageForTrip(trip), tb.ModeMarkdown,
				)
			}
		}
	}
}

func main() {
	session, err := mgo.Dial(os.Getenv("MONGODB_URLS"))
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	bot, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	sl := client.NewSLClient()

	app := handlers.App{Session: session, Bot: bot, Sl: sl}

	bot.Handle("/start", handlers.StartHandler(&app))
	bot.Handle("/home", handlers.HomeHandler(&app))
	bot.Handle("/work", handlers.WorkHandler(&app))
	bot.Handle("/set_home", handlers.SetHomeHandler(&app))
	bot.Handle("/set_home_reminder", handlers.SetHomeReminderHandler(&app))
	bot.Handle("/set_work", handlers.SetWorkHandler(&app))
	bot.Handle("/set_work_reminder", handlers.SetWorkReminderHandler(&app))

	go worker(&app)

	bot.Start()
}
