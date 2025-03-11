package models

import "time"

type Invoice struct {
    ID             int       `json:"id"`
    InvoiceNumber  string    `json:"invoice_number"`
    CustomerID     int       `json:"customer_id"`
    IssueDate      string    `json:"issue_date"`
    DueDate        string    `json:"due_date"`
    TotalAmount    float64   `json:"total_amount"`
    Status         string    `json:"status"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}