package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

const (
	defaultHost    = "localhost"
	defaultPort    = "8081"
	defaultPage    = 1
	defaultLimit   = 10
	createNoteUrl  = "/notes"
	getNoteUrl     = "/notes/{id}"
	getAllNotesUrl = "/notes"
	updateNoteUrl  = "/notes/{id}"
	modifyNoteUrl  = "/notes/{id}"
	deleteNoteUrl  = "/notes/{id}"
)

type NoteInfo struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Author   string `json:"author"`
	IsPublic bool   `json:"is_public"`
}

type Note struct {
	Id        int64     `json:"id"`
	Info      NoteInfo  `json:"info"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotesPage struct {
	Notes      []Note `json:"notes"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalCount int    `json:"total_count"`
	TotalPages int    `json:"total_pages"`
}

type NoteInfoPatch struct {
	Title    *string `json:"title"`
	Content  *string `json:"content"`
	Author   *string `json:"author"`
	IsPublic *bool   `json:"is_public"`
}

type SyncMap struct {
	elems map[int64]*Note
	m     sync.RWMutex
}

var notes = SyncMap{
	elems: make(map[int64]*Note),
}

func createNoteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(color.BlueString("Received request: %s %s", r.Method, r.URL.Path))
	info := &NoteInfo{}
	if err := json.NewDecoder(r.Body).Decode(info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rand.Seed(time.Now().UnixNano())
	now := time.Now()

	newNoteId := rand.Int63()

	// check I do not have note with such id already, if I do - generate 500 error
	notes.m.RLock()
	if _, exists := notes.elems[newNoteId]; exists {
		notes.m.RUnlock()
		http.Error(w, "error generating note id", http.StatusInternalServerError)
		return
	}
	notes.m.RUnlock()

	note := &Note{
		Id:        newNoteId,
		Info:      *info,
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notes.m.Lock()
	defer notes.m.Unlock()
	notes.elems[note.Id] = note
	log.Println(color.GreenString("Note created: %+v", note))
}

func getNoteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(color.BlueString("Received request: %s %s", r.Method, r.URL.Path))
	noteId := chi.URLParam(r, "id")
	id, err := parseNoteId(noteId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	notes.m.RLock()
	defer notes.m.RUnlock()

	note, ok := notes.elems[id]
	if !ok {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(color.GreenString("Note retrieved: %+v", note))
}

func getAllNotesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(color.BlueString("Received request: %s %s", r.Method, r.URL.Path))
	page, limit, err := getPaginationParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	notes.m.RLock()
	defer notes.m.RUnlock()

	notesList := make([]Note, 0, len(notes.elems))
	for _, note := range notes.elems {
		notesList = append(notesList, *note)
	}
	totalCount := len(notesList)

	sort.Slice(notesList, func(i, j int) bool {
		return notesList[i].CreatedAt.Before(notesList[j].CreatedAt)
	})

	start := (page - 1) * limit
	if start >= len(notesList) {
		notesList = []Note{}
	} else {
		end := start + limit
		if end > len(notesList) {
			end = len(notesList)
		}
		notesList = notesList[start:end]
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}

	response := NotesPage{
		Notes:      notesList,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(color.GreenString("Notes retrieved: page=%d limit=%d returned=%d total=%d", page, limit, len(notesList), totalCount))
}

func updateNoteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(color.BlueString("Received request: %s %s", r.Method, r.URL.Path))
	noteId := chi.URLParam(r, "id")
	id, err := parseNoteId(noteId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedInfo := &NoteInfo{}
	if err := json.NewDecoder(r.Body).Decode(updatedInfo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	notes.m.Lock()
	defer notes.m.Unlock()

	note, ok := notes.elems[id]
	if !ok {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	note.Info = *updatedInfo
	note.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(color.GreenString("Note updated: %+v", note))
}

func modifyNoteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(color.BlueString("Received request: %s %s", r.Method, r.URL.Path))
	noteId := chi.URLParam(r, "id")
	id, err := parseNoteId(noteId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	patch := &NoteInfoPatch{}
	if err := json.NewDecoder(r.Body).Decode(patch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	notes.m.Lock()
	defer notes.m.Unlock()

	note, ok := notes.elems[id]
	if !ok {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	if patch.Title != nil {
		note.Info.Title = *patch.Title
	}
	if patch.Content != nil {
		note.Info.Content = *patch.Content
	}
	if patch.Author != nil {
		note.Info.Author = *patch.Author
	}
	if patch.IsPublic != nil {
		note.Info.IsPublic = *patch.IsPublic
	}
	note.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(color.GreenString("Note modified: %+v", note))
}

func deleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(color.BlueString("Received request: %s %s", r.Method, r.URL.Path))
	noteId := chi.URLParam(r, "id")
	id, err := parseNoteId(noteId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	notes.m.Lock()
	defer notes.m.Unlock()

	if _, ok := notes.elems[id]; !ok {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	delete(notes.elems, id)
	w.WriteHeader(http.StatusNoContent)
	log.Println(color.GreenString("Note deleted: id=%d", id))
}

func parseNoteId(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func getPaginationParams(r *http.Request) (int, int, error) {
	page := defaultPage
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		parsedPage, err := strconv.Atoi(pageParam)
		if err != nil || parsedPage < 1 {
			return 0, 0, fmt.Errorf("page must be a positive integer")
		}
		page = parsedPage
	}

	limit := defaultLimit
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit < 1 {
			return 0, 0, fmt.Errorf("limit must be a positive integer")
		}
		limit = parsedLimit
	}

	return page, limit, nil
}

func getListenAddr() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = defaultHost
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	return net.JoinHostPort(host, port)
}

func main() {
	_ = godotenv.Load()

	r := chi.NewRouter()

	r.Post(createNoteUrl, createNoteHandler)
	r.Get(getNoteUrl, getNoteHandler)
	r.Get(getAllNotesUrl, getAllNotesHandler)
	r.Put(updateNoteUrl, updateNoteHandler)
	r.Patch(modifyNoteUrl, modifyNoteHandler)
	r.Delete(deleteNoteUrl, deleteNoteHandler)

	listenAddr := getListenAddr()
	log.Printf("Server starting on %s", listenAddr)
	err := http.ListenAndServe(listenAddr, r)
	if err != nil {
		log.Fatal(err)
	}

}
