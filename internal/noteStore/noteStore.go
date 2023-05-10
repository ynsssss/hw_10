package noteStore

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"
	"github.com/redis/go-redis/v9"
)

type Note struct {
    Id        int `json:"id"`
    Text      string `json:"text"`
    CreatedAt *time.Time `json:"created_at"`
    UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type NoteStore struct {
	rdb *redis.Client
}

func New() *NoteStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	ns := &NoteStore{rdb: rdb}
	return ns
}

func (ns *NoteStore) CreateNote(text string) (error, int) {
	curtime := time.Now().Unix()
	index, err := ns.rdb.Incr(context.Background(), "index").Result()
	if err != nil {
		return err, 0
	}
	err = ns.rdb.Set(context.Background(), fmt.Sprintf("%dtext", index), text, 0).Err()
	if err != nil {
		return err, 0
	}
	err = ns.rdb.Set(context.Background(), fmt.Sprintf("%dcreatedat", index), curtime, 0).Err()
    if err != nil {
		return err, 0
	}
	return nil, int(index)
}

func (ns *NoteStore) UpdateNote(text string, id int) error {
	curtime := time.Now().Unix()
	err := ns.rdb.Set(context.Background(), fmt.Sprintf("%dtext", id), text, 0).Err()
	if err != nil {
		return err
	}
	err = ns.rdb.Set(context.Background(), fmt.Sprintf("%dupdatedat", id), curtime, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (ns *NoteStore) DeleteNote(id int) error {
	err := ns.rdb.Del(context.Background(), fmt.Sprintf("%dtext", id)).Err()
	if err != nil {
		return err
	}
	err = ns.rdb.Del(context.Background(), fmt.Sprintf("%dupdatedat", id)).Err()
	if err != nil {
		return err
	}
	err = ns.rdb.Del(context.Background(), fmt.Sprintf("%dcreatedat", id)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (ns *NoteStore) GetNote(id int) (*Note, error) {
	noteText, err := ns.rdb.Get(context.Background(), fmt.Sprintf("%dtext", id)).Result()
	if err != nil {
		return &Note{}, err
	}

	note := &Note{Id: id, Text: noteText}

	noteCreatedAt, err := ns.rdb.Get(context.Background(), fmt.Sprintf("%dcreatedat", id)).Result()
	if err != nil {
		return &Note{}, err
	}
	if noteCreatedAt != "" {
		i, err := strconv.ParseInt(noteCreatedAt, 10, 64)
		if err != nil {
			return &Note{}, err
		}
		note.CreatedAt = toTimePtr(time.Unix(i, 0))
	}

	noteUpdatedAt, err := ns.rdb.Get(context.Background(), fmt.Sprintf("%dupdatedat", id)).Result()
	if err != nil && err != redis.Nil {
		return &Note{}, nil
	}
	if noteUpdatedAt != "" {
		i, err := strconv.ParseInt(noteUpdatedAt, 10, 64)
		if err != nil {
			return &Note{}, err
		}
		note.CreatedAt = toTimePtr(time.Unix(i, 0))
	}
	return note, nil
}

func (ns *NoteStore) GetAllNotes() ([]*Note, error) {
	var notes []*Note
	notescount, err := ns.rdb.Get(context.Background(), "index").Result()
    if err != nil {
		return []*Note{}, err
	}
	notescountint, err := strconv.ParseInt(notescount, 10, 64)
	if err != nil {
		return []*Note{}, err
	}
	for i := 1; i <= int(notescountint); i++ {
		note, err := ns.GetNote(i)
		if err != nil {
			return []*Note{}, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (ns *NoteStore) SortBy(param string, notes []*Note) {
	switch param {
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
	}
}

func toTimePtr(t time.Time) *time.Time {
    return &t
}