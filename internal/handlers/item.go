package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"invoice-system/internal/database"
	"invoice-system/internal/models"

	"github.com/gorilla/mux"
)

func GetItems(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit < 1 {
        limit = 10
    }
    offset := (page - 1) * limit

    rows, err := database.DB.Query(`
        SELECT id, name, price, created_at, updated_at
        FROM items
        LIMIT ? OFFSET ?
    `, limit, offset)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var items []models.Item
    for rows.Next() {
        var i models.Item
        err := rows.Scan(&i.ID, &i.Name, &i.Price, &i.CreatedAt, &i.UpdatedAt)
        if err != nil {
            http.Error(w, "Scan error", http.StatusInternalServerError)
            return
        }
        items = append(items, i)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(items)
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
    var req models.Item
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    err = validate.Struct(req)
    if err != nil {
        http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
        return
    }

    res, err := database.DB.Exec(`
        INSERT INTO items (name, price)
        VALUES (?, ?)
    `, req.Name, req.Price)
    if err != nil {
        http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    id, _ := res.LastInsertId()
    req.ID = int(id)
    req.CreatedAt = time.Now()
    req.UpdatedAt = time.Now()

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(req)
}

func GetItem(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var item models.Item
    err = database.DB.QueryRow(`
        SELECT id, name, price, created_at, updated_at
        FROM items
        WHERE id = ?
    `, id).Scan(
        &item.ID,
        &item.Name,
        &item.Price,
        &item.CreatedAt,
        &item.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        http.Error(w, "Item not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(item)
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var req models.Item
    err = json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    err = validate.Struct(req)
    if err != nil {
        http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
        return
    }

    _, err = database.DB.Exec(`
        UPDATE items
        SET name = ?, price = ?, updated_at = ?
        WHERE id = ?
    `, req.Name, req.Price, time.Now(), id)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    req.ID = id
    req.UpdatedAt = time.Now()
    json.NewEncoder(w).Encode(req)
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    _, err = database.DB.Exec("DELETE FROM items WHERE id = ?", id)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}