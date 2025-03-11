package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"invoice-system/internal/database"
	"invoice-system/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

var validate = validator.New()

func GetCustomers(w http.ResponseWriter, r *http.Request) {
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
        SELECT id, name, email, address, created_at, updated_at
        FROM customers
        LIMIT ? OFFSET ?
    `, limit, offset)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var customers []models.Customer
    for rows.Next() {
        var c models.Customer
        err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Address, &c.CreatedAt, &c.UpdatedAt)
        if err != nil {
            http.Error(w, "Scan error", http.StatusInternalServerError)
            return
        }
        customers = append(customers, c)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(customers)
}

func CreateCustomer(w http.ResponseWriter, r *http.Request) {
    var req models.Customer
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
        INSERT INTO customers (name, email, address)
        VALUES (?, ?, ?)
    `, req.Name, req.Email, req.Address)
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

func GetCustomer(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var customer models.Customer
    err = database.DB.QueryRow(`
        SELECT id, name, email, address, created_at, updated_at
        FROM customers
        WHERE id = ?
    `, id).Scan(
        &customer.ID,
        &customer.Name,
        &customer.Email,
        &customer.Address,
        &customer.CreatedAt,
        &customer.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        http.Error(w, "Customer not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(customer)
}

func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var req models.Customer
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
        UPDATE customers
        SET name = ?, email = ?, address = ?, updated_at = ?
        WHERE id = ?
    `, req.Name, req.Email, req.Address, time.Now(), id)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    req.ID = id
    req.UpdatedAt = time.Now()
    json.NewEncoder(w).Encode(req)
}

func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    _, err = database.DB.Exec("DELETE FROM customers WHERE id = ?", id)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}