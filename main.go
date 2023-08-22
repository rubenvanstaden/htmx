package main

import (
	"html/template"
	"log"
	"net/http"
)

type Data struct {
	Title       string
	Content     string
	Contacts    []Contact
	SearchQuery string
}

type Contact struct {
	Id    int
	First string
	Last  string
	Phone string
	Email string
}

type Client struct {
	db map[string]Contact
}

func main() {

	client := Client{
		db: make(map[string]Contact),
	}

	client.db["alice"] = Contact{
		Id:    0,
		First: "alice",
	}

	client.db["bob"] = Contact{
		Id:    1,
		First: "bob",
	}

	http.HandleFunc("/", client.indexHandler)
	http.HandleFunc("/contact", client.contact)
	http.HandleFunc("/update", client.updateHandler)

	http.ListenAndServe(":8080", nil)
}

func (s *Client) contact(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query().Get("q")

	data := Data{
		Title:       "HTMX with Go",
		SearchQuery: q,
	}

    log.Printf("%s", q)

	data.Contacts = append(data.Contacts, s.db["alice"])
	data.Contacts = append(data.Contacts, s.db["bob"])

	tmpl, err := template.ParseFiles("template/layout.html", "template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func (s *Client) indexHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/contact", http.StatusMovedPermanently)
}

func (s *Client) updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Updated Content using HTMX and Go!"))
}
