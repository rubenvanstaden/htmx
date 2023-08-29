package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	repository := Repository{
		db: make(map[string]*Profile),
	}

	for i := 0; i < 22; i++ {

		p := &Profile{
			Name:   fmt.Sprintf("alice%d", i),
			Email:  fmt.Sprintf("alice%d@gmail.com", i),
			PubKey: fmt.Sprintf("npub%d", i),
		}

		repository.Store(p)
	}

	r := mux.NewRouter()

	handler := Handler{
		repository: repository,
	}

	r.HandleFunc("/", handler.IndexHandler)
	r.HandleFunc("/contact", handler.SearchProfile)
	r.HandleFunc("/contact/new", handler.NewProfile).Methods("GET")
	r.HandleFunc("/contact/new", handler.AddProfile).Methods("POST")
	r.HandleFunc("/contact/{pubkey:[a-zA-Z0-9]+}/email", handler.ParseEmail).Methods("GET")
	r.HandleFunc("/contact/{pubkey:[a-zA-Z0-9]+}", handler.ShowProfile).Methods("POST")
	r.HandleFunc("/contact/{pubkey:[a-zA-Z0-9]+}", handler.DeleteProfile).Methods("DELETE")
	r.HandleFunc("/contact/{pubkey:[a-zA-Z0-9]+}/edit", handler.EditProfile).Methods("GET")
	r.HandleFunc("/contact/{pubkey:[a-zA-Z0-9]+}/edit", handler.SaveProfile).Methods("POST")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalln("There's an error with the server", err)
	}
}
