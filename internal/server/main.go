package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"ffxiv.anid.dev/internal/manager"
	"ffxiv.anid.dev/internal/models"
	"ffxiv.anid.dev/internal/utils"
	"github.com/Netflix/go-env"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
)

type Environment struct {
	OAuth struct {
		RedirectURL  *string `env:"OAUTH_REDIRECT_URL"`
		ClientID     *string `env:"OAUTH_CLIENT_ID"`
		ClientSecret *string `env:"OAUTH_CLIENT_SECRET"`
	}

	Stage string `env:"STAGE,default=development"`
}

var conf = &oauth2.Config{
	Scopes: []string{
		"identify"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://discord.com/api/oauth2/authorize",
		TokenURL: "https://discord.com/api/oauth2/token",
	},
}

type Server struct {
	router *mux.Router
	tasks  *manager.TasksManager
	store  *sessions.CookieStore
}

func (s *Server) Start() {
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
	}).Handler(s.router)

	http.ListenAndServe(":8000", handler)
}

// NewServer is a wire provider that returns a server
func NewServer(tasks *manager.TasksManager) (*Server, error) {
	var environment Environment
	_, err := env.UnmarshalFromEnviron(&environment)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	r := mux.NewRouter()

	s := &Server{
		router: r,
		tasks:  tasks,
		store:  sessions.NewCookieStore([]byte("password")),
	}

	if environment.OAuth.ClientID != nil && environment.OAuth.ClientSecret != nil && environment.OAuth.RedirectURL != nil {
		conf.ClientID = *environment.OAuth.ClientID
		conf.ClientSecret = *environment.OAuth.ClientSecret
		conf.RedirectURL = *environment.OAuth.RedirectURL

		r.HandleFunc("/oauth", s.oauth).Methods("GET")
		r.HandleFunc("/login", s.login).Methods("GET")
	} else {
		log.Printf("auth disabled due to missing environment vars: %+v", environment.OAuth)
	}

	r.HandleFunc("/", s.yourHandler).Methods("GET")
	r.HandleFunc("/tasks", s.getMaster).Methods("GET")
	r.HandleFunc("/user/tasks", s.getUserMaster).Methods("GET")
	r.HandleFunc("/user/tasks", s.saveUserMaster).Methods("POST")
	r.HandleFunc("/user/tasks/current", s.getUserTasksForToday).Methods("GET")
	r.HandleFunc("/user/tasks/{date:20.*}", s.getUserTasksForDate).Methods("GET")
	r.HandleFunc("/me", s.me).Methods("GET")

	return s, nil
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
	u := s.getUser(r)

	tasks, err := s.tasks.GetUserMasterTasks(u.ID)
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

	u := s.getUser(r)

	tasks, err := s.tasks.SaveUserMasterTasks(u.ID, t)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, tasks)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)

	state := base64.URLEncoding.EncodeToString(b)

	session, _ := s.store.Get(r, "sess")
	session.Values["state"] = state
	session.Options.HttpOnly = true
	session.Options.Secure = true
	session.Options.SameSite = http.SameSiteNoneMode
	session.Save(r, w)

	url := conf.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.Get(r, "sess")
	if err != nil {
		fmt.Fprintln(w, "aborted")
		return
	}

	fmt.Printf("%+v", session.Values)

	if userM, ok := session.Values["user"]; ok {
		if u, ok := userM.(user); ok {
			respondWithJSON(w, 200, u)
		}
	}

	fmt.Fprintln(w, time.Now())
	ye,we := utils.GetFFWeekYear(time.Now())
	fmt.Fprintln(w, ye, we)
}

func (s *Server) getUser(r *http.Request) *user {
	session, err := s.store.Get(r, "sess")
	if err != nil {
		fmt.Printf("returning default user due to error: %s", err)

		return &user{ID: "1"}
	}

	if userM, ok := session.Values["user"]; ok {
		if u, ok := userM.(user); ok {
			return &u
		}

		fmt.Printf("returning default user due to invalid user struct: %v", userM)

		return &user{ID: "1"}
	}

	fmt.Printf("returning default user due to missing user in session: %v", session.Values)

	return &user{ID: "1"}
}

type user struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

func (s *Server) oauth(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.Get(r, "sess")
	if err != nil {
		fmt.Fprintln(w, "aborted")
		return
	}

	if r.FormValue("state") != session.Values["state"] {
		fmt.Fprintln(w, "no state match; possible csrf OR cookies not enabled")
		return
	}

	code := r.FormValue("code")
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		fmt.Fprintf(w, "Code exchange failed with error %s\n", err.Error())
		return
	}

	if !token.Valid() {
		fmt.Fprintln(w, "Retreived invalid token")
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://discord.com/api/users/@me", nil)
	if err != nil {
		fmt.Fprintf(w, "Failed to get user: %s\n", err.Error())
		return
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error getting user from token %s\n", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintf(w, "Failed decoding user: %s\n", err.Error())
		return
	}

	gob.Register(user{})

	var me user
	err = json.Unmarshal(contents, &me)
	if err != nil {
		log.Printf("Error unmarshaling user %s\n", err.Error())
		return
	}

	session.Values["user"] = me
	session.Options.HttpOnly = true
	session.Options.Secure = true
	session.Options.SameSite = http.SameSiteNoneMode
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		fmt.Printf("%+v", session.Values)
	}

	fmt.Printf("%+v", me)

	http.Redirect(w, r, "http://localhost:3000", http.StatusTemporaryRedirect)
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

func (s *Server) getUserTasksForToday(w http.ResponseWriter, r *http.Request) {
	u := s.getUser(r)

	tasks, err := s.tasks.GetUserTasks(u.ID, time.Now())
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, tasks)
}

func (s *Server) getUserTasksForDate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	date, err := time.Parse("2006-01-02", vars["date"])
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	u := s.getUser(r)

	tasks, err := s.tasks.GetUserTasks(u.ID, date)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, tasks)
}
