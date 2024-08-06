package strava

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gleich/lumber/v2"
	"github.com/minio/minio-go/v7"
)

type stravaActivity struct {
	Name      string    `json:"name"`
	SportType string    `json:"sport_type"`
	StartDate time.Time `json:"start_date"`
	Timezone  string    `json:"timezone"`
	Map       struct {
		SummaryPolyline string `json:"summary_polyline"`
	} `json:"map"`
	Trainer            bool    `json:"trainer"`
	Commute            bool    `json:"commute"`
	Private            bool    `json:"private"`
	AverageSpeed       float32 `json:"average_speed"`
	MaxSpeed           float32 `json:"max_speed"`
	AverageTemp        int32   `json:"average_temp,omitempty"`
	AverageCadence     float32 `json:"average_cadence,omitempty"`
	AverageWatts       float32 `json:"average_watts,omitempty"`
	DeviceWatts        bool    `json:"device_watts,omitempty"`
	AverageHeartrate   float32 `json:"average_heartrate,omitempty"`
	TotalElevationGain float32 `json:"total_elevation_gain"`
	MovingTime         uint32  `json:"moving_time"`
	SufferScore        float32 `json:"suffer_score,omitempty"`
	PrCount            uint32  `json:"pr_count"`
	Distance           float32 `json:"distance"`
	ID                 uint64  `json:"id"`
}

type activityStream struct {
	Data         []int  `json:"data"`
	SeriesType   string `json:"series_type"`
	OriginalSize int    `json:"original_size"`
	Resolution   string `json:"resolution"`
}

type activity struct {
	Name               string    `json:"name"`
	SportType          string    `json:"sport_type"`
	StartDate          time.Time `json:"start_date"`
	Timezone           string    `json:"timezone"`
	MapBlurImage       *string   `json:"map_blur_image"`
	MapImageURL        *string   `json:"map_image_url"`
	HasMap             bool      `json:"has_map"`
	TotalElevationGain float32   `json:"total_elevation_gain"`
	MovingTime         uint32    `json:"moving_time"`
	Distance           float32   `json:"distance"`
	ID                 uint64    `json:"id"`
	AverageHeartrate   float32   `json:"average_heartrate"`
	HeartrateData      []int     `json:"heartrate_data"`
}

func fetchActivities(minioClient minio.Client, tokens tokens) []activity {
	req, err := http.NewRequest("GET", "https://www.strava.com/api/v3/athlete/activities", nil)
	if err != nil {
		lumber.Error(err, "creating request failed")
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Access)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lumber.Error(err, "sending request for Strava activities failed")
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body failed")
		return nil
	}

	var stravaActivities []stravaActivity
	err = json.Unmarshal(body, &stravaActivities)
	if err != nil {
		lumber.Error(err, "failed to parse json")
		lumber.Debug(string(body))
		return nil
	}

	var activities []activity
	for _, stravaActivity := range stravaActivities {
		if len(activities) >= 3 {
			break
		}
		if stravaActivity.Private {
			continue
		}
		a := activity{
			Name:               stravaActivity.Name,
			SportType:          stravaActivity.SportType,
			StartDate:          stravaActivity.StartDate,
			Timezone:           stravaActivity.Timezone,
			TotalElevationGain: stravaActivity.TotalElevationGain,
			MovingTime:         stravaActivity.MovingTime,
			Distance:           stravaActivity.Distance,
			ID:                 stravaActivity.ID,
			AverageHeartrate:   stravaActivity.AverageHeartrate,
			HasMap:             stravaActivity.Map.SummaryPolyline != "",
			HeartrateData:      fetchHeartrate(stravaActivity.ID, tokens),
		}
		if a.HasMap {
			mapData := fetchMap(stravaActivity.Map.SummaryPolyline)
			uploadMap(minioClient, stravaActivity.ID, mapData)
			a.MapBlurImage = mapBlurData(mapData)
			imgurl := fmt.Sprintf(
				"https://minio-api.dev.mattglei.ch/mapbox-maps/%d.png",
				a.ID,
			)
			a.MapImageURL = &imgurl
		}
		activities = append(activities, a)
	}
	removeOldMaps(minioClient, stravaActivities)

	return activities
}

func fetchHeartrate(id uint64, tokens tokens) []int {
	params := url.Values{
		"key_by_type": {"true"},
		"keys":        {"heartrate"},
		"resolution":  {"medium"},
	}
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://www.strava.com/api/v3/activities/%d/streams?"+params.Encode(), id),
		nil,
	)
	if err != nil {
		lumber.Error(err, "creating request failed")
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Access)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lumber.Error(err, "Failed to send request for HR data for", id)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body failed")
		return nil
	}

	var stream struct{ Heartrate activityStream }
	err = json.Unmarshal(body, &stream)
	if err != nil {
		lumber.Error(err, "failed to parse json")
		lumber.Debug(string(body))
		return nil
	}

	return stream.Heartrate.Data
}