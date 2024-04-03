# MANAGE PLANS MIGRATE DATA

### MIGRATE MANAGE PLANS
- First Install GO
- Change in 73 line database connection:
  ```
  db, err := sqlx.Connect("postgres", "user=svq dbname=services_quotation sslmode=disable password=svq host=localhost")
  ```
- Install Dependences
  ```
  go mod tidy
  ```
- Run importer
  ```
  go run plans.go
  ```

### MIGRATE PRO RATED
- First Install GO
- Change in 75 line database connection:
  ```
  db, err := sqlx.Connect("postgres", "user=svq dbname=services_quotation sslmode=disable password=svq host=localhost")
  ```
- Install Dependences
  ```
  go mod tidy
  ```
- Run importer
  ```
  go run import.go
  ```
