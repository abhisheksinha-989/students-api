package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/abhisheksinha-989/students-api/internal/storage"
	"github.com/abhisheksinha-989/students-api/internal/types"
	"github.com/abhisheksinha-989/students-api/internal/utils/response"
	"github.com/go-playground/validator"
)

var validate = validator.New()

func decodeAndValidate(r *http.Request) (types.StudentRequest, error) {
	var req types.StudentRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return req, fmt.Errorf("request body is empty")
		}
		return req, fmt.Errorf("invalid JSON %w", err)
	}

	if err := validate.Struct(req); err != nil {
		return req, err
	}
	return req, nil
}

func parseId(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id %q id must be a number", idStr)
	}
	return id, nil
}

// create
// Route: POST /api/students
func Create(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("creating a student")

		req, err := decodeAndValidate(r)
		if err != nil {
			var validationErrs validator.ValidationErrors
			if errors.As(err, &validationErrs) {
				response.WriteJson(w, http.StatusBadRequest, response.ValidationErrors(validationErrs))
				return
			}
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		id, err := storage.CreateStudent(req.Name, req.Email, req.Age)
		if err != nil {
			slog.Error("failed to create student", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		slog.Info("student created", slog.Int64("id", id))
		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": id})
	}
}

// Read One
// Router: GET api/students/{id}
func GetById(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseId(r)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		slog.Info("getting student", slog.Int64("id", id))

		student, err := store.GetStudentById(id)
		if err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}
			slog.Error("error getting student", slog.Int64("id", id), slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		response.WriteJson(w, http.StatusOK, student)
	}
}

// Read All
// Route: GET /api/students
func GetList(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("getting all students")

		students, err := storage.GetStudents()
		if err != nil {
			slog.Error("failed to get students", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		response.WriteJson(w, http.StatusOK, students)
	}
}

// Update
// Route: PUT /api/students/{id}
func Update(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("updating a student")

		id, err := parseId(r)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		req, err := decodeAndValidate(r)
		if err != nil {
			var validationErrs validator.ValidationErrors
			if errors.As(err, &validationErrs) {
				response.WriteJson(w, http.StatusBadRequest, response.ValidationErrors(validationErrs))
				return
			}
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		err = store.UpdateStudent(id, req.Name, req.Email, req.Age)
		if err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}

			slog.Error("failed to update student", slog.Int64("id", id), slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return

		}

		slog.Info("student updated", slog.Int64("id", id))
		response.WriteJson(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("student %d updated successfully", id)})
	}
}

// DELETE
// Route: DELETE /api/students/{id}
func Delete(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("deleting a student")

		id, err := parseId(r)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		err = store.DeleteStudent(id)
		if err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return 
			}

			slog.Error("failed to delete student", slog.Int64("id",id), slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return 
		}
		slog.Info("student deleted", slog.Int64("id",id))
		response.WriteJson(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("student %d deleted sucessfully")})
	}
}
