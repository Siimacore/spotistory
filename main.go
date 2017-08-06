package main

import (
	"spotistory/spoticlient"
	"time"
	"fmt"
)

func main() {
	client := spoticlient.NewClient()
	ticker := time.NewTicker(time.Duration(30) * time.Minute)
	client.CreateWeekPlaylist()
	_, week := time.Now().ISOWeek()
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				_, nWeek := time.Now().ISOWeek()
				if week == nWeek{
					client.CreateWeekPlaylist()
				}
				client.AddTracksToPlaylist()
				fmt.Println("Tick")
			}
		}
	}()

	<-quit
}

