package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shipyard-run/tty/server"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	// "github.com/sevlyar/go-daemon"
)

func main() {
	// if cfg.Daemon {
	// 	ctx := &daemon.Context{
	// 		PidFileName: "/var/run/shipyard.pid",
	// 		PidFilePerm: 0644,
	// 		LogFileName: "/var/log/shipyard.log",
	// 		LogFilePerm: 0640,
	// 		WorkDir:     "/",
	// 		Umask:       027,
	// 		Args:        []string{"[shipyard]"},
	// 	}

	// 	d, err := ctx.Reborn()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if d != nil {
	// 		return
	// 	}
	// 	defer ctx.Release()
	// }

	r := mux.NewRouter()
	server.HandleTerminal(r.PathPrefix("/terminal").Subrouter())

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Origin", "Accept", "Token"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHandler := handlers.CORS(headersOk, originsOk, methodsOk)

	address := fmt.Sprintf("%s:%d", "0.0.0.0", 27950)
	log.Printf("Listening on http://%s", address)
	log.Fatal(http.ListenAndServe(address, corsHandler(r)))
}
