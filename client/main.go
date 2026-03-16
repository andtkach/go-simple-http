package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
)

const (
	defaultBaseURL = "http://localhost:8081"
	notesUrl       = "/notes"
)

var baseURL = getBaseURL()

type NoteInfo struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Author   string `json:"author"`
	IsPublic bool   `json:"is_public"`
}

type NoteInfoPatch struct {
	Title    *string `json:"title"`
	Content  *string `json:"content"`
	Author   *string `json:"author"`
	IsPublic *bool   `json:"is_public"`
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

func createNote(title string, content string) (Note, error) {
	note := NoteInfo{
		Title:    title,
		Content:  content,
		Author:   "Andrii",
		IsPublic: true,
	}

	data, err := json.Marshal(note)
	if err != nil {
		return Note{}, err
	}

	resp, err := http.Post(baseURL+notesUrl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return Note{}, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		return Note{}, err
	}

	var cretedNote Note
	if err := json.NewDecoder(resp.Body).Decode(&cretedNote); err != nil {
		return Note{}, err
	}

	return cretedNote, nil
}

func getNote(id int64) (Note, error) {
	resp, err := http.Get(fmt.Sprintf(baseURL+notesUrl+"/%d", id))
	if err != nil {
		return Note{}, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return Note{}, fmt.Errorf("note with id %d not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		return Note{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var note Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return Note{}, err
	}

	return note, nil
}

func getAllNotes() (NotesPage, error) {
	resp, err := http.Get(baseURL + notesUrl)
	if err != nil {
		return NotesPage{}, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return NotesPage{}, fmt.Errorf("notes not found")
	}

	if resp.StatusCode != http.StatusOK {
		return NotesPage{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var notesPage NotesPage
	if err := json.NewDecoder(resp.Body).Decode(&notesPage); err != nil {
		return NotesPage{}, err
	}

	return notesPage, nil
}

func updateNote(id int64, title string, content string) (Note, error) {
	updatedInfo := NoteInfo{
		Title:    title,
		Content:  content,
		Author:   "Andrii",
		IsPublic: false,
	}

	data, err := json.Marshal(updatedInfo)
	if err != nil {
		return Note{}, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf(baseURL+notesUrl+"/%d", id), bytes.NewBuffer(data))
	if err != nil {
		return Note{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Note{}, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return Note{}, fmt.Errorf("note with id %d not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		return Note{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var note Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return Note{}, err
	}

	return note, nil
}

func modifyNote(id int64) (Note, error) {
	content := "This note was partially updated"
	patch := NoteInfoPatch{
		Content: &content,
	}

	data, err := json.Marshal(patch)
	if err != nil {
		return Note{}, err
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf(baseURL+notesUrl+"/%d", id), bytes.NewBuffer(data))
	if err != nil {
		return Note{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Note{}, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return Note{}, fmt.Errorf("note with id %d not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		return Note{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var note Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return Note{}, err
	}

	return note, nil
}

func deleteNote(id int64) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(baseURL+notesUrl+"/%d", id), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("note with id %d not found", id)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func getBaseURL() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return v
	}

	return defaultBaseURL
}

func main() {

	log.Println("Start notes client")

	// Create a note
	note, err := createNote("Note 1", "This is the content of note 1")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(color.RedString("Created note with id: %d\t", note.Id), color.GreenString("%+v", note))
	log.Println()

	// Get the note
	note, err = getNote(note.Id)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(color.RedString("Got note by id: %d\t", note.Id), color.GreenString("%+v", note))
	log.Println()

	// Get all notes
	log.Println(color.RedString("Getting all notes..."))
	notesPage, err := getAllNotes()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(color.RedString("Page %d/%d, limit=%d, total=%d", notesPage.Page, notesPage.TotalPages, notesPage.Limit, notesPage.TotalCount))

	for _, n := range notesPage.Notes {
		log.Printf(color.RedString("\tGot note %d:\t", n.Id), color.GreenString("%+v", n))
	}
	log.Println()

	// update the note
	note, err = updateNote(note.Id, "Updated Note 1", "This is the updated content of note 1")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(color.RedString("Updated note:\t"), color.GreenString("%+v", note))
	log.Println()

	// modify the note
	note, err = modifyNote(note.Id)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(color.RedString("Modified note:\t"), color.GreenString("%+v", note))
	log.Println()

	// delete the note
	if err := deleteNote(note.Id); err != nil {
		log.Fatal(err)
	}
	log.Printf(color.RedString("Deleted note with id: %d\t", note.Id), color.GreenString("%+v", note))
	log.Println()

	// get all notes again
	log.Println(color.RedString("Getting all notes again..."))
	notesPage, err = getAllNotes()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(color.RedString("Page %d/%d, limit=%d, total=%d", notesPage.Page, notesPage.TotalPages, notesPage.Limit, notesPage.TotalCount))

	for _, n := range notesPage.Notes {
		log.Printf(color.RedString("\tGot note after delete:"), color.GreenString("%+v", n))
	}

	log.Println("Notes client finished")
}
