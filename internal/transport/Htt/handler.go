package Htt

import (
	service "GO_Assignment_3/internal/service"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler - struct to handle HTTP requests
type Handler struct {
	Service *service.Service
}

// NewHandler - creates a new Handler instance
func NewHandler(service *service.Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/students", h.GetAllStudents).Methods("GET")
	router.HandleFunc("/students/{id}", h.GetStudent).Methods("GET")
	router.HandleFunc("/students", h.CreateStudent).Methods("POST")
	router.HandleFunc("/students/{id}", h.UpdateStudent).Methods("PUT")
	router.HandleFunc("/students/{id}", h.DeleteStudent).Methods("DELETE")
	router.HandleFunc("/register", h.RegisterUser).Methods("POST")
}

func (h *Handler) GetAllStudents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	students, err := h.Service.GetAllStudents(ctx)
	if err != nil {
		http.Error(w, "Error fetching students", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func (h *Handler) GetStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid student ID format", http.StatusBadRequest)
		return
	}

	student, err := h.Service.GetStudent(ctx, int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

func (h *Handler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var student service.Student

	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdStudent, err := h.Service.AddStudent(ctx, student)
	if err != nil {
		http.Error(w, "Error adding student", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdStudent)
}

func (h *Handler) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid student ID format", http.StatusBadRequest)
		return
	}

	var updatedStudent service.Student
	if err := json.NewDecoder(r.Body).Decode(&updatedStudent); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	student, err := h.Service.UpdateStudent(ctx, int32(id), updatedStudent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

func (h *Handler) DeleteStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid student ID format", http.StatusBadRequest)
		return
	}

	err = h.Service.DeleteStudent(ctx, int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "userType", "user")

	var student service.Student

	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err := h.Service.AddStudent(ctx, student)
	if err != nil {
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	user_ID, err := h.Service.GetLastInsertedStudentID()
	if err != nil {
		http.Error(w, "Error retrieving student ID", http.StatusInternalServerError)
		return
	}

	token, err := GenerateJWT(user_ID, student.Name)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *Handler) Serve() error {
	router := mux.NewRouter()

	router.Use(JSONMiddleware)
	router.Use(TimeoutMiddleware)
	router.Use(LoggingMiddleware)
	router.Use(CORSHandler)
	router.Use(RecoverMiddleware)

	protectedRoutes := router.PathPrefix("/").Subrouter()
	protectedRoutes.Use(JWTAuthMiddleware)
	protectedRoutes.HandleFunc("/students", h.GetAllStudents).Methods("GET")
	protectedRoutes.HandleFunc("/students/{id}", h.GetStudent).Methods("GET")
	protectedRoutes.HandleFunc("/students", h.CreateStudent).Methods("POST")
	protectedRoutes.HandleFunc("/students/{id}", h.UpdateStudent).Methods("PUT")
	protectedRoutes.HandleFunc("/students/{id}", h.DeleteStudent).Methods("DELETE")

	router.HandleFunc("/register", h.RegisterUser).Methods("POST")

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	logrus.Info("Starting server on :8080")
	return server.ListenAndServe()
}
