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

type playerAchievementsResponse struct {
	PlayerStats struct {
		Achievements *[]struct {
			ApiName    string `json:"apiname"`
			Achieved   int    `json:"achieved"`
			UnlockTime *int64 `json:"unlocktime"`
		}
	} `json:"playerStats"`
}

type schemaGameResponse struct {
	Game struct {
		GameStats struct {
			Achievements []struct {
				DisplayName string  `json:"displayName"`
				Icon        string  `json:"icon"`
				Description *string `json:"description"`
				Name        string  `json:"name"`
			} `json:"achievements"`
		} `json:"availableGameStats"`
	} `json:"game"`
}

type Achievement struct {
	ApiName     string     `json:"api_name"`
	Achieved    bool       `json:"achieved"`
	Icon        string     `json:"icon"`
	DisplayName string     `json:"display_name"`
	Description *string    `json:"description"`
	UnlockTime  *time.Time `json:"unlock_time"`
}

func FetchGameAchievements(appID int32) *[]Achievement {
	params := url.Values{
		"key":     {secrets.SECRETS.SteamKey},
		"steamid": {secrets.SECRETS.SteamID},
		"appid":   {fmt.Sprint(appID)},
		"format":  {"json"},
	}
	resp, err := http.Get("https://api.steampowered.com/ISteamUserStats/GetPlayerAchievements/v0001?" + params.Encode())
	if err != nil {
		lumber.Error(err, "sending request for player achievements from", appID, "failed")
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body for player achievements from", appID, "failed")
		return nil
	}
	if string(body) == `{"playerstats":{"error":"Requested app has no stats","success":false}}` {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		lumber.ErrorMsg(resp.StatusCode, "when trying to get player achievements for", appID, string(body))
		return nil
	}

	var playerAchievements playerAchievementsResponse
	err = json.Unmarshal(body, &playerAchievements)
	if err != nil {
		lumber.Error(err, "failed to parse json for player achievements for", appID)
		lumber.Debug("body:", string(body))
		return nil
	}

	if playerAchievements.PlayerStats.Achievements == nil {
		return nil
	}

	params = url.Values{
		"key":    {secrets.SECRETS.SteamKey},
		"appid":  {fmt.Sprint(appID)},
		"format": {"json"},
	}
	resp, err = http.Get("https://api.steampowered.com/ISteamUserStats/GetSchemaForGame/v2?" + params.Encode())
	if err != nil {
		lumber.Error(err, "sending request for owned games failed")
		return nil
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body for game schema failed for", appID)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		lumber.ErrorMsg(resp.StatusCode, "when trying to get player achievements for", appID, string(body))
		return nil
	}

	var gameSchema schemaGameResponse
	err = json.Unmarshal(body, &gameSchema)
	if err != nil {
		lumber.Error(err, "failed to parse json for game schema for", appID)
		lumber.Debug("body:", string(body))
		return nil
	}

	var achievements []Achievement
	for _, playerAchievement := range *playerAchievements.PlayerStats.Achievements {
		for _, schemaAchievement := range gameSchema.Game.GameStats.Achievements {
			if playerAchievement.ApiName == schemaAchievement.Name {
				var unlockTime time.Time
				if playerAchievement.UnlockTime != nil && *playerAchievement.UnlockTime != 0 {
					unlockTime = time.Unix(*playerAchievement.UnlockTime, 0)
				}
				achievements = append(achievements, Achievement{
					ApiName:     playerAchievement.ApiName,
					Achieved:    playerAchievement.Achieved == 1,
					Icon:        schemaAchievement.Icon,
					DisplayName: schemaAchievement.DisplayName,
					Description: schemaAchievement.Description,
					UnlockTime:  &unlockTime,
				})
			}
		}
	}

	sort.Slice(achievements, func(i, j int) bool {
		if achievements[i].UnlockTime == nil && achievements[j].UnlockTime == nil {
			return false
		}
		if achievements[i].UnlockTime == nil {
			return false
		}
		if achievements[j].UnlockTime == nil {
			return true
		}
		return achievements[i].UnlockTime.After(*achievements[j].UnlockTime)
	})

	if len(achievements) > 20 {
		achievements = achievements[:20]
	}

	return &achievements
}
