package main

import (
	"log"
	"net/http"
	"os"

	"invoice-system/internal/database"
	"invoice-system/internal/handlers"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
    // Initialize database
    database.InitDB()

    // Initialize router
    r := mux.NewRouter()

    // Customer routes
    r.HandleFunc("/api/customers", handlers.GetCustomers).Methods("GET")
    r.HandleFunc("/api/customers", handlers.CreateCustomer).Methods("POST")
    r.HandleFunc("/api/customers/{id}", handlers.GetCustomer).Methods("GET")
    r.HandleFunc("/api/customers/{id}", handlers.UpdateCustomer).Methods("PUT")
    r.HandleFunc("/api/customers/{id}", handlers.DeleteCustomer).Methods("DELETE")

    // Item routes
    r.HandleFunc("/api/items", handlers.GetItems).Methods("GET")
    r.HandleFunc("/api/items", handlers.CreateItem).Methods("POST")
    r.HandleFunc("/api/items/{id}", handlers.GetItem).Methods("GET")
    r.HandleFunc("/api/items/{id}", handlers.UpdateItem).Methods("PUT")
    r.HandleFunc("/api/items/{id}", handlers.DeleteItem).Methods("DELETE")

    // Invoice routes
    r.HandleFunc("/api/invoices", handlers.GetInvoices).Methods("GET")
    r.HandleFunc("/api/invoices", handlers.CreateInvoice).Methods("POST")
    r.HandleFunc("/api/invoices/{id}", handlers.GetInvoice).Methods("GET")
    r.HandleFunc("/api/invoices/{id}", handlers.UpdateInvoice).Methods("PUT")
    r.HandleFunc("/api/invoices/{id}", handlers.DeleteInvoice).Methods("DELETE")
    r.HandleFunc("/api/invoices/{id}/pay", handlers.MarkInvoiceAsPaid).Methods("POST")

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server listening on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}