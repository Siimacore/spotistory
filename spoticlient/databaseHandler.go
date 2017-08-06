package spoticlient

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
)


type DbHandler struct {
	database *sql.DB
}

func NewDB(filename string) *DbHandler{
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		panic(err)
	}

	return &DbHandler{database: db}
}


func (d *DbHandler) Init(){
	history, err := ioutil.ReadFile("./scripts/create_history.sql")
	if err != nil {
		panic(err)
	}
	_, err = d.database.Exec(string(history))

	if err != nil {
		panic(err)
	}

	playlist, err := ioutil.ReadFile("./scripts/create_playlist.sql")
	if err != nil {
		panic(err)
	}
	_, err = d.database.Exec(string(playlist))

	if err != nil {
		panic(err)
	}
}

func (d *DbHandler) AddPlaylist(week int, playlist string){
	stmt, err := d.database.Prepare("INSERT INTO playlists VALUES(?,?)")
	if err != nil {
		panic(err)
	}
	stmt.Exec(week, playlist)
}

func (d *DbHandler) AddTrack(track *Track){
	stmt, err := d.database.Prepare("INSERT INTO history VALUES(?,?,?,?,?)")
	if err != nil {
		panic(err)
	}
	stmt.Exec(track.Id, track.Uri, track.PlayedAt, track.ToAdd, track.Week)
}

func (d *DbHandler) GetTracksToAdd() []Track{
	rows, err := d.database.Query("SELECT * FROM history WHERE to_add=1 ORDER BY played_at ASC")
	if err != nil {
		panic(err)
	}
	var id string
	var uri string
	var playedAt int64
	var toAdd bool
	var week int
	tracks := ([]Track)(nil)
	for rows.Next() {
		err = rows.Scan(&id, &uri, &playedAt, &toAdd, &week)
		if err != nil {
			panic(err)
		}
		tracks = append(tracks, Track{Id:id, Uri:uri, PlayedAt:playedAt, ToAdd:toAdd, Week:week})
	}

	return tracks
}

func (d *DbHandler) SetAdded(){
	d.database.Exec("UPDATE history set to_add=0")
}