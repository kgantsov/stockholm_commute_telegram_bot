package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
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

var transportTypesMap = map[string]string{
	"BUS": "B",
	"MET": "T",
	"TRM": "L",
	"TRN": "J",
	"SHP": "S",
}

type slClient struct {
	http        *http.Client
	userTextMap map[int]UserPoints
	mutex       sync.RWMutex
}

func NewSLClient() *slClient {
	cl := new(slClient)

	cl.http = &http.Client{
		Timeout: time.Second * 10,
	}

	cl.userTextMap = make(map[int]UserPoints)
	cl.mutex = sync.RWMutex{}

	return cl
}

func (cl *slClient) getTravelHomeURL(u UserPoints) string {
	return fmt.Sprintf(
		"http://api.sl.se/api2/TravelplannerV3/trip.json?key=%s&originID=%s&destID=%s&lang=en",
		os.Getenv("SL_PLANNING_API_KEY"),
		u.WorkID,
		u.HomeID,
	)
}

func (cl *slClient) getTravelWorkURL(u UserPoints) string {
	return fmt.Sprintf(
		"http://api.sl.se/api2/TravelplannerV3/trip.json?key=%s&originID=%s&destID=%s&lang=en",
		os.Getenv("SL_PLANNING_API_KEY"),
		u.HomeID,
		u.WorkID,
	)
}

func (cl *slClient) GetLookupStationURL(query string) string {
	return fmt.Sprintf(
		"http://api.sl.se/api2/typeahead.json?key=%s&SearchString=%s&StationOnly=True&MaxResults=6&lang=en",
		os.Getenv("SL_LOOKUP_API_KEY"),
		url.QueryEscape(query),
	)
}

func (cl *slClient) GetMessageForTrip(trip Trip) string {
	var items []string
	var tripDuration time.Duration
	startTripTime, _ := time.Parse("15:04:05", trip.LegList.Leg[0].Origin.Time)
	endTripTime, _ := time.Parse(
		"15:04:05", trip.LegList.Leg[len(trip.LegList.Leg)-1].Destination.Time,
	)
	tripDuration = endTripTime.Sub(startTripTime)

	transportType := transportTypesMap[trip.LegList.Leg[0].Product.CatOutS]

	items = append(
		items,
		fmt.Sprintf(
			"*%s%s* %s (*%s*)",
			transportType,
			trip.LegList.Leg[0].Product.Line,
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

		transportType = transportTypesMap[leg.Product.CatOutS]

		items = append(
			items,
			fmt.Sprintf(
				"*%s%s* %s (*%s* | *%s*)%s",
				transportType,
				leg.Product.Line,
				leg.Destination.Name,
				destinationTime.Format("15:04"),
				duration,
				warning,
			),
		)
	}
	return fmt.Sprintf("*Trip:* %s \n *Duration:* %s", strings.Join(items, " *=>* "), tripDuration)
}

func (cl *slClient) GetHomeTrips(u UserPoints) *TripsResult {
	response, _ := cl.http.Get(cl.getTravelHomeURL(u))

	var trips TripsResult
	err := json.NewDecoder(response.Body).Decode(&trips)
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return new(TripsResult)
	}

	return &trips
}

func (cl *slClient) GetWorkTrips(u UserPoints) *TripsResult {
	response, _ := cl.http.Get(cl.getTravelWorkURL(u))

	var trips TripsResult
	err := json.NewDecoder(response.Body).Decode(&trips)
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return new(TripsResult)
	}

	return &trips
}

func (cl *slClient) GetStationsByName(name string) *LookupResult {
	response, _ := cl.http.Get(cl.GetLookupStationURL(name))

	var lookup LookupResult
	err := json.NewDecoder(response.Body).Decode(&lookup)
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return new(LookupResult)
	}

	return &lookup
}
