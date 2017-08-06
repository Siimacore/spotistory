package spoticlient

import (
	"github.com/zmb3/spotify"
	"os"
	"golang.org/x/oauth2"
	"fmt"
	"encoding/json"
	"net/http"
	"log"
	"time"
	"io/ioutil"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate, spotify.ScopeUserReadRecentlyPlayed)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)


type Client struct {
	spotifyClient *spotify.Client
	db *DbHandler
	weeks map[int]string
}

const TOKEN_PATH = ".token"

func loadFromFile(filename string) *oauth2.Token{
	file, _ := os.Open(filename)
	var token oauth2.Token
	bytes, _ := ioutil.ReadAll(file)
	json.Unmarshal(bytes, &token)
	return &token
}

func authenticateFromUrl() *spotify.Client{
	server := &http.Server{Addr:":8080", Handler:nil}
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go server.ListenAndServe()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch
	defer close(ch)
	defer server.Shutdown(nil)

	token, _ := client.Token()

	err := saveToken(token)

	if err != nil {
		panic(err)
	}

	return client
}

func saveToken(token *oauth2.Token) error {
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(TOKEN_PATH, os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	_, err = file.Write(tokenBytes)

	return err
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}


func NewClient() *Client {
	db := NewDB("database.sqlite")
	db.Init()
	if _, err := os.Stat(TOKEN_PATH); !os.IsNotExist(err) {
		token := loadFromFile(TOKEN_PATH)
		client := auth.NewClient(token)
		return &Client{spotifyClient: &client, db: db, weeks: make(map[int]string)}
	}
	return &Client{spotifyClient: authenticateFromUrl(), db:db, weeks: make(map[int]string)}
}

func (c *Client) ShowPlaylists(){
	fmt.Println(c.spotifyClient.CurrentUser())
}

func (c *Client) CreateWeekPlaylist() *spotify.FullPlaylist{
	_, week := time.Now().ISOWeek()
	user, err := c.spotifyClient.CurrentUser()

	if err != nil {
		panic(err)
	}

	playlist, err := c.spotifyClient.CreatePlaylistForUser(user.ID, fmt.Sprintf("W%d", week), false)

	c.db.AddPlaylist(week, playlist.ID.String())

	c.weeks[week] = playlist.ID.String()

	if err != nil {
		panic(err)
	}
	return playlist
}

func (c *Client) getLastPlayed() []Track {
	lastPlayed, err := c.spotifyClient.PlayerRecentlyPlayedOpt(&spotify.RecentlyPlayedOptions{Limit: 50})

	if err != nil {
		panic(err)
	}
	slice := ([]Track)(nil)
	for track := range lastPlayed {
		tmp := lastPlayed[track]
		track := NewTrack(&tmp)
		slice = append(slice, *track)
	}
	return slice
}

func (c *Client) AddTracksToPlaylist(){
	tracks := c.getLastPlayed()
	for track := range tracks {
		tmp := tracks[track]
		c.db.AddTrack(&tmp)
	}
	tracksToAdd := c.db.GetTracksToAdd()
	toAdd := make(map[int][]spotify.ID)
	for track := range tracksToAdd {
		tmp := tracksToAdd[track]
		toAdd[tmp.Week] = append(toAdd[tmp.Week], spotify.ID(tmp.Id))
	}
	user, _ := c.spotifyClient.CurrentUser()
	for key, value := range toAdd {
		playlistId := spotify.ID(c.weeks[key])
		c.spotifyClient.AddTracksToPlaylist(user.ID, playlistId, value...)
	}
	c.db.SetAdded()
}