package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ffxiv.anid.dev/internal/manager"
	"ffxiv.anid.dev/internal/models"
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

	r.HandleFunc("/", s.yourHandler).Methods("GET")
	r.HandleFunc("/tasks", s.getMaster).Methods("GET")
	r.HandleFunc("/user/tasks", s.getUserMaster).Methods("GET")
	r.HandleFunc("/user/tasks", s.saveUserMaster).Methods("POST")

	return s
}

func (s *Server) yourHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!\n"))
}

func (s *Server) getMaster(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.tasks.GetMasterTasks()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, tasks)
}

func (s *Server) getUserMaster(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.tasks.GetUserMasterTasks("1")
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, tasks)
}

func (s *Server) saveUserMaster(w http.ResponseWriter, r *http.Request) {
	var t *models.UserMasterTaskRequest

    err := json.NewDecoder(r.Body).Decode(&t)
    if err != nil {
		fmt.Printf("INFO: failed to decode user request: %s", err)
		
		respondWithError(w, 400, "unable to decode json request")
		return
    }

	if t == nil || t.Frequency == "" {
		respondWithError(w, 400, "frequency must be specified")
		return
	}

	userErr, err := s.tasks.ValidateUserMasterTaskRequest(t)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return		
	}

	if userErr != nil {
		respondWithError(w, 400, userErr.Error())
		return		
	}
	
	tasks, err := s.tasks.SaveUserMasterTasks("1", t) // todo: replace with user id
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
