# Invoice System API

## Requirements

- Go 1.20.x
- MySQL 5.7
- Postman

## Setup

1. **Create `.env` file:**

   ```env
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=
   DB_NAME=invoice_system
   ```

2. Initialize

   - go mod init invoice-system
   - # Install all dependencies from go.mod
     go mod tidy

3. Import Database Schema

   - mysql -u root -p invoice_system < migrations/schema.sql
   - use your mysql username and password

4. Run the Application
   - go run cmd/main.go
