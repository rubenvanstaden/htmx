package main

import (
	"sort"
	"strconv"
	"strings"
)

type Repository struct {
	db map[string]*Profile
}

func (s *Repository) Store(profile *Profile) error {
	s.db[profile.PubKey] = profile
	return nil
}

func (s *Repository) All() []*Profile {
	var profiles []*Profile
	for _, v := range s.db {
		profiles = append(profiles, v)
	}

	sort.Slice(profiles, func(i, j int) bool {
		intI, _ := strconv.Atoi(strings.TrimPrefix(profiles[i].PubKey, "npub"))
		intJ, _ := strconv.Atoi(strings.TrimPrefix(profiles[j].PubKey, "npub"))
		return intI < intJ
	})

	return profiles
}

func (s *Repository) Find(id string) (*Profile, error) {
	return s.db[id], nil
}

func (s *Repository) Search(search string) []*Profile {

	profiles := []*Profile{}

	for _, p := range s.db {
		if strings.Contains(p.Name, search) {
			profiles = append(profiles, p)
		}
	}

	return profiles
}

func (s *Repository) Update(name, email, pub string) *Profile {

	p := &Profile{
		Name:   name,
		Email:  email,
		PubKey: pub,
	}

	s.db[pub] = p

	return p
}

func (s *Repository) Delete(id string) {
	delete(s.db, id)
}
