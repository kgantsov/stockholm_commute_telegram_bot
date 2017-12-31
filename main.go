package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
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

type Station struct {
	Name   string `json:"Name"`
	SiteID string `json:"SiteId"`
	Type   string `json:"Type"`
	X      string `json:"X"`
	Y      string `json:"Y"`
}

type LookupResult struct {
	ExecutionTime int         `json:"ExecutionTime"`
	Message       interface{} `json:"Message"`
	ResponseData  []Station   `json:"ResponseData"`
	StatusCode    int         `json:"StatusCode"`
}

type UserPoints struct {
	HomeName string
	HomeID   string
	WorkName string
	WorkID   string
}

var userTextMap map[int]UserPoints

func getTravelHomeURL(u UserPoints) string {
	return fmt.Sprintf(
		"http://api.sl.se/api2/TravelplannerV3/trip.json?key=%s&originID=%s&destID=%s&lang=en",
		os.Getenv("SL_PLANNING_API_KEY"),
		u.WorkID,
		u.HomeID,
	)
}

func getTravelWorkURL(u UserPoints) string {
	return fmt.Sprintf(
		"http://api.sl.se/api2/TravelplannerV3/trip.json?key=%s&originID=%s&destID=%s&lang=en",
		os.Getenv("SL_PLANNING_API_KEY"),
		u.HomeID,
		u.WorkID,
	)
}

func getLookupStationURL(query string) string {
	return fmt.Sprintf(
		"http://api.sl.se/api2/typeahead.json?key=%s&SearchString=%s&StationOnly=True&MaxResults=6&lang=en",
		os.Getenv("SL_LOOKUP_API_KEY"),
		url.QueryEscape(query),
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
	userTextMap = make(map[int]UserPoints)
	var mutex = &sync.Mutex{}

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
		mutex.Lock()
		u, ok := userTextMap[m.Sender.ID]
		mutex.Unlock()

		if !ok {
			b.Send(m.Sender, "Please setup home and work locations", tb.ModeMarkdown)
		}

		response, _ := netClient.Get(getTravelHomeURL(u))

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
		mutex.Lock()
		u, ok := userTextMap[m.Sender.ID]
		mutex.Unlock()

		if !ok {
			b.Send(m.Sender, "Please setup home and work locations", tb.ModeMarkdown)
		}

		response, _ := netClient.Get(getTravelWorkURL(u))

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

	b.Handle("/set_home", func(m *tb.Message) {
		response, _ := netClient.Get(getLookupStationURL(m.Payload))

		var lookup LookupResult
		err := json.NewDecoder(response.Body).Decode(&lookup)
		if err != nil {
			fmt.Println("ERROR", err.Error())
			return
		}

		var replyKeys [][]tb.ReplyButton

		for _, station := range lookup.ResponseData {
			replyBtn := tb.ReplyButton{Text: station.Name}

			b.Handle(&replyBtn, func(st Station) func(m *tb.Message) {
				return func(m *tb.Message) {
					mutex.Lock()
					if _, ok := userTextMap[m.Sender.ID]; !ok {
						userTextMap[m.Sender.ID] = UserPoints{}
					}

					userTextMap[m.Sender.ID] = UserPoints{
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
		response, _ := netClient.Get(getLookupStationURL(m.Payload))

		var lookup LookupResult
		err := json.NewDecoder(response.Body).Decode(&lookup)
		if err != nil {
			fmt.Println("ERROR", err.Error())
			return
		}

		var replyKeys [][]tb.ReplyButton

		for _, station := range lookup.ResponseData {
			replyBtn := tb.ReplyButton{Text: station.Name}

			b.Handle(&replyBtn, func(st Station) func(m *tb.Message) {
				return func(m *tb.Message) {
					mutex.Lock()
					if _, ok := userTextMap[m.Sender.ID]; !ok {
						userTextMap[m.Sender.ID] = UserPoints{}
					}

					userTextMap[m.Sender.ID] = UserPoints{
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
