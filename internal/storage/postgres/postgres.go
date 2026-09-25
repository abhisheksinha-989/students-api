package postgres

import (
	"database/sql"
	"fmt"

	"github.com/abhisheksinha-989/students-api/internal/storage"
	"github.com/abhisheksinha-989/students-api/internal/types"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	Db *sql.DB
}

func New(databaseURL string) (*Postgres, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS students (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	age INTEGER NOT NULL
	);`

	if _, err := db.Exec(createTableQuery); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Postgres{Db: db}, nil
}

// close shuts the connection down main.go calls this with defer
func (s *Postgres) Close() error {
	return s.Db.Close()
}

// create
func (s *Postgres) CreateStudent(name string, email string, age int) (int64, error) {
	query := `INSERT INTO students (name, email, age) VALUES($1,$2,$3) RETURNING id`

	var id int64
	err := s.Db.QueryRow(query, name, email, age).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create student: %w", err)
	}
	return id, nil
}

// Read
func (s *Postgres) GetStudentById(id int64) ([]types.Student, error) {
	query := `SELECT id, name, email, age FROM students ORDER BY id`

	rows, err := s.Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []types.Student

	for rows.Next() {
		var student types.Student
		if err := rows.Scan(&student.Id, &student.Name, &student.Email, &student.Age); err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

// update
func (s *Postgres) UpdateStudent(id int64, name string, email string, age int) error {
	query := `UPDATE students SET name = $1, email = $2, age = $3 WHERE id = $4`

	result, err := s.Db.Exec(query, name, email, age, id)
	if err != nil {
		return fmt.Errorf("failed to update student: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return storage.ErrStudentNotFound
	}
	return nil
}

// DELETE
func (s *Postgres) DeleteStudent(id int64) error {
	query := `DELETE FROM students WHERE id = $1`

	result, err := s.Db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return storage.ErrStudentNotFound
	}
	return nil
}
