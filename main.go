package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/gorilla/mux"
)

type Data struct {
	Title       string
	Content     string
	Contacts    []*Contact
	SearchQuery string
}

type Contact struct {
	Id    int
	First string
	Last  string
	Phone string
	Email string
    Error string
}

// Assuming a function to save the contact; for this example, it just returns true.
func (c *Contact) Save() bool {
	// Your database saving logic goes here.
	return true // assuming it always succeeds for this example.
}

type Client struct {
	db map[int]*Contact
}

func (s *Client) Find(id int) *Contact {
	return s.db[id]
}

func (s *Client) Delete(id int) {
    delete(s.db, id)
}

func (s *Client) Update(id int, firstName, lastName, phone, email string) *Contact {

    c := &Contact{
        First: firstName,
        Last: lastName,
        Phone: phone,
        Email: email,
    }

    s.db[id] = c

	return c
}

func main() {

	client := Client{
		db: make(map[int]*Contact),
	}

	client.db[0] = &Contact{
		Id:    0,
		First: "alice",
        Email: "alice@gmail.com",
        Phone: "000",
        Error: "none",
	}

	client.db[1] = &Contact{
		Id:    1,
		First: "bob",
        Email: "bob@gmail.com",
        Phone: "001",
        Error: "none",
	}

	r := mux.NewRouter()

	r.HandleFunc("/", client.IndexHandler)
	r.HandleFunc("/contact", client.ContactHandler)
	r.HandleFunc("/contact/new", client.ContactsNewGetHandler).Methods("GET")
	r.HandleFunc("/contact/new", client.ContactsNewPostHandler).Methods("POST")
	r.HandleFunc("/contact/{contact_id:[0-9]+}", client.ContactViewHandler).Methods("POST")
	r.HandleFunc("/contact/{contact_id:[0-9]+}/edit", client.ContactEditHandler).Methods("GET")
    r.HandleFunc("/contact/{contact_id:[0-9]+}/email", client.ContactEmailGetHandler).Methods("GET")
	r.HandleFunc("/contact/{contact_id:[0-9]+}/edit", client.ContactEditPostHandler).Methods("POST")
    r.HandleFunc("/contact/{contact_id:[0-9]+}", client.ContactDeletePostHandler).Methods("DELETE")

    err := http.ListenAndServe(":8080", r)
    if err != nil {
        log.Fatalln("There's an error with the server", err)
    }
}

func (s *Client) ContactEmailGetHandler(w http.ResponseWriter, r *http.Request) {

    log.Println("Checking mail")

    email := r.URL.Query().Get("email")

	if _, err := mail.ParseAddress(email); err != nil {
        log.Println("invalid email")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Invalid email address"))
		return
	}
}

func (s *Client) ContactDeletePostHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Deleting Contact")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["contact_id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	s.Delete(id)

	http.Redirect(w, r, "/contact", http.StatusSeeOther)
}

func (s *Client) ContactEditPostHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Editing Contact POST")

	data := Data{
		Title:       "HTMX with Go",
		SearchQuery: "",
	}

	vars := mux.Vars(r)
	contactID, err := strconv.Atoi(vars["contact_id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

    c := s.Update(contactID, r.FormValue("first_name"), r.FormValue("last_name"), r.FormValue("phone"), r.FormValue("email"))

	tmpl, err := template.ParseFiles("template/layout.html", "template/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    if c.Save() {
        http.Redirect(w, r, fmt.Sprintf("/contact/%d", contactID), http.StatusSeeOther)
    } else {
		data.Contacts = append(data.Contacts, c)
        tmpl.ExecuteTemplate(w, "layout.html", data) // Renders the template with the contact data.
    }
}

func (s *Client) ContactEditHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Editing Contact GET")

	tmpl, err := template.ParseFiles("template/layout.html", "template/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := Data{
		Title:       "HTMX with Go",
		SearchQuery: "",
	}

	vars := mux.Vars(r)
	contactID, err := strconv.Atoi(vars["contact_id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	c := s.Find(contactID)

    data.Contacts = append(data.Contacts, c)

	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func (s *Client) ContactViewHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Viewing Contact")

	tmpl, err := template.ParseFiles("template/show.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	contactID, err := strconv.Atoi(vars["contact_id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	c := s.Find(contactID)

	tmpl.ExecuteTemplate(w, "show.html", c)
}

func (s *Client) ContactHandler(w http.ResponseWriter, r *http.Request) {

    log.Println("Contact new contact")

	q := r.URL.Query().Get("q")

	data := Data{
		Title:       "HTMX with Go",
		SearchQuery: q,
	}

    for _, contact := range s.db {
        data.Contacts = append(data.Contacts, contact)
    }

	tmpl, err := template.ParseFiles("template/layout.html", "template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func (s *Client) ContactsNewGetHandler(w http.ResponseWriter, r *http.Request) {

    log.Println("Getting new contact")

	data := Data{
		Title: "HTMX with Go",
		SearchQuery: "",
	}

	tmpl, err := template.ParseFiles("template/layout.html", "template/new.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    data.Contacts = append(data.Contacts, &Contact{})

    tmpl.ExecuteTemplate(w, "layout.html", data)
}

func (s *Client) IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/contact", http.StatusMovedPermanently)
}

func (s *Client) ContactsNewPostHandler(w http.ResponseWriter, r *http.Request) {

    log.Println("Posting new contact")

	tmpl, err := template.ParseFiles("template/layout.html", "template/new.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    c := &Contact{
        Id:    3,
        First: r.FormValue("first_name"),
        Last:  r.FormValue("last_name"),
        Phone: r.FormValue("phone"),
        Email: r.FormValue("email"),
    }

    if c.Save() {
        http.Redirect(w, r, "/contact", http.StatusSeeOther)
    } else {
        tmpl.ExecuteTemplate(w, "layout.html", c)
    }
}
