package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"notesapp/internal/noteStore"
	"sort"
	"strconv"
	"strings"
)

type noteServer struct {
	store *noteStore.NoteStore
}

func newNoteServer() *noteServer {
	store := noteStore.New()
	return &noteServer{store: store}
}

func (ns *noteServer) noteHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/note" {
		switch req.Method {
		case http.MethodGet:
			ns.GetAllNotesHandler(w, req)
		case http.MethodPost:
			ns.CreateNoteHandler(w, req)
		}
	} else {
		path := strings.Trim(req.URL.Path, "/")
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			http.Error(w, "expect /task/<id> in task handler", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch req.Method {
		case http.MethodGet:
			ns.GetNoteByIdHandler(w, req, id)
		case http.MethodPut:
			ns.ChangeNoteByIdHandler(w, req, id)
		case http.MethodDelete:
			ns.DeleteNoteByIdHandler(w, req, id)
		}
	}
}

func (ns *noteServer) GetAllNotesHandler(w http.ResponseWriter, req *http.Request) {
	notes, err := ns.store.GetAllNotes()
	order_by := req.URL.Query().Get("order_by")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	switch order_by {
	case "id":
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Id < notes[j].Id
		})
	case "text":
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Text < notes[j].Text
		})
	case "created_at":
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].CreatedAt.Unix() < notes[j].CreatedAt.Unix()
		})
	case "updated_at":
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].UpdatedAt.Unix() < notes[j].UpdatedAt.Unix()
		})
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
	j, err := json.Marshal(notes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(j)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ns *noteServer) CreateNoteHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	note := noteStore.Note{}
	err = json.Unmarshal(body, &note)
	if (err != nil) || (note.Text == "") {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	err, note.Id = ns.store.CreateNote(note.Text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	j := []byte(fmt.Sprintf("{\"id\":%d}", note.Id))
	_, err = w.Write(j)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ns *noteServer) GetNoteByIdHandler(w http.ResponseWriter, req *http.Request, id int) {
	note, err := ns.store.GetNote(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	j, err := json.Marshal(note)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(j)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ns *noteServer) ChangeNoteByIdHandler(w http.ResponseWriter, req *http.Request, id int) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	var note noteStore.Note
	err = json.Unmarshal(body, &note)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	err = ns.store.UpdateNote(note.Text, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (ns *noteServer) DeleteNoteByIdHandler(w http.ResponseWriter, req *http.Request, id int) {
	if err := ns.store.DeleteNote(id); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type App struct {
	mux    *http.ServeMux
	server *noteServer
}

func main() {
	mux := http.NewServeMux()
	server := newNoteServer()
	app := &App{mux: mux, server: server}

	app.mux.HandleFunc("/note/", server.noteHandler)

	if err := http.ListenAndServe(":1234", mux); err != nil {
		fmt.Printf("error while starting service: %v", err)
	}
}