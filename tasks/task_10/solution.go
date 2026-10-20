package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{
		clock: clock,
		tasks: make(map[string]Task),
	}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrInvalidTitle
	}

	r.seq++
	id := strconv.FormatUint(r.seq, 10)

	task := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}

	r.tasks[id] = task

	return task, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, false
	}

	return task, true
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()

	tasks := make([]Task, 0)

	for _, task := range r.tasks {
		if task.Done == done {
			tasks = append(tasks, task)
		}
	}

	r.mu.RUnlock()

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].UpdatedAt.Equal(tasks[j].UpdatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	return tasks
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if task.Done != done {
		task.Done = done
		task.UpdatedAt = r.clock.Now()
		r.tasks[id] = task
	}

	return task, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo: repo}
}
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/tasks" {
		switch r.Method {
		case http.MethodPost:
			h.handleCreate(w, r)
		case http.MethodGet:
			h.handleList(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	} else if strings.HasPrefix(path, "/tasks/") {
		id := strings.TrimPrefix(path, "/tasks/")

		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, r, id)
		case http.MethodPatch:
			h.handlePatch(w, r, id)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	} else {
		http.NotFound(w, r)
	}
}
func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
    	Title *string `json:"title"`
	}
	err := decodeStrictJSON(r.Body, &body)

	if err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	} else if body.Title == nil {
		http.Error(w, "Необходимо ввести название", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(*body.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			http.Error(w, "Некорректное имя", http.StatusBadRequest)
			return
		}

		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, task)
}
func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	values, exists := r.URL.Query()["done"]

	done := false

	if exists {
		if len(values) != 1 || values[0] == "" {
			http.Error(w, "Неверные параметры", http.StatusBadRequest)
			return
		}

		switch values[0] {
		case "true":
			done = true
		case "false":
			done = false
		default:
			http.Error(w, "Некорректный параметр done", http.StatusBadRequest)
			return
		}
	}

	tasks := h.repo.List(done)

	writeJSON(w, http.StatusOK, tasks)
}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Done *bool `json:"done"`
	}

	err := decodeStrictJSON(r.Body, &body);
	if err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	if body.Done == nil {
		http.Error(w, "Необходимо ввести готовность", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *body.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "Некорректное имя", http.StatusNotFound)
			return
		}

		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, task)
}
func decodeStrictJSON(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '{' {
		return errors.New("Ожидался JSON")
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	decodeErr := dec.Decode(v);
	if  decodeErr != nil {
		return decodeErr
	}

	var extra any
	secondDecodeErr := dec.Decode(&extra)
	if  secondDecodeErr != io.EOF {
		if secondDecodeErr != nil {
			return secondDecodeErr
		}
		return errors.New("multiple JSON values")
	}

	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v);
	if  err != nil {
		return
	}
}
