package main

import "ayu/server"
import "ayu/server/storage/datastore"
import "appengine"
import "net/http"

func get_datastore(r *http.Request) server.SaveLoader {
	return &datastore.DataStore{appengine.NewContext(r)}
}

func init() {
	server.Setup("", 55, get_datastore)
}
