package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fatih/color"
)

const (
	baseUrl  = "http://localhost:8081"
	notesUrl = "/notes"
)

type NoteInfo struct {
	Title    string `json:"title"`
	Context  string `json:"context"`
	Author   string `json:"author"`
	IsPiblic bool   `json:"is_public"`
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
		Context:  "This is a test note",
		Author:   "Test author",
		IsPiblic: true,
	}

	data, err := json.Marshal(note)
	if err != nil {
		return Note{}, err
	}

	resp, err := http.Post(baseUrl+notesUrl, "application/json", bytes.NewBuffer(data))
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
	resp, err := http.Get(fmt.Sprintf(baseUrl+notesUrl+"/%d", id))
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
	resp, err := http.Get(baseUrl + notesUrl)
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

	//print notes
	for _, n := range notes {
		log.Printf(color.RedString("\tGot note:"), color.GreenString("%+v", n))
	}
}
