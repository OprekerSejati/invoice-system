The architecture NOTE
The architecture used in this implementation is MVC (Model-View-Controller) , adapted for a Go-based REST API

1. Model
   Location : internal/models/
   Purpose :
   Defines the data structures (Customer, Item, Invoice).
   Maps directly to database tables (e.g., invoices, invoice_items).
   Handles data validation (via github.com/go-playground/validator).
2. Controller
   Location : internal/handlers/
   Purpose :
   Processes HTTP requests/responses (e.g., CreateCustomer, GetInvoice).
   Validates input, calls business logic, and returns formatted JSON.
   Implements separation of concerns (e.g., customer.go handles all customer-related endpoints).
3. "View" (Implicit)
   Implementation :
   Since this is an API, the "view" is replaced by JSON responses.
   Controllers directly serialize Go structs to JSON (e.g., json.NewEncoder(w).Encode(customer)).
4. Database Layer
   Location : internal/database/
   Purpose :
   Manages database connections (InitDB).
   Handles transactions and direct SQL queries (e.g., tx.Exec(...)).
   Abstracts database operations from handlers (separation of concerns).

Key Design Choices

1. Separation of Concerns :
   Handlers focus on HTTP logic.
   Models define data structures and validation.
   Database layer handles persistence.
2. Layered Architecture :
   Presentation Layer : HTTP handlers (controllers).
   Business Logic Layer : Embedded in handlers (e.g., validation, transaction logic).
   Data Layer : Database operations.
3. Go Idioms :
   Uses standard net/http and Gorilla Mux for routing.
   Avoids over-engineering (common in Go projects).
