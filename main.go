package onlinestat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	StatusOffline      = 0
	StatusOnline       = 1
	StatusOnlineMobile = 2
)

func GetStatus(token string) (int, error) {
	resp, err := http.Get("http://api.vk.com/method/users.get?user_ids=9914880&v=5.130&fields=online_info&access_token=" + token)
	if err != nil {
		return 0, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading body: %w", err)
	}

	body := struct {
		Response []struct {
			OnlineInfo struct {
				IsOnline bool `json:"is_online"`
				IsMobile bool `json:"is_mobile"`
			} `json:"online_info"`
		} `json:"response"`
		Error json.RawMessage `json:"error"`
	}{}
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return 0, fmt.Errorf("error parsing json: %w", err)
	} else if body.Error != nil {
		return 0, fmt.Errorf("api error: %s", bodyBytes)
	}

	if len(body.Response) != 1 {
		return 0, fmt.Errorf("must be 1 user: %s", bodyBytes)
	}

	if body.Response[0].OnlineInfo.IsOnline {
		if body.Response[0].OnlineInfo.IsMobile {
			return StatusOnlineMobile, nil
		} else {
			return StatusOnline, nil
		}
	} else {
		return StatusOffline, nil
	}
}
