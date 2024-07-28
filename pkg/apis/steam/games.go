package steam

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/gleich/lcp-v2/pkg/secrets"
	"github.com/gleich/lumber/v2"
)

type ownedGamesResponse struct {
	Response struct {
		Games []struct {
			Name            string `json:"name"`
			AppID           int32  `json:"appid"`
			ImgIconURL      string `json:"img_icon_url"`
			RTimeLastPlayed int64  `json:"rtime_last_played"`
			PlaytimeForever int32  `json:"playtime_forever"`
		} `json:"games"`
	} `json:"response"`
}

type Game struct {
	Name                string         `json:"name"`
	AppID               int32          `json:"app_id"`
	IconURL             string         `json:"icon_url"`
	RTimeLastPlayed     time.Time      `json:"rtime_last_played"`
	PlaytimeForever     int32          `json:"playtime_forever"`
	URL                 string         `json:"url"`
	HeaderURL           string         `json:"header_url"`
	LibraryURL          *string        `json:"library_url"`
	AchievementProgress *float32       `json:"achievement_progress"`
	Achievements        *[]Achievement `json:"achievements"`
}

func FetchRecentlyPlayedGames() []Game {
	params := url.Values{
		"key":             {secrets.SECRETS.SteamKey},
		"steamid":         {secrets.SECRETS.SteamID},
		"include_appinfo": {"true"},
		"format":          {"json"},
	}
	resp, err := http.Get("https://api.steampowered.com/IPlayerService/GetOwnedGames/v1?" + params.Encode())
	if err != nil {
		lumber.Error(err, "sending request for owned games failed")
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body for owned games failed")
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		lumber.ErrorMsg(resp.StatusCode, "when trying to get owned games", string(body))
		return nil
	}

	var ownedGames ownedGamesResponse
	err = json.Unmarshal(body, &ownedGames)
	if err != nil {
		lumber.Error(err, "failed to parse json for owned games")
		lumber.Debug("body:", string(body))
		return nil
	}

	sort.Slice(ownedGames.Response.Games, func(i, j int) bool {
		return ownedGames.Response.Games[i].RTimeLastPlayed > ownedGames.Response.Games[j].RTimeLastPlayed
	})
	ownedGames.Response.Games = ownedGames.Response.Games[:10]

	var games []Game
	for _, g := range ownedGames.Response.Games {
		libraryURL := fmt.Sprintf("https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/library_600x900.jpg", g.AppID)
		libraryImageResponse, err := http.Get(libraryURL)
		if err != nil {
			lumber.Error(err, "getting library image for", g.Name, "failed")
			return nil
		}
		defer libraryImageResponse.Body.Close()

		var libraryURLPtr *string
		if libraryImageResponse.StatusCode == http.StatusOK {
			libraryURLPtr = &libraryURL
		}

		achievementPercentage, achievements := FetchGameAchievements(g.AppID)

		games = append(games, Game{
			Name:                g.Name,
			AppID:               g.AppID,
			IconURL:             fmt.Sprintf("https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg", g.AppID, g.ImgIconURL),
			RTimeLastPlayed:     time.Unix(g.RTimeLastPlayed, 0),
			PlaytimeForever:     g.PlaytimeForever,
			URL:                 fmt.Sprintf("https://store.steampowered.com/app/%d/", g.AppID),
			HeaderURL:           fmt.Sprintf("https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/header.jpg", g.AppID),
			LibraryURL:          libraryURLPtr,
			AchievementProgress: achievementPercentage,
			Achievements:        achievements,
		})
	}

	return games
}
