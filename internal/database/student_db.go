package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	student "GO_Assignment_3/internal/service"

	log "github.com/sirupsen/logrus"
)

var (
	ErrStudentNotFound = errors.New("student not found")
	ErrStudentUpdate   = errors.New("unable to update student details")
	ErrStudentDelete   = errors.New("unable to delete student")
)

type StudentRow struct {
	User_ID   int32          `db:"user_id"`
	Password  sql.NullString `db:"password"`
	Name      sql.NullString `db:"name"`
	Course    sql.NullString `db:"course"`
	Grade     sql.NullString `db:"grade"`
	CreatedBy sql.NullString `db:"created_by"`
	CreatedOn time.Time      `db:"created_on"`
	UpdatedBy sql.NullString `db:"updated_by"`
	UpdatedOn time.Time      `db:"updated_on"`
}

func convertStudentRowToStudent(s StudentRow) student.Student {
	return student.Student{
		ID:        s.User_ID,
		Password:  s.Password.String,
		Name:      s.Name.String,
		Course:    s.Course.String,
		Grade:     s.Grade.String,
		CreatedBy: s.CreatedBy.String,
		CreatedOn: s.CreatedOn.Format(time.RFC3339),
		UpdatedBy: s.UpdatedBy.String,
		UpdatedOn: s.UpdatedOn.Format(time.RFC3339),
	}
}

func (db *Database) GetAllStudents(ctx context.Context) ([]student.Student, error) {
	var students []student.Student
	query := "SELECT * FROM students"

	rows, err := db.Client.QueryContext(ctx, query)
	if err != nil {
		log.Errorf("Error querying students: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var studentRow StudentRow
		if err := rows.Scan(
			&studentRow.User_ID,
			&studentRow.Password,
			&studentRow.Name,
			&studentRow.Course,
			&studentRow.Grade,
			&studentRow.CreatedBy,
			&studentRow.CreatedOn,
			&studentRow.UpdatedBy,
			&studentRow.UpdatedOn,
		); err != nil {
			log.Errorf("Error scanning student row: %s", err.Error())
			return nil, err
		}
		students = append(students, convertStudentRowToStudent(studentRow))
	}

	if err := rows.Err(); err != nil {
		log.Errorf("Error iterating over student rows: %s", err.Error())
		return nil, err
	}

	return students, nil
}

func (d *Database) GetStudent(ctx context.Context, userID int32) (student.Student, error) {
	var studentRow StudentRow
	row := d.Client.QueryRowContext(
		ctx,
		`SELECT user_id, password, name, course, grade, created_by, created_on, updated_by, updated_on
		 FROM students WHERE user_id = ?`,
		userID,
	)

	err := row.Scan(&studentRow.User_ID, &studentRow.Password, &studentRow.Name, &studentRow.Course,
		&studentRow.Grade, &studentRow.CreatedBy, &studentRow.CreatedOn, &studentRow.UpdatedBy, &studentRow.UpdatedOn)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnf("No student found with user ID: %d", userID)
			return student.Student{}, ErrStudentNotFound
		}
		log.Errorf("Error fetching student by user ID: %d, Error: %v", userID, err)
		return student.Student{}, fmt.Errorf("error fetching student by user ID: %w", err)
	}

	log.Infof("Successfully retrieved student with user ID: %d", userID)
	return convertStudentRowToStudent(studentRow), nil
}

