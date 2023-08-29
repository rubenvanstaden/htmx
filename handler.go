package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/url"
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
	Profiles    []*Profile
	SearchQuery string
}

type Handler struct {
	repository Repository
}

func (s *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/contact", http.StatusMovedPermanently)
}

func (s *Handler) SearchProfile(w http.ResponseWriter, r *http.Request) {

	page, contacts := s.paging(r.URL.Query().Get("page"))

	search := r.URL.Query().Get("q")
	if search != "" {
		header := r.Header.Get("HX-Trigger")
		if header == "search" {
			log.Println("Header search FOUND")
			contacts = s.repository.Search(search)
		}
	}

	tmpl, err := template.New("home").Funcs(funcMap).ParseFiles("template/layout.html", "template/index.html", "template/row.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := Data{
		Page:        page,
		Profiles:    contacts,
		SearchQuery: search,
	}
	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func (s *Handler) paging(pageStr string) (int, []*Profile) {

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	profiles := s.repository.All()

	if len(profiles) == 0 {
		log.Fatalln("no profiles found")
	}

	startIndex := (page - 1) * 10
	endIndex := startIndex + 10
	if endIndex > len(profiles) {
		endIndex = len(profiles)
	}

	return page, profiles[startIndex:endIndex]
}

func (s *Handler) ParseEmail(w http.ResponseWriter, r *http.Request) {

	log.Println("Checking mail")

	email := r.URL.Query().Get("email")

	if _, err := mail.ParseAddress(email); err != nil {
		log.Println("invalid email")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Invalid email address"))
		return
	}
}

func (s *Handler) DeleteBulk(w http.ResponseWriter, r *http.Request) {

	log.Println("Deleting BULK")

	// Read the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	r.Body.Close() // Always close the body when done

	// Parse the form data from the body
	values, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}

	// Retrieve the selected_contact_ids values
	selectedContactIDs := values["selected_pubkeys"]
	if selectedContactIDs == nil {
		http.Error(w, "No contacts selected", http.StatusBadRequest)
		return
	}

	for _, pub := range selectedContactIDs {
		log.Printf("pubkey deleteing: %s", pub)
		s.repository.Delete(pub)
	}

	page, profiles := s.paging(r.URL.Query().Get("page"))
	data := Data{
		Page:     page,
		Profiles: profiles,
	}

	tmpl, err := template.ParseFiles("template/layout.html", "template/index.html", "template/row.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func (s *Handler) DeleteProfile(w http.ResponseWriter, r *http.Request) {

	log.Println("Deleting Profile")

	vars := mux.Vars(r)
	pubkey := vars["pubkey"]

	s.repository.Delete(pubkey)

	if r.Header.Get("HX-Trigger") == "delete-btn" {
		log.Println("Header DeLETE")
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Handler) SaveProfile(w http.ResponseWriter, r *http.Request) {

	pubkey := r.FormValue("pubkey")
	email := r.FormValue("email")

	c := s.repository.Update(r.FormValue("name"), email, pubkey)

	tmpl, err := template.ParseFiles("template/layout.html", "template/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := Data{}
	err = s.repository.Store(c)
	if err != nil {
		data.Profiles = append(data.Profiles, c)
		tmpl.ExecuteTemplate(w, "layout.html", data) // Renders the template with the contact data.
	} else {
		http.Redirect(w, r, fmt.Sprintf("/contact/%s", pubkey), http.StatusSeeOther)
	}
}

func (s *Handler) EditProfile(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	pubkey := vars["pubkey"]
	if pubkey == "" {
		log.Fatalln("pubkey cannot be empty")
	}
	p, err := s.repository.Find(pubkey)
	if err != nil {
		log.Fatalln(err)
	}

	data := Data{}
	data.Profiles = append(data.Profiles, p)
	if len(data.Profiles) != 1 {
		log.Fatalln("cannot edit more than one profile")
	}

	tmpl, err := template.ParseFiles("template/layout.html", "template/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "edit.html", data)
}

func (s *Handler) ShowProfile(w http.ResponseWriter, r *http.Request) {

	log.Println("Viewing Profile")

	vars := mux.Vars(r)
	pubkey := vars["pubkey"]
	if pubkey == "" {
		log.Fatalln("pubkey cannot be empty")
	}
	p, err := s.repository.Find(pubkey)
	if err != nil {
		log.Fatalln(err)
	}

	tmpl, err := template.ParseFiles("template/layout.html", "template/show.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "show.html", p)
}

func (s *Handler) NewProfile(w http.ResponseWriter, r *http.Request) {

	log.Println("Getting new contact")

	data := Data{}
	data.Profiles = append(data.Profiles, &Profile{})

	tmpl, err := template.ParseFiles("template/layout.html", "template/new.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "new.html", data)
}

func (s *Handler) AddProfile(w http.ResponseWriter, r *http.Request) {

	log.Println("Posting new contact")

	tmpl, err := template.ParseFiles("template/layout.html", "template/new.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	profile := &Profile{
		Name:   r.FormValue("name"),
		Email:  r.FormValue("email"),
		PubKey: r.FormValue("pubkey"),
	}

	// FIXME
	err = s.repository.Store(profile)
	if err != nil {
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
	} else {
		tmpl.ExecuteTemplate(w, "new.html", profile)
	}
}
