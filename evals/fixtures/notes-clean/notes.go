// EVAL FIXTURE — intentionally sound. Checks that the audit reports a clean result.
package notes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

const maxBody = 64 << 10

type Server struct {
	db *sql.DB
}

type noteInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type noteView struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	UpdatedAt string `json:"updatedAt"`
}

// Routes are mounted behind RequireSession, which rejects requests with no valid session.
func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /notes", s.list)
	mux.HandleFunc("POST /notes", s.create)
	mux.HandleFunc("PUT /notes/{id}", s.update)
	mux.HandleFunc("DELETE /notes/{id}", s.remove)
}

func decode(w http.ResponseWriter, r *http.Request) (noteInput, error) {
	var in noteInput
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return in, err
	}
	if in.Title == "" || len(in.Title) > 200 {
		return in, errors.New("title must be between 1 and 200 characters")
	}
	return in, nil
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	user := sessionUser(r.Context())
	rows, err := s.db.QueryContext(r.Context(),
		`SELECT id, title, body, updated_at FROM notes WHERE owner_id = $1 ORDER BY updated_at DESC LIMIT 100`,
		user.ID)
	if err != nil {
		http.Error(w, "could not load notes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	notes := []noteView{}
	for rows.Next() {
		var n noteView
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.UpdatedAt); err != nil {
			http.Error(w, "could not load notes", http.StatusInternalServerError)
			return
		}
		notes = append(notes, n)
	}
	json.NewEncoder(w).Encode(notes)
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	user := sessionUser(r.Context())
	in, err := decode(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var n noteView
	err = s.db.QueryRowContext(r.Context(),
		`INSERT INTO notes (id, owner_id, title, body, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, now(), now())
		 RETURNING id, title, body, updated_at`,
		user.ID, in.Title, in.Body).Scan(&n.ID, &n.Title, &n.Body, &n.UpdatedAt)
	if err != nil {
		http.Error(w, "could not save note", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(n)
}

func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	user := sessionUser(r.Context())
	in, err := decode(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var n noteView
	err = s.db.QueryRowContext(r.Context(),
		`UPDATE notes SET title = $1, body = $2, updated_at = now()
		 WHERE id = $3 AND owner_id = $4
		 RETURNING id, title, body, updated_at`,
		in.Title, in.Body, r.PathValue("id"), user.ID).Scan(&n.ID, &n.Title, &n.Body, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "could not save note", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(n)
}

func (s *Server) remove(w http.ResponseWriter, r *http.Request) {
	user := sessionUser(r.Context())
	res, err := s.db.ExecContext(r.Context(),
		`DELETE FROM notes WHERE id = $1 AND owner_id = $2`, r.PathValue("id"), user.ID)
	if err != nil {
		http.Error(w, "could not delete note", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
