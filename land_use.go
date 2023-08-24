package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jehiah/legislator/legistar"
	"github.com/julienschmidt/httprouter"
)

// IsValidFileNumber matches 0123-2020
func IsValidFileNumber(file string) bool {
	if ok, _ := regexp.MatchString("^[0-9]{4}-(19|20)[9012][0-9]$", file); !ok {
		return false
	}
	n := strings.Split(file, "-")
	seq, _ := strconv.Atoi(n[0])
	if seq > 3500 || seq < 1 {
		return false
	}
	year, _ := strconv.Atoi(n[1])
	if year > time.Now().Year() || year < 1996 {
		return false
	}
	return true
}

// LandUseRedirect redirects from /1234-2020 to the URL for File "LU 1234-2020"
//
// Redirects are cached for the lifetime of the process but not persisted
func (a *App) LandUseRedirect(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := ps.ByName("file")
	if !IsValidFileNumber(file) {
		http.Error(w, "Not Found", 404)
		return
	}
	file = fmt.Sprintf("LU %s", file)

	if redirect, ok := a.cachedRedirects[file]; ok {
		a.addExpireHeaders(w, time.Hour)
		http.Redirect(w, r, redirect, 302)
		return
	}

	filter := legistar.AndFilters(
		legistar.MatterTypeFilter("Land Use Application"),
		legistar.MatterFileFilter(file),
	)

	matters, err := a.legistar.Matters(r.Context(), filter)
	if err != nil {
		log.Print(err)
		http.Error(w, "unknown error", 500)
		return
	}
	if len(matters) != 1 {
		// TODO: cache?
		http.Error(w, "Not Found", 404)
		return
	}

	// we have one
	redirect, err := a.legistar.LookupWebURL(r.Context(), matters[0].ID)
	if err != nil {
		log.Print(err)
		http.Error(w, "unknown error", 500)
		return
	}
	a.cachedRedirects[file] = redirect
	a.addExpireHeaders(w, time.Hour)
	http.Redirect(w, r, redirect, 302)
}

// Index returns the root path of `/` for in-browser search
func (a *App) Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	t := newTemplate(a.templateFS, "index.html")
	w.Header().Set("content-type", "text/html")
	a.addExpireHeaders(w, time.Minute*5)
	type Page struct {
		Page  string
		Title string
	}
	body := Page{
		Page:  "land-use",
		Title: "Land Use",
	}
	err := t.ExecuteTemplate(w, "index.html", body)
	if err != nil {
		log.Print(err)
		http.Error(w, "Internal Server Error", 500)
	}
}
