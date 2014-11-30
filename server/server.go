package server

import "ayu"
import "container/list"
import "crypto/rand"
import "encoding/json"
import "encoding/hex"
import "io/ioutil"
import "log"
import "net/http"
import "os"
import "strconv"
import "sync"
import "time"

var poll_delay time.Duration
var database SaveLoader

type Saver interface {
	Save(kind string, key, value []byte) error
}

type Loader interface {
	Load(kind string, key []byte) ([]byte, error)
}

type SaveLoader interface {
	Saver
	Loader
}

type game struct {
	State    *ayu.State
	TimeUsed [2]time.Duration
	LastTime time.Time
	Keys     [2]string
	waiting  *list.List
	mutex    sync.Mutex // must be held while accessing fields above
}

func (g *game) version() int { return len(g.State.History) }

type Client struct {
	output chan<- string
}

var games = make(map[string]*game)
var games_mutex sync.Mutex // must be held while accessing games

func getGame(id string) (res *game) {
	games_mutex.Lock()
	res = games[id]
	games_mutex.Unlock()
	if res == nil && database != nil {
		// Game not in memory. Try to read it from database instead.
		if encoded, err := database.Load("Game", []byte(id)); err == nil {
			var game game
			if err := json.Unmarshal([]byte(encoded), &game); err != nil {
				log.Printf("Could not unmarshal game %s: %s. (Encoded: '%s')",
					id, err, encoded)
			} else {
				game.waiting = list.New()
				res = &game
				games_mutex.Lock()
				if games[id] != nil {
					// Some other thread already loaded the game. Use it instead.
					res = games[id]
				} else {
					games[id] = res
				}
				games_mutex.Unlock()
			}
		} else {
			log.Printf("Failed to load game %s: %s", id, err)
		}
	}
	return
}

func writeJsonResponse(w http.ResponseWriter, obj interface{}) {
	if text, err := json.Marshal(obj); err != nil {
		log.Fatalln(err)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(text)
	}
}

func handlePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	log.Print("GET /poll")

	game := getGame(r.FormValue("game"))
	if game == nil {
		http.Error(w, "Not Found", 404)
		return
	}

	// Wait for game to reach requested version.
	version, _ := strconv.Atoi(r.FormValue("version"))
	game.mutex.Lock()
	timeout_ch := time.After(poll_delay)
	update_ch := make(chan bool, 1)
	timed_out := false
	for game.version() < version && !timed_out {
		elem := game.waiting.PushBack(update_ch)
		game.mutex.Unlock()
		select {
		case <-update_ch:
			break
		case <-timeout_ch:
			timed_out = true
			break
		}
		game.mutex.Lock()
		game.waiting.Remove(elem)
	}

	// Write response.
	w.Header().Set("Cache-Control", "no-cache")
	if timed_out {
		w.WriteHeader(204) // HTTP 204 "No Content"
	} else {
		time_used := [2]float64{
			game.TimeUsed[0].Seconds(),
			game.TimeUsed[1].Seconds()}
		if !game.LastTime.IsZero() {
			time_used[game.State.Next()] +=
				time.Now().Sub(game.LastTime).Seconds()
		}
		writeJsonResponse(w, map[string]interface{}{
			"nextPlayer": game.State.NextPlayer(),
			"size":       len(game.State.Fields),
			"fields":     game.State.Fields,
			"history":    game.State.History,
			"timeUsed":   time_used})
	}
	game.mutex.Unlock()
}

func createRandomKey() string {
	var buf [10]byte
	if _, err := rand.Read(buf[:]); err != nil {
		log.Fatalln(err)
	}
	return hex.EncodeToString(buf[:])
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	log.Print("POST /create")
	var create struct {
		Size int
	}
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	} else if err := json.Unmarshal(body, &create); err != nil {
		http.Error(w, "Bad Request\n"+err.Error(), 400)
		return
	}
	if create.Size == 0 {
		create.Size = ayu.DefaultSize
	}
	if !ayu.IsValidSize(create.Size) {
		http.Error(w, "Bad Request\nInvalid board size.", 400)
		return
	}
	id := createRandomKey()
	games_mutex.Lock()
	defer games_mutex.Unlock()
	if games[id] != nil {
		// This should be extremely improbable!
		http.Error(w, "Internal Server Error", 500)
		return
	}
	games[id] = &game{ayu.CreateState(create.Size),
		[2]time.Duration{0, 0}, time.Time{},
		[2]string{createRandomKey(), createRandomKey()},
		list.New(), sync.Mutex{}}
	writeJsonResponse(w, map[string]interface{}{
		"game": id,
		"keys": games[id].Keys,
		"size": create.Size,
	})
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	log.Print("POST /update")

	var update struct {
		Game    string
		Version int
		Key     string
		Move    ayu.Move
	}
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	} else if err := json.Unmarshal(body, &update); err != nil {
		http.Error(w, "Bad Request\n"+err.Error(), 400)
		return
	}
	game := getGame(update.Game)
	if game == nil {
		http.Error(w, "Not Found", 404)
		return
	}
	game.mutex.Lock()
	defer game.mutex.Unlock()
	if update.Version != game.version() {
		http.Error(w, "Wrong Version", 409)
		return
	}
	player := game.State.Next()
	if update.Key != game.Keys[player] {
		http.Error(w, "Forbidden", 403)
		return
	}
	if !game.State.Execute(update.Move) {
		http.Error(w, "Illegal move", 403)
		return
	}

	// Update time used by last player.
	now := time.Now()
	if !game.LastTime.IsZero() {
		game.TimeUsed[player] += now.Sub(game.LastTime)
	}
	game.LastTime = now

	if database != nil {
		if encoded, err := json.Marshal(game); err != nil {
			log.Fatalln(err)
		} else if err := database.Save("Game", []byte(update.Game), encoded); err != nil {
			log.Printf("Failed to save game %s: %s", update.Game, err)
		}
	}

	// Notify goroutines waiting for updates.
	for {
		elem := game.waiting.Front()
		if elem == nil {
			break
		}
		game.waiting.Remove(elem).(chan bool) <- true
	}
}

func Setup(static_data_dir string, poll_delay_seconds int, db SaveLoader) {
	http.HandleFunc("/poll", handlePoll)
	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/update", handleUpdate)
	if static_data_dir != "" {
		if info, err := os.Stat(static_data_dir); err != nil {
			log.Fatalln(err)
		} else if !info.IsDir() {
			log.Fatalln(static_data_dir, "is not a directory")
		}
		http.Handle("/", http.FileServer(http.Dir(static_data_dir)))
	}
	poll_delay = time.Duration(poll_delay_seconds) * time.Second
	database = db
}
