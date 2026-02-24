package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/federicopalou/sacrif-station/internal/models"
	_ "modernc.org/sqlite"
)

// application holds the dependencies for our HTTP handlers
type application struct {
	entries *models.EntryModel
}

func main() {
	// Initialize the SQLite database connection
	db, err := sql.Open("sqlite", "sacrif.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize our custom application struct
	app := &application{
		entries: &models.EntryModel{DB: db},
	}

	// Ensure the database tables exist
	if err := app.entries.InitSchema(); err != nil {
		log.Fatal("Failed to initialize schema:", err)
	}

	// Check if DB is empty, if so, SEED initial testing data
	count, err := app.entries.Count()
	if err == nil && count == 0 {
		log.Println("Database is empty. Injecting seed data...")
		app.entries.Insert("Hyperion", "book", "Dan Simmons. A structural masterpiece. The Priest's Tale is one of the most haunting things I've ever read.", "")
		app.entries.Insert("The Expanse", "anime", "The most grounded sci-fi television currently in existence. The political tension between Earth, Mars, and the Belt is perfectly executed.", "")
		app.entries.Insert("Inertia", "thought", "The concept of an organic compendium fits perfectly. Things don't need rigid boxes, just a type tag and a display heuristic. Building this feels like carving out a quiet corner of the internet.", "")
	}

	mux := http.NewServeMux()

	// Define the routes for our sectors
	mux.HandleFunc("GET /", app.homeHandler)
	mux.HandleFunc("GET /media", app.mediaHandler)
	mux.HandleFunc("GET /thoughts", app.thoughtsHandler)
	mux.HandleFunc("GET /admin/add", app.createEntryHandler)
	mux.HandleFunc("POST /admin/add", app.createEntryPostHandler)

	log.Println("Starting server on :4000")
	err = http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}

// homeHandler renders the Root Domain landing page
func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ts, err := template.ParseFiles("./ui/html/base.tmpl", "./ui/html/pages/home.tmpl")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}
}

// mediaHandler renders the Media Compendium (everything EXCEPT thoughts/logs)
func (app *application) mediaHandler(w http.ResponseWriter, r *http.Request) {
	// Let's fetch the latest 50 entries that are NOT "thought"
	latestEntries, err := app.entries.LatestExcluded("thought", 50)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	ts, err := template.ParseFiles("./ui/html/base.tmpl", "./ui/html/pages/media.tmpl")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.ExecuteTemplate(w, "base", latestEntries)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}
}

// thoughtsHandler renders the Organic Thoughts Sector (ONLY thoughts/logs)
func (app *application) thoughtsHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch the latest 50 thought entries
	latestEntries, err := app.entries.LatestByType("thought", 50)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	ts, err := template.ParseFiles("./ui/html/base.tmpl", "./ui/html/pages/thoughts.tmpl")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.ExecuteTemplate(w, "base", latestEntries)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}
}

// createEntryHandler renders the admin form GET /admin/add
func (app *application) createEntryHandler(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("./ui/html/base.tmpl", "./ui/html/pages/create.tmpl")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	}
}

// createEntryPostHandler processes the form submission POST /admin/add
func (app *application) createEntryPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	title := r.PostForm.Get("title")
	entryType := r.PostForm.Get("type")
	content := r.PostForm.Get("content")
	url := r.PostForm.Get("url")

	// Insert into SQLite database
	_, err = app.entries.Insert(title, entryType, content, url)
	if err != nil {
		log.Println("Database insert error:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Redirect back to root to drop them into the appropriate sector automatically
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
