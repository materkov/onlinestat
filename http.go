package onlinestat

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const pageTemplate = `
<!doctype html>
<html>
<style>
body {
	margin: 0;
	padding: 0;
}
.DayRow {
	height: 3px;
	display: inline-block;
}
.Status {
	display: inline-block;
	width: 1px;
	height: 2px;
	background: #f0f0f0;
}
.Status:hover {
	opacity: 0.7;
}
.Online {
	background: #008000;
}
.OnlineMobile {
	background: #56db56;
}
#time-pointer {
	position: fixed;
	width: 1px;
	height: 100%;
	top: 0;
	bottom: 0;
	background: #aaa;
}
#time-value {
	position: fixed;
	top: 0px;
	padding: 2px 3px;
	font-size: 12px;
	background: #e5e5e5;
	border-radius: 2px;
}
</style>
<head></head>
<body>
<script>
document.querySelector('body').addEventListener('mousemove', (e) => {
	let offset = e.clientX;
	if (offset < 0) {
		offset = 0;
	}
	if (offset > 24*60) {
		offset = 24*60;
	}

	document.querySelector("#time-pointer").style.left = offset + "px";
	document.querySelector("#time-value").style.left = String(offset + 2) + "px";

	let timeMinutes = String(Math.floor(offset / 60)).padStart(2, 0);
	let timeSeconds = String(offset % 60).padStart(2, 0);

	document.querySelector("#time-value").innerText = timeMinutes + ":" + timeSeconds;
});
</script>
{content}
<div id="time-pointer"></div>
<div id="time-value">12:55</div>
</body>
</html>
`

var RedisClient *redis.Client

func getLastOnlines() (map[int]int, error) {
	nowDate := time.Now().UTC()

	allOnlines := make(map[int]int, 0)

	for i := 0; i < 15; i++ {
		date := nowDate.Add(time.Hour * 24 * time.Duration(i) * -1)
		redisKey := fmt.Sprintf("online_info:%d:%d:%d", date.Year(), date.Month(), date.Day())
		values, err := RedisClient.HGetAll(context.Background(), redisKey).Result()
		if err != nil {
			return nil, fmt.Errorf("error getting data from redis: %w", err)
		}

		for tsStr, statusStr := range values {
			ts, _ := strconv.Atoi(tsStr)
			status, _ := strconv.Atoi(statusStr)

			allOnlines[ts] = status
		}
	}

	return allOnlines, nil
}

func ServeHTTP() {
	tzLocation, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalf("[ERROR] Failed to load TZ: %s", err)
	}

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		statuses, err := getLastOnlines()
		if err != nil {
			log.Printf("[ERROR] DB error: %s", err)
			fmt.Fprint(w, "db error")
			return
		}

		rows := map[string][]int{}

		for ts, status := range statuses {
			date := time.Unix(int64(ts), 0).In(tzLocation)
			row := date.Format("2006-01-02")
			if _, ok := rows[row]; !ok {
				rows[row] = make([]int, 60*24)
			}

			rows[row][date.Hour()*60+date.Minute()] = status
		}

		rowKeys := make([]string, len(rows))
		idx := 0
		for key := range rows {
			rowKeys[idx] = key
			idx++
		}
		sort.Strings(rowKeys)

		allRows := ""
		for _, row := range rowKeys {
			statusesHTML := ""
			for _, status := range rows[row] {
				if status == StatusOnline {
					statusesHTML += "<div class=\"Status Online\"></div>"
				} else if status == StatusOnlineMobile {
					statusesHTML += "<div class=\"Status OnlineMobile\"></div>"
				} else {
					statusesHTML += "<div class=\"Status Offline\"></div>"
				}
			}

			allRows = "<div>" + row + "</div><div class=\"DayRow\">" + statusesHTML + "</div>" + allRows
		}

		page := pageTemplate
		page = strings.Replace(page, "{content}", allRows, 1)
		fmt.Fprint(w, page)
	})
	http.ListenAndServe("127.0.0.1:8008", nil)
}
