package service

import (
	"context"
	//"database/sql"
	"errors"
	//"fmt"

	log "github.com/sirupsen/logrus"
)

// Define the Student struct
type Student struct {
	ID        int32  `json:"id"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Course    string `json:"course"`
	Grade     string `json:"grade"`
	CreatedBy string `json:"created_by"`
	CreatedOn string `json:"created_on"`
	UpdatedBy string `json:"updated_by"`
	UpdatedOn string `json:"updated_on"`
}

var (
	ErrFetchingStudent = errors.New("no student with that ID found")
	ErrUpdatingStudent = errors.New("update unsuccessful")
	ErrDeletingStudent = errors.New("could not delete student")
	ErrStudentNotFound = errors.New("not found")
)

type StudentStore interface {
	GetAllStudents(context.Context) ([]Student, error)
	GetStudent(context.Context, int32) (Student, error)
	AddStudent(context.Context, Student) (Student, error)
	UpdateStudent(context.Context, int32, Student) (Student, error)
	DeleteStudent(context.Context, int32) error
	Ping(context.Context) error
	ReadyCheck(ctx context.Context) error
	GetLastInsertedStudentID() (int32, error)
}

type Service struct {
	Store StudentStore
}

func NewService(store StudentStore) *Service {
	return &Service{
		Store: store,
	}
}

func (s *Service) GetAllStudents(ctx context.Context) ([]Student, error) {
	students, err := s.Store.GetAllStudents(ctx)
	if err != nil {
		log.Errorf("Error fetching all students: %s", err.Error())
		return nil, ErrFetchingStudent
	}
	return students, nil
}

func (s *Service) GetStudent(ctx context.Context, userID int32) (Student, error) {
	student, err := s.Store.GetStudent(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrStudentNotFound) {
			log.Warnf("No student found with user ID: %d", userID)
		} else {
			log.Errorf("Error fetching student by user ID: %d, Error: %v", userID, err)
		}
		return Student{}, err
	}
	log.Infof("Successfully retrieved student with user ID: %d", userID)
	return student, nil
}

func (s *Service) GetLastInsertedStudentID() (int32, error) {

	return s.Store.GetLastInsertedStudentID()
}

func (s *Service) AddStudent(ctx context.Context, student Student) (Student, error) {
	log.Error("User type from context in service layer", ctx.Value("userType"))
	student, err := s.Store.AddStudent(ctx, student)
	if err != nil {
		log.Errorf("Failed to insert student with user ID: %d, Error: %v", student.ID, err)
		return Student{}, err
	}
	log.Infof("Successfully added new student with user ID: %d", student.ID)
	return student, nil
}

func (s *Service) UpdateStudent(ctx context.Context, userID int32, student Student) (Student, error) {
	student, err := s.Store.UpdateStudent(ctx, userID, student)
	if err != nil {
		log.Errorf("Failed to update student with user ID: %d, Error: %v", userID, err)
		return Student{}, err
	}
	log.Infof("Successfully updated student with user ID: %d", userID)
	return student, nil
}

func (s *Service) DeleteStudent(ctx context.Context, userID int32) error {
	err := s.Store.DeleteStudent(ctx, userID)
	if err != nil {
		log.Errorf("Failed to delete student with user ID: %d, Error: %v", userID, err)
		return err
	}
	log.Infof("Successfully deleted student with user ID: %d", userID)
	return nil
}

func (s *Service) ReadyCheck(ctx context.Context) error {
	log.Info("Performing readiness check")
	return s.Store.Ping(ctx)
}

func (s *Service) Ping(ctx context.Context) error {
	log.Info("Pinging the database from service layer")
	return s.Store.Ping(ctx)
}
