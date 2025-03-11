package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"invoice-system/internal/database"
	"invoice-system/internal/models"

	"github.com/gorilla/mux"
)

func GetInvoices(w http.ResponseWriter, r *http.Request) {
    status := r.URL.Query().Get("status")
    startDate := r.URL.Query().Get("start_date")
    endDate := r.URL.Query().Get("end_date")

    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit < 1 {
        limit = 10
    }
    offset := (page - 1) * limit

    query := `
        SELECT i.id, i.invoice_number, i.customer_id, i.issue_date, i.due_date, 
            i.total_amount, i.status, i.created_at, i.updated_at
        FROM invoices i
    `
    var args []interface{}
    clauses := []string{}

    if status != "" {
        clauses = append(clauses, "i.status = ?")
        args = append(args, status)
    }
    if startDate != "" {
        clauses = append(clauses, "i.issue_date >= ?")
        args = append(args, startDate)
    }
    if endDate != "" {
        clauses = append(clauses, "i.issue_date <= ?")
        args = append(args, endDate)
    }

    if len(clauses) > 0 {
        query += " WHERE " + joinClauses(clauses, " AND ")
    }

    query += " LIMIT ? OFFSET ?"
    args = append(args, limit, offset)

    rows, err := database.DB.Query(query, args...)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var invoices []models.Invoice
    for rows.Next() {
        var inv models.Invoice
        err := rows.Scan(
            &inv.ID,
            &inv.InvoiceNumber,
            &inv.CustomerID,
            &inv.IssueDate,
            &inv.DueDate,
            &inv.TotalAmount,
            &inv.Status,
            &inv.CreatedAt,
            &inv.UpdatedAt,
        )
        if err != nil {
            http.Error(w, "Scan error", http.StatusInternalServerError)
            return
        }
        invoices = append(invoices, inv)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(invoices)
}

func CreateInvoice(w http.ResponseWriter, r *http.Request) {
    var req struct {
        CustomerID int `json:"customer_id" validate:"required"`
        IssueDate  string `json:"issue_date" validate:"required"`
        DueDate    string `json:"due_date" validate:"required"`
        Items      []struct {
            ItemID   int     `json:"item_id" validate:"required"`
            Quantity int     `json:"quantity" validate:"required,min=1"`
        } `json:"items" validate:"required,min=1"`
    }

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

    tx, err := database.DB.Begin()
    if err != nil {
        http.Error(w, "Transaction error", http.StatusInternalServerError)
        return
    }

    invoiceNumber := "INV-" + strconv.FormatInt(time.Now().UnixNano(), 10)
    totalAmount := 0.0

    // Insert invoice
    res, err := tx.Exec(`
        INSERT INTO invoices (invoice_number, customer_id, issue_date, due_date, total_amount)
        VALUES (?, ?, ?, ?, ?)
    `, invoiceNumber, req.CustomerID, req.IssueDate, req.DueDate, totalAmount)
    if err != nil {
        tx.Rollback()
        http.Error(w, "Invoice creation error", http.StatusInternalServerError)
        return
    }

    invoiceID, _ := res.LastInsertId()

    // Insert invoice items and calculate total
    for _, item := range req.Items {
        var price float64
        err = tx.QueryRow("SELECT price FROM items WHERE id = ?", item.ItemID).Scan(&price)
        if err != nil {
            tx.Rollback()
            http.Error(w, "Item not found", http.StatusBadRequest)
            return
        }

        _, err = tx.Exec(`
            INSERT INTO invoice_items (invoice_id, item_id, quantity, price)
            VALUES (?, ?, ?, ?)
        `, invoiceID, item.ItemID, item.Quantity, price)
        if err != nil {
            tx.Rollback()
            http.Error(w, "Item insertion error", http.StatusInternalServerError)
            return
        }

        totalAmount += price * float64(item.Quantity)
    }

    // Update total amount
    _, err = tx.Exec("UPDATE invoices SET total_amount = ? WHERE id = ?", totalAmount, invoiceID)
    if err != nil {
        tx.Rollback()
        http.Error(w, "Total update error", http.StatusInternalServerError)
        return
    }

    tx.Commit()

    var invoice models.Invoice
    err = database.DB.QueryRow(`
        SELECT id, invoice_number, customer_id, issue_date, due_date, 
            total_amount, status, created_at, updated_at
        FROM invoices
        WHERE id = ?
    `, invoiceID).Scan(
        &invoice.ID,
        &invoice.InvoiceNumber,
        &invoice.CustomerID,
        &invoice.IssueDate,
        &invoice.DueDate,
        &invoice.TotalAmount,
        &invoice.Status,
        &invoice.CreatedAt,
        &invoice.UpdatedAt,
    )

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(invoice)
}

func GetInvoice(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var invoice models.Invoice
    err = database.DB.QueryRow(`
        SELECT id, invoice_number, customer_id, issue_date, due_date, 
            total_amount, status, created_at, updated_at
        FROM invoices
        WHERE id = ?
    `, id).Scan(
        &invoice.ID,
        &invoice.InvoiceNumber,
        &invoice.CustomerID,
        &invoice.IssueDate,
        &invoice.DueDate,
        &invoice.TotalAmount,
        &invoice.Status,
        &invoice.CreatedAt,
        &invoice.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        http.Error(w, "Invoice not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(invoice)
}

func UpdateInvoice(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var req struct {
        IssueDate string `json:"issue_date"`
        DueDate   string `json:"due_date"`
        Status    string `json:"status"`
    }
    err = json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    tx, err := database.DB.Begin()
    if err != nil {
        http.Error(w, "Transaction error", http.StatusInternalServerError)
        return
    }

    query := "UPDATE invoices SET updated_at = ?"
    args := []interface{}{time.Now()}

    if req.IssueDate != "" {
        query += ", issue_date = ?"
        args = append(args, req.IssueDate)
    }
    if req.DueDate != "" {
        query += ", due_date = ?"
        args = append(args, req.DueDate)
    }
    if req.Status != "" {
        query += ", status = ?"
        args = append(args, req.Status)
    }

    query += " WHERE id = ?"
    args = append(args, id)

    _, err = tx.Exec(query, args...)
    if err != nil {
        tx.Rollback()
        http.Error(w, "Update error", http.StatusInternalServerError)
        return
    }

    tx.Commit()

    var updatedInvoice models.Invoice
    err = database.DB.QueryRow(`
        SELECT id, invoice_number, customer_id, issue_date, due_date, 
            total_amount, status, created_at, updated_at
        FROM invoices
        WHERE id = ?
    `, id).Scan(
        &updatedInvoice.ID,
        &updatedInvoice.InvoiceNumber,
        &updatedInvoice.CustomerID,
        &updatedInvoice.IssueDate,
        &updatedInvoice.DueDate,
        &updatedInvoice.TotalAmount,
        &updatedInvoice.Status,
        &updatedInvoice.CreatedAt,
        &updatedInvoice.UpdatedAt,
    )

    json.NewEncoder(w).Encode(updatedInvoice)
}

func DeleteInvoice(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    tx, err := database.DB.Begin()
    if err != nil {
        http.Error(w, "Transaction error", http.StatusInternalServerError)
        return
    }

    _, err = tx.Exec("DELETE FROM invoice_items WHERE invoice_id = ?", id)
    if err != nil {
        tx.Rollback()
        http.Error(w, "Item deletion error", http.StatusInternalServerError)
        return
    }

    _, err = tx.Exec("DELETE FROM invoices WHERE id = ?", id)
    if err != nil {
        tx.Rollback()
        http.Error(w, "Invoice deletion error", http.StatusInternalServerError)
        return
    }

    tx.Commit()
    w.WriteHeader(http.StatusNoContent)
}

func MarkInvoiceAsPaid(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    _, err = database.DB.Exec(`
        UPDATE invoices
        SET status = 'paid', updated_at = ?
        WHERE id = ?
    `, time.Now(), id)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    var invoice models.Invoice
    err = database.DB.QueryRow(`
        SELECT id, invoice_number, customer_id, issue_date, due_date, 
            total_amount, status, created_at, updated_at
        FROM invoices
        WHERE id = ?
    `, id).Scan(
        &invoice.ID,
        &invoice.InvoiceNumber,
        &invoice.CustomerID,
        &invoice.IssueDate,
        &invoice.DueDate,
        &invoice.TotalAmount,
        &invoice.Status,
        &invoice.CreatedAt,
        &invoice.UpdatedAt,
    )

    json.NewEncoder(w).Encode(invoice)
}

func joinClauses(clauses []string, sep string) string {
    return "(" + strings.Join(clauses, sep) + ")"
}