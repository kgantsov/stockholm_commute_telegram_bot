package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type Trip struct {
	LegList struct {
		Leg []struct {
			Destination struct {
				Date          string  `json:"date"`
				ExtID         string  `json:"extId"`
				HasMainMast   bool    `json:"hasMainMast"`
				ID            string  `json:"id"`
				Lat           float64 `json:"lat"`
				Lon           float64 `json:"lon"`
				MainMastExtID string  `json:"mainMastExtId"`
				MainMastID    string  `json:"mainMastId"`
				Name          string  `json:"name"`
				PrognosisType string  `json:"prognosisType"`
				Time          string  `json:"time"`
				Track         string  `json:"track"`
				Type          string  `json:"type"`
			} `json:"Destination"`
			JourneyDetailRef struct {
				Ref string `json:"ref"`
			} `json:"JourneyDetailRef"`
			JourneyStatus string `json:"JourneyStatus"`
			Messages      struct {
				Message []struct {
					Act      bool   `json:"act"`
					Category string `json:"category"`
					EDate    string `json:"eDate"`
					ETime    string `json:"eTime"`
					Head     string `json:"head"`
					ID       string `json:"id"`
					Priority int    `json:"priority"`
					Products int    `json:"products"`
					SDate    string `json:"sDate"`
					STime    string `json:"sTime"`
					Text     string `json:"text"`
				} `json:"Message"`
			} `json:"Messages"`
			Origin struct {
				Date          string  `json:"date"`
				ExtID         string  `json:"extId"`
				HasMainMast   bool    `json:"hasMainMast"`
				ID            string  `json:"id"`
				Lat           float64 `json:"lat"`
				Lon           float64 `json:"lon"`
				MainMastExtID string  `json:"mainMastExtId"`
				MainMastID    string  `json:"mainMastId"`
				Name          string  `json:"name"`
				PrognosisType string  `json:"prognosisType"`
				Time          string  `json:"time"`
				Track         string  `json:"track"`
				Type          string  `json:"type"`
			} `json:"Origin"`
			Product struct {
				Admin        string `json:"admin"`
				CatCode      string `json:"catCode"`
				CatIn        string `json:"catIn"`
				CatOut       string `json:"catOut"`
				CatOutL      string `json:"catOutL"`
				CatOutS      string `json:"catOutS"`
				Line         string `json:"line"`
				Name         string `json:"name"`
				Num          string `json:"num"`
				Operator     string `json:"operator"`
				OperatorCode string `json:"operatorCode"`
			} `json:"Product"`
			Category  string `json:"category"`
			Direction string `json:"direction"`
			Idx       string `json:"idx"`
			Name      string `json:"name"`
			Number    string `json:"number"`
			Reachable bool   `json:"reachable"`
			Type      string `json:"type"`
		} `json:"Leg"`
	} `json:"LegList"`
	ServiceDays []struct {
		PlanningPeriodBegin string `json:"planningPeriodBegin"`
		PlanningPeriodEnd   string `json:"planningPeriodEnd"`
		SDaysB              string `json:"sDaysB"`
		SDaysI              string `json:"sDaysI"`
		SDaysR              string `json:"sDaysR"`
	} `json:"ServiceDays"`
	TariffResult struct {
		FareSetItem []struct {
			Desc     string `json:"desc"`
			FareItem []struct {
				Cur   string `json:"cur"`
				Desc  string `json:"desc"`
				Name  string `json:"name"`
				Price int    `json:"price"`
			} `json:"fareItem"`
			Name string `json:"name"`
		} `json:"fareSetItem"`
	} `json:"TariffResult"`
	Checksum string `json:"checksum"`
	CtxRecon string `json:"ctxRecon"`
	Duration string `json:"duration"`
	Idx      int    `json:"idx"`
	TripID   string `json:"tripId"`
}

type TripsResult struct {
	Trip           []Trip `json:"Trip"`
	DialectVersion string `json:"dialectVersion"`
	RequestID      string `json:"requestId"`
	ScrB           string `json:"scrB"`
	ScrF           string `json:"scrF"`
	ServerVersion  string `json:"serverVersion"`
}

func getTravelHomeURL() string {
	return fmt.Sprintf(
		"http://api.sl.se/api2/TravelplannerV3/trip.json?key=%s&originID=9117&destID=9701&lang=en",
		os.Getenv("SL_PLANNING_API_KEY"),
	)
}

func getTravelWorkURL() string {
	return fmt.Sprintf(
		"http://api.sl.se/api2/TravelplannerV3/trip.json?key=%s&originID=9701&destID=9117&lang=en",
		os.Getenv("SL_PLANNING_API_KEY"),
	)
}

func getMessageForTrip(trip Trip) string {
	var items []string
	var tripDuration time.Duration
	startTripTime, _ := time.Parse("15:04:05", trip.LegList.Leg[0].Origin.Time)
	endTripTime, _ := time.Parse(
		"15:04:05", trip.LegList.Leg[len(trip.LegList.Leg)-1].Destination.Time,
	)
	tripDuration = endTripTime.Sub(startTripTime)

	items = append(
		items,
		fmt.Sprintf(
			"%s (*%s*)",
			trip.LegList.Leg[0].Origin.Name,
			startTripTime.Format("15:04"),
		),
	)

	for _, leg := range trip.LegList.Leg {
		originTime, _ := time.Parse("15:04:05", leg.Origin.Time)
		destinationTime, _ := time.Parse("15:04:05", leg.Destination.Time)
		duration := destinationTime.Sub(originTime)

		warning := ""
		if len(leg.Messages.Message) > 0 {
			warning = fmt.Sprintf(
				" *%s:* %s", leg.Messages.Message[0].Head, leg.Messages.Message[0].Text,
			)
		}

		items = append(
			items,
			fmt.Sprintf(
				"%s (*%s* | *%s*)%s",
				leg.Destination.Name,
				destinationTime.Format("15:04"),
				duration,
				warning,
			),
		)
	}
	return fmt.Sprintf("*Trip:* %s \n *Duration:* %s", strings.Join(items, " *=>* "), tripDuration)
}

func main() {
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Sender, "Hi this is a StockholmCommuteBot")
	})

	b.Handle("/home", func(m *tb.Message) {
		response, _ := netClient.Get(getTravelHomeURL())

		var trips TripsResult
		err := json.NewDecoder(response.Body).Decode(&trips)
		if err != nil {
			fmt.Println("ERROR", err.Error())
			return
		}

		for _, trip := range trips.Trip {
			b.Send(m.Sender, getMessageForTrip(trip), tb.ModeMarkdown)
		}
	})

	b.Handle("/work", func(m *tb.Message) {
		response, _ := netClient.Get(getTravelWorkURL())

		var trips TripsResult
		err := json.NewDecoder(response.Body).Decode(&trips)
		if err != nil {
			fmt.Println("ERROR", err.Error())
			return
		}

		for _, trip := range trips.Trip {
			b.Send(m.Sender, getMessageForTrip(trip), tb.ModeMarkdown)
		}
	})

	b.Start()
}
