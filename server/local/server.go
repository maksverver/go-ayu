package main

import "ayu/server"
import "flag"
import "fmt"
import "log"
import "net/http"

var host = flag.String("host", "localhost", "Hostname to bind HTTP server on")
var port = flag.Int("port", 8027, "TCP port to bind HTTP server on")
var static_data_dir = flag.String("static_data_dir", "static", "Directory containing static files to serve")
var poll_delay = flag.Int("poll_delay", 55, "Maximum time to block on poll requests (in seconds)")

// TODO: local (disk-based) "database" implementation.

func main() {
	flag.Parse()
	if len(flag.Args()) > 0 {
		log.Fatalln("Extra command line arguments", flag.Args())
	}
	addr := fmt.Sprintf("%s:%d", *host, *port)
	server.Setup(*static_data_dir, *poll_delay, nil)
	log.Println("Binding to address:", addr)
	log.Fatalln(http.ListenAndServe(addr, nil))
}
