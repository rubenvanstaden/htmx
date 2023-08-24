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

// Define template with the functions
var funcMap = template.FuncMap{
	"subtract": func(a, b int) int { return a - b },
	"add":      func(a, b int) int { return a + b },
}

type Data struct {
	Title       string
	Page        int
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
		Last:  lastName,
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

	for i := 0; i < 1000; i++ {

		client.db[i] = &Contact{
			Id:    i,
			First: "alice",
			Email: "alice@gmail.com",
			Phone: fmt.Sprintf("000-%d", i),
		}
	}

	r := mux.NewRouter()

	r.HandleFunc("/", client.IndexHandler)
	r.HandleFunc("/contact", client.PagingContacts)
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

func (s *Client) PagingContacts(w http.ResponseWriter, r *http.Request) {

	log.Println("Paging contacts")

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	log.Println(pageStr)
	log.Println(page)

	var contacts []*Contact
	for _, c := range s.db {
		contacts = append(contacts, c)
	}

	startIndex := (page - 1) * 10
	endIndex := startIndex + 10
	if endIndex > len(contacts) {
		endIndex = len(contacts)
	}

	data := Data{
		Page:     page,
		Title:    "HTMX with Go",
		Contacts: contacts[startIndex:endIndex],
	}

	log.Println(len(data.Contacts))

	tmpl, err := template.New("pagination").Funcs(funcMap).ParseFiles("template/layout.html", "template/index.html", "template/row.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(tmpl)

	tmpl.ExecuteTemplate(w, "layout.html", data)
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

	tmpl, err := template.ParseFiles("template/layout.html", "template/index.html", "template/row.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func (s *Client) ContactsNewGetHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Getting new contact")

	data := Data{
		Title:       "HTMX with Go",
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
