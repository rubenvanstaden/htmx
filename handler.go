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

	data := Data{
		Page:        page,
		Title:       "HTMX with Go",
		Profiles:    contacts,
		SearchQuery: search,
	}

	tmpl, err := template.New("home").Funcs(funcMap).ParseFiles("template/layout.html", "template/index.html", "template/row.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

func (s *Handler) DeleteProfile(w http.ResponseWriter, r *http.Request) {

	log.Println("Deleting Profile")

	vars := mux.Vars(r)
	pubkey := vars["pubkey"]

	s.repository.Delete(pubkey)

	http.Redirect(w, r, "/contact", http.StatusSeeOther)
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
