package main

import (
	"devarminas/project-name/server"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Todo struct {
	ID          string `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	Description string `json:"description"`
	IsDone      bool   `json:"is_done"`
}

type createTodoRequest struct {
	Description string `json:"description"`
}

type updateTodoRequest struct {
	Description *string `json:"description,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

func main() {
	r := server.NewRouter()
	r.Get("/todos", func(w http.ResponseWriter, r *http.Request) {
		dsn := "host=localhost user=postgres password=example dbname=project-name port=5432"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			http.Error(w, "failed to connect to database", http.StatusInternalServerError)
			return
		}

		var todos []Todo
		if err := db.Find(&todos).Error; err != nil {
			http.Error(w, "failed to fetch todos", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todos)

	})

	r.Get("/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := server.PathParam(r, "id")
		if id == "" {
			http.Error(w, "missing todo id", http.StatusBadRequest)
			return
		}

		dsn := "host=localhost user=postgres password=example dbname=project-name port=5432"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			http.Error(w, "failed to connect to database", http.StatusInternalServerError)
			return
		}

		var todo Todo
		if err := db.First(&todo, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, "failed to fetch todo", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todo)
	})

	r.Delete("/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := server.PathParam(r, "id")
		if id == "" {
			http.Error(w, "missing todo id", http.StatusBadRequest)
			return
		}

		dsn := "host=localhost user=postgres password=example dbname=project-name port=5432"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			http.Error(w, "failed to connect to database", http.StatusInternalServerError)
			return
		}

		result := db.Delete(&Todo{}, "id = ?", id)
		if result.Error != nil {
			http.Error(w, "failed to delete todo", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	r.Post("/todos", func(w http.ResponseWriter, r *http.Request) {
		dsn := "host=localhost user=postgres password=example dbname=project-name port=5432"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			http.Error(w, "failed to connect to database", http.StatusInternalServerError)
			return
		}

		var reqBody createTodoRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if reqBody.Description == "" {
			http.Error(w, "description is required", http.StatusBadRequest)
			return
		}

		todo := Todo{Description: reqBody.Description}

		if err := db.Create(&todo).Error; err != nil {
			http.Error(w, "failed to create todo", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(todo)

	})

	r.Put("/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := server.PathParam(r, "id")
		if id == "" {
			http.Error(w, "missing todo id", http.StatusBadRequest)
			return
		}

		dsn := "host=localhost user=postgres password=example dbname=project-name port=5432"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			http.Error(w, "failed to connect to database", http.StatusInternalServerError)
			return
		}

		var input updateTodoRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if input.Description == nil && input.Done == nil {
			http.Error(w, "nothing to update", http.StatusBadRequest)
			return
		}

		var todo Todo
		if err := db.First(&todo, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, "failed to fetch todo", http.StatusInternalServerError)
			return
		}

		updates := make(map[string]interface{})
		if input.Description != nil {
			if *input.Description == "" {
				http.Error(w, "description cannot be empty", http.StatusBadRequest)
				return
			}
			updates["description"] = *input.Description
		}

		if input.Done != nil {
			updates["is_done"] = *input.Done
		}

		if err := db.Model(&todo).Updates(updates).Error; err != nil {
			http.Error(w, "failed to update todo", http.StatusInternalServerError)
			return
		}

		if input.Description != nil {
			todo.Description = *input.Description
		}
		if input.Done != nil {
			todo.IsDone = *input.Done
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todo)

	})
	http.ListenAndServe(":3000", r)
}
