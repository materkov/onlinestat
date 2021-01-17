package onlinestat

import (
	"context"
	"fmt"
	"log"
	"time"
)

func FetchForever(vkToken string) {
	for {
		status, err := GetStatus(vkToken)
		if err != nil {
			log.Printf("[ERROR] Error getting status: %s", err)
			time.Sleep(time.Minute)
			continue
		}

		log.Printf("Status: %d", status)

		ts := time.Now().UTC()
		y, m, d := ts.Date()

		if status == StatusOffline {
			time.Sleep(time.Minute)
			continue
		}

		redisKey := fmt.Sprintf("online_info:%d:%d:%d", y, m, d)

		_, err = RedisClient.HSet(context.Background(), redisKey, ts.Unix(), status).Result()
		if err != nil {
			log.Printf("[ERROR] Error setting redis key: %s", err)
		}

		time.Sleep(time.Minute)
	}
}