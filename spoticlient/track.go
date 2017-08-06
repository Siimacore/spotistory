package spoticlient

import "github.com/zmb3/spotify"

type Track struct {
	Id string
	Uri string
	PlayedAt int64
	ToAdd bool
	Week int
}

func NewTrack(item *spotify.RecentlyPlayedItem) *Track{
	_, week := item.PlayedAt.ISOWeek()
	return &Track{Id:item.Track.ID.String(), Uri:string(item.Track.URI), PlayedAt:item.PlayedAt.Unix(), ToAdd:true, Week:week}
}