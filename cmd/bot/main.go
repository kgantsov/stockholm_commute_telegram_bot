package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kgantsov/stockholm_commute_bot/pkg/client"
	"github.com/kgantsov/stockholm_commute_bot/pkg/handlers"
	"github.com/kgantsov/stockholm_commute_bot/pkg/models"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	tb "gopkg.in/tucnak/telebot.v2"
)

func worker(app *handlers.App) {
	log.Info("Starting reminder worker...")

	c := app.Session.DB("commute_bot").C("users")

	ticker := time.NewTicker(time.Minute * 1)

	for t := range ticker.C {
		var users []models.User
		c.Find(bson.M{"home_time": t.UTC().Format(time.Kitchen)}).All(&users)

		for _, user := range users {
			log.Info(fmt.Sprintf("Sending home trips for the user %d", user.ID))
			trips := app.Sl.GetHomeTrips(user)

			for _, trip := range trips.Trip {
				app.Bot.Send(
					&tb.Chat{ID: user.ChatID}, app.Sl.GetMessageForTrip(trip), tb.ModeMarkdown,
				)
			}
		}

		c.Find(bson.M{"work_time": t.UTC().Format(time.Kitchen)}).All(&users)

		for _, user := range users {
			log.Info(fmt.Sprintf("Sending work trips for the user %d", user.ID))
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
	logLevel := flag.String("log_level", "info", "Log level")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	if err != nil {
		log.Fatal("Fatal error: ", err.Error())
	}
	log.SetLevel(level)

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

	log.Info("Starting stockholm commute bot")
	bot.Start()
}