func (d *Database) AddStudent(ctx context.Context, student student.Student) (student.Student, error) {
	userType, ok := ctx.Value("userType").(string)
	log.Error("UserType value", userType)
	if !ok {
		log.Error("User type not found in context")
		return student, errors.New("context Error")
	}

	studentRow := StudentRow{
		Password:  sql.NullString{String: student.Password, Valid: true},
		Name:      sql.NullString{String: student.Name, Valid: true},
		Course:    sql.NullString{String: student.Course, Valid: true},
		Grade:     sql.NullString{String: student.Grade, Valid: true},
		CreatedBy: sql.NullString{String: userType, Valid: true},
		UpdatedBy: sql.NullString{String: student.UpdatedBy, Valid: true},
	}

	result, err := d.Client.NamedExecContext(
		ctx,
		`INSERT INTO students (password, name, course, grade, created_by, updated_by) 
		 VALUES (:password, :name, :course, :grade, :created_by, :updated_by)`,
		studentRow,
	)

	if err != nil {
		log.Errorf("Failed to insert student, Error: %v", err)
		return student, fmt.Errorf("failed to insert student: %w", err)
	}

	studentID, err := result.LastInsertId()
	if err != nil {
		log.Errorf("Failed to retrieve last inserted ID, Error: %v", err)
		return student, fmt.Errorf("failed to retrieve last inserted ID: %w", err)
	}

	student.ID = int32(studentID) // Update student ID with the new value
	log.Infof("Successfully added new student with user ID: %d", student.ID)
	return student, nil
}

func (d *Database) UpdateStudent(ctx context.Context, userID int32, student student.Student) (student.Student, error) {
	var exists bool
	err := d.Client.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM students WHERE user_id = ?)`,
		userID,
	).Scan(&exists)

	if err != nil {
		log.Errorf("Failed to check existence of student with user ID: %d, Error: %v", userID, err)
		return student, fmt.Errorf("failed to check existence of student: %w", err)
	}

	if !exists {
		log.Warnf("Attempted to update student with user ID: %d, but no such student exists", userID)
		return student, fmt.Errorf("student with user ID %d does not exist", userID)
	}

	userType, ok := ctx.Value("userType").(string)
	if !ok {
		log.Error("User type not found in context")
		return student, errors.New("context error")
	}

	studentRow := StudentRow{
		User_ID:   userID,
		Password:  sql.NullString{String: student.Password, Valid: true},
		Name:      sql.NullString{String: student.Name, Valid: true},
		Course:    sql.NullString{String: student.Course, Valid: true},
		Grade:     sql.NullString{String: student.Grade, Valid: true},
		UpdatedBy: sql.NullString{String: userType, Valid: true},
	}

	_, err = d.Client.NamedExecContext(
		ctx,
		`UPDATE students SET password = :password, name = :name, course = :course, 
		 grade = :grade, updated_by = :updated_by WHERE user_id = :user_id`,
		studentRow,
	)

	if err != nil {
		log.Errorf("Failed to update student with user ID: %d, Error: %v", userID, err)
		return student, fmt.Errorf("failed to update student: %w", err)
	}

	log.Infof("Successfully updated student with user ID: %d", userID)
	return convertStudentRowToStudent(studentRow), nil
}

func (d *Database) DeleteStudent(ctx context.Context, userID int32) error {
	var exists bool
	err := d.Client.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM students WHERE user_id = ?)`,
		userID,
	).Scan(&exists)

	if err != nil {
		log.Errorf("Failed to check existence of student with user ID: %d, Error: %v", userID, err)
		return fmt.Errorf("failed to check existence of student: %w", err)
	}

	if !exists {
		log.Warnf("Attempted to delete student with user ID: %d, but no such student exists", userID)
		return fmt.Errorf("student with user ID %d does not exist", userID)
	}

	_, err = d.Client.ExecContext(
		ctx,
		`DELETE FROM students WHERE user_id = ?`,
		userID,
	)

	if err != nil {
		log.Errorf("Failed to delete student with user ID: %d, Error: %v", userID, err)
		return fmt.Errorf("failed to delete student from the database: %w", err)
	}

	log.Infof("Successfully deleted student with user ID: %d", userID)
	return nil
}

func (db *Database) GetLastInsertedStudentID() (int32, error) {
	query := `SELECT LAST_INSERT_ID()`

	var lastInsertID int32
	err := db.Client.QueryRow(query).Scan(&lastInsertID)
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

func (d *Database) ReadyCheck(ctx context.Context) error {
	return d.Ping(ctx)
}
