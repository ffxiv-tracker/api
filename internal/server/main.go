package server

import (
	"encoding/json"
	"net/http"

	"ffxiv.anid.dev/internal/manager"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	router *mux.Router
	tasks  *manager.TasksManager
}

func (s *Server) Start() {
	handler := cors.Default().Handler(s.router)

	http.ListenAndServe(":8000", handler)
}

// NewServer is a wire provider that returns a server
func NewServer(tasks *manager.TasksManager) *Server {
	r := mux.NewRouter()

	s := &Server{
		router: r,
		tasks:  tasks,
	}

	r.HandleFunc("/", s.yourHandler)
	r.HandleFunc("/tasks", s.tasksHandler)

	return s
}

func (s *Server) yourHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!\n"))
}

func (s *Server) tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.tasks.GetMasterTasks()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, tasks)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
