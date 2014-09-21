package main

import "ayu"
import "crypto/rand"
import "flag"
import "fmt"
import "encoding/json"
import "encoding/hex"
import "io/ioutil"
import "log"
import "net/http"
import "strconv"
import "sync"

var host = flag.String("host", "localhost", "Hostname to bind HTTP server on")
var port = flag.Int("port", 8027, "TCP port to bind HTTP server on")

type game struct {
	state   *ayu.State
	updated *sync.Cond
	keys    [2]string
}

func (g *game) version() int { return len(g.state.History) }

type Client struct {
	output chan<- string
}

var games = make(map[string]*game)
var games_mutex sync.Mutex

func writeJsonResponse(w http.ResponseWriter, obj interface{}) {
	if text, err := json.Marshal(obj); err != nil {
		log.Panic(err)
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

	version, _ := strconv.Atoi(r.FormValue("version"))

	game.updated.L.Lock()
	defer game.updated.L.Unlock()
	for game.version() < version {
		game.updated.Wait()
	}
	w.Header().Set("Cache-Control", "no-cache")
	writeJsonResponse(w, map[string]interface{}{
		"nextPlayer": game.state.NextPlayer(),
		"size":       ayu.S,
		"fields":     game.state.Fields,
		"history":    game.state.History,
	})
}

func createRandomKey() string {
	var buf [10]byte
	if _, err := rand.Read(buf[:]); err != nil {
		log.Panic(err)
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
		// This should be extremely inprobably!
		http.Error(w, "Internal Server Error", 500)
		return
	}
	games[id] = &game{ayu.CreateState(), sync.NewCond(&sync.Mutex{}),
		[2]string{createRandomKey(), createRandomKey()}}
	writeJsonResponse(w, map[string]interface{}{
		"game": id,
		"keys": games[id].keys,
	})
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	var update struct {
		Game string
		Version int
		Key string
		Move ayu.Move
	}

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	log.Print("POST /update")

	if body,err := ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	} else if json.Unmarshal(body, &update) != nil {
		http.Error(w, "Bad Request\n" + err.Error(), 400)
		return
	}

	games_mutex.Lock()
	game := games[update.Game]
	games_mutex.Unlock()
	if game == nil {
		http.Error(w, "Not Found", 404)
		return
	}
	game.updated.L.Lock()
	defer game.updated.L.Unlock()

	if update.Version != game.version() {
		http.Error(w, "Wrong Version", 409)
		return
	}

	if update.Key != game.keys[game.state.Next()] {
		http.Error(w, "Forbidden", 403)
		return
	}

	if !game.state.Execute(update.Move) {
		http.Error(w, "Illegal move", 403)
		return
	}
	game.updated.Broadcast()
}

func serveHttp() {
	http.HandleFunc("/poll", handlePoll)
	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/update", handleUpdate)
	http.Handle("/", http.FileServer(http.Dir("static")))
	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Binding to address %s.", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func main() {
	flag.Parse()
	serveHttp()
}
