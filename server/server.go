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

type game struct {
	state    *ayu.State
	timeUsed [2]time.Duration
	lastTime time.Time
	keys     [2]string
	waiting  *list.List
	mutex    sync.Mutex // must be held while accessing fields above
}

func (g *game) version() int { return len(g.state.History) }

type Client struct {
	output chan<- string
}

var games = make(map[string]*game)
var games_mutex sync.Mutex // must be held while accessing games

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

	games_mutex.Lock()
	game := games[r.FormValue("game")]
	games_mutex.Unlock()
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
			game.timeUsed[0].Seconds(),
			game.timeUsed[1].Seconds()}
		if !game.lastTime.IsZero() {
			time_used[game.state.Next()] +=
				time.Now().Sub(game.lastTime).Seconds()
		}
		writeJsonResponse(w, map[string]interface{}{
			"nextPlayer": game.state.NextPlayer(),
			"size":       ayu.S,
			"fields":     game.state.Fields,
			"history":    game.state.History,
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
	id := createRandomKey()
	games_mutex.Lock()
	defer games_mutex.Unlock()
	if games[id] != nil {
		// This should be extremely improbable!
		http.Error(w, "Internal Server Error", 500)
		return
	}
	games[id] = &game{ayu.CreateState(),
		[2]time.Duration{0, 0}, time.Time{},
		[2]string{createRandomKey(), createRandomKey()},
		list.New(), sync.Mutex{}}
	writeJsonResponse(w, map[string]interface{}{
		"game": id,
		"keys": games[id].keys,
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
	} else if json.Unmarshal(body, &update) != nil {
		http.Error(w, "Bad Request\n"+err.Error(), 400)
		return
	}
	games_mutex.Lock()
	game := games[update.Game]
	games_mutex.Unlock()
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
	player := game.state.Next()
	if update.Key != game.keys[player] {
		http.Error(w, "Forbidden", 403)
		return
	}
	if !game.state.Execute(update.Move) {
		http.Error(w, "Illegal move", 403)
		return
	}

	// Update time used by last player.
	now := time.Now()
	if !game.lastTime.IsZero() {
		game.timeUsed[player] += now.Sub(game.lastTime)
	}
	game.lastTime = now

	// Notify goroutines waiting for updates.
	for {
		elem := game.waiting.Front()
		if elem == nil {
			break
		}
		game.waiting.Remove(elem).(chan bool) <- true
	}
}

func Setup(static_data_dir string, poll_delay_seconds int) {
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
}
