package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"notesapp/internal/noteStore"
	"testing"
)

func TestCreateNoteHandler(t *testing.T) {
	server := newNoteServer()

	note := &noteStore.Note{Text: "do the dishes"}

	jsonText, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("Couldnt marshal text to json: %v\n", err)
	}
	bodyReader := bytes.NewReader(jsonText)

	req, err := http.NewRequest(http.MethodPost, "/note", bodyReader)
	if err != nil {
		t.Fatalf("Couldnt create request: %v\n", err)
	}

	w := httptest.NewRecorder()

	server.noteHandler(w, req)

	if w.Code != 201 {
		t.Errorf("wrong status code: %d", w.Code)
	}
	wrongNote := struct {
		Message string `json:"message"`
	}{
		Message: "wrongNote",
	}
	jsonText, err = json.Marshal(wrongNote)
	if err != nil {
		t.Fatalf("Couldnt marshal text to json: %v\n", err)
	}
	bodyReader = bytes.NewReader(jsonText)

	req, err = http.NewRequest(http.MethodPost, "/note", bodyReader)
	if err != nil {
		t.Fatalf("Couldnt create request: %v\n", err)
	}

	w = httptest.NewRecorder()

	server.noteHandler(w, req)
	if w.Code != 422 {
		t.Errorf("wrong status code: %d", w.Code)
	}
}

//first we send post method with test node to append to db and later check if it is retrievable by its id
func TestGetNoteHandler(t *testing.T) {
	server := newNoteServer()
	postNote := &noteStore.Note{Text: "post note text"}
    jsonText, err := json.Marshal(postNote)
	if err != nil {
		t.Fatalf("Couldnt marshal text to json: %v\n", err)
	}
	bodyReader := bytes.NewReader(jsonText)

	req, err := http.NewRequest(http.MethodPost, "/note", bodyReader)
	if err != nil {
		t.Fatalf("Couldnt create request: %v\n", err)
	}

	w := httptest.NewRecorder()

	server.noteHandler(w, req)
    var result map[string]int
    err = json.NewDecoder(w.Body).Decode(&result)
    if err != nil {
		t.Fatalf("Couldnt decode data: %v\n", err)
    }
    id := result["id"]
    req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/note/%d", id), nil)
    if err != nil {
		t.Fatalf("Couldnt create request: %v\n", err)
	}
    w = httptest.NewRecorder()
	server.noteHandler(w, req)
    var getNote noteStore.Note
    err = json.NewDecoder(w.Body).Decode(&getNote)
    if err != nil {
		t.Fatalf("Couldnt decode data: %v\n", err)
    }
    if getNote.Text != postNote.Text {
        t.Errorf("wrong data: want %s, got %s",postNote.Text, getNote.Text)
    }
}