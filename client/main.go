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
	ID        int64     `json:"id"`
	Info      NoteInfo  `json:"info"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func createNote() (Note, error) {
	note := NoteInfo{
		Title:    "Test note",
		Content:  "This is a test note",
		Author:   "Test author",
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

func getAllNotes() ([]Note, error) {
	resp, err := http.Get(baseURL + notesUrl)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("notes not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var notes []Note
	if err := json.NewDecoder(resp.Body).Decode(&notes); err != nil {
		return nil, err
	}

	return notes, nil
}

func updateNote(id int64) (Note, error) {
	updatedInfo := NoteInfo{
		Title:    "Updated test note",
		Content:  "This note was fully replaced",
		Author:   "Updated author",
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

	// Create a note
	note, err := createNote()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(color.RedString("Created note:"), color.GreenString("%+v", note))

	// Get the note
	note, err = getNote(note.ID)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(color.RedString("Got note:"), color.GreenString("%+v", note))

	// Get all notes
	notes, err := getAllNotes()
	if err != nil {
		log.Fatal(err)
	}

	for _, n := range notes {
		log.Printf(color.RedString("\tGot note:"), color.GreenString("%+v", n))
	}

	// update the note
	note, err = updateNote(note.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(color.RedString("Updated note:"), color.GreenString("%+v", note))

	// modify the note
	note, err = modifyNote(note.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(color.RedString("Modified note:"), color.GreenString("%+v", note))

	// delete the note
	if err := deleteNote(note.ID); err != nil {
		log.Fatal(err)
	}
	log.Printf(color.RedString("Deleted note with id:"), color.GreenString("%d", note.ID))

	// get all notes again
	notes, err = getAllNotes()
	if err != nil {
		log.Fatal(err)
	}

	for _, n := range notes {
		log.Printf(color.RedString("\tGot note after delete:"), color.GreenString("%+v", n))
	}
}
