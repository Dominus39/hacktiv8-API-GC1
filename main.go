package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

// customer struct
type customer struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

var db *sql.DB

func ConnectDB() {
	var err error
	db_name := "postgres"
	db_user := "postgres.muemjkqukaslsdhybtir"
	db_pass := "adminklewear5"
	db_host := "aws-0-ap-southeast-1.pooler.supabase.com"

	db, err = sql.Open("postgres", fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=require", db_name, db_user, db_pass, db_host))
	if err != nil {
		log.Print("Error connecting to the database: ", err)
		log.Fatal(err)
	}

	// Check the connection
	if err = db.Ping(); err != nil {
		log.Print("Error pinging the database: ", err)
		log.Fatal(err)
	}

	fmt.Println("Connected to the database successfully!")
}

func getAllCustomers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Query the database
	rows, err := db.Query("SELECT id, name, email, phone FROM Customers WHERE deleted_at IS NULL")
	if err != nil {
		http.Error(w, "Failed to query users", http.StatusInternalServerError)
		log.Println("Database query error:", err)
		return
	}
	defer rows.Close()

	// Parse the rows into a slice of users
	var customers []customer
	for rows.Next() {
		var c customer
		if err := rows.Scan(&c.Id, &c.Name, &c.Email, &c.Phone); err != nil {
			http.Error(w, "Failed to parse customer data", http.StatusInternalServerError)
			log.Println("Row scan error:", err)
			return
		}
		customers = append(customers, c)
	}

	// Handle errors from iteration
	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading rows", http.StatusInternalServerError)
		log.Println("Row iteration error:", err)
		return
	}

	// Respond with the users in JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(customers); err != nil {
		log.Println("JSON encoding error:", err)
	}
}

func getCustomerDetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	// Query the database
	var c customer
	err := db.QueryRow("SELECT id, name, email, phone FROM Customers WHERE id = $1 AND deleted_at IS NULL", id).Scan(&c.Id, &c.Name, &c.Email, &c.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the case where no customer is found
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			// Handle other database errors
			http.Error(w, "Failed to query customer", http.StatusInternalServerError)
			log.Println("Database query error:", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

func addCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var customers customer
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&customers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error : %v", err)
		return
	}

	defer r.Body.Close()

	// Insert the new customer into the database
	query := `INSERT INTO Customers (name, email, phone) VALUES ($1, $2, $3) RETURNING id`
	var newCustomerID int
	err = db.QueryRow(query, customers.Name, customers.Email, customers.Phone).Scan(&newCustomerID)
	if err != nil {
		http.Error(w, "Failed to insert customer", http.StatusInternalServerError)
		log.Println("Database insert error:", err)
		return
	}

	// Set the ID of the newly created customer
	customers.Id = newCustomerID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(customers)
}

func editCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	var updatedCustomer customer

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&updatedCustomer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error : %v", err)
		return
	}

	defer r.Body.Close()

	// Query the database to check if the customer exists
	var existingCustomer customer
	query := `SELECT id, name, email, phone FROM Customers WHERE id = $1 AND deleted_at IS NULL`
	err = db.QueryRow(query, id).Scan(&existingCustomer.Id, &existingCustomer.Name, &existingCustomer.Email, &existingCustomer.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			// If the customer does not exist, return an error message
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			// Any other database error
			http.Error(w, "Failed to fetch customer data", http.StatusInternalServerError)
		}
		log.Println("Database query error:", err)
		return
	}

	// update the customer in the database
	updateQuery := `UPDATE Customers SET 
		name = COALESCE(NULLIF($1, ''), name), 
		email = COALESCE(NULLIF($2, ''), email), 
		phone = COALESCE(NULLIF($3, ''), phone) 
		WHERE id = $4`
	_, err = db.Exec(updateQuery, updatedCustomer.Name, updatedCustomer.Email, updatedCustomer.Phone, id)
	if err != nil {
		http.Error(w, "Failed to update customer", http.StatusInternalServerError)
		log.Println("Database update error:", err)
		return
	}

	// Set the response header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Respond with a success message and the updated customer data
	response := map[string]interface{}{
		"message":  "Customer updated successfully",
		"customer": updatedCustomer,
	}
	json.NewEncoder(w).Encode(response)
}

func deleteCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get the customer ID from the URL parameters
	id := ps.ByName("id")

	// Query the database to check if the customer exists and is not already deleted (deleted_at is NULL)
	var existingCustomer customer
	query := `SELECT id, name, email, phone FROM Customers WHERE id = $1 AND deleted_at IS NULL`
	err := db.QueryRow(query, id).Scan(&existingCustomer.Id, &existingCustomer.Name, &existingCustomer.Email, &existingCustomer.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			// If the customer is not found or already deleted
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			// Any other database error
			http.Error(w, "Failed to fetch customer data", http.StatusInternalServerError)
		}
		log.Println("Database query error:", err)
		return
	}

	// Proceed to mark the customer as deleted by setting deleted_at to the current timestamp
	updateQuery := `UPDATE Customers SET deleted_at = NOW() WHERE id = $1`
	_, err = db.Exec(updateQuery, id)
	if err != nil {
		http.Error(w, "Failed to delete customer", http.StatusInternalServerError)
		log.Println("Database update error:", err)
		return
	}

	// Set the response header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Prepare the response with the deleted customer data and a success message
	response := map[string]interface{}{
		"message":  "Customer deleted successfully",
		"customer": existingCustomer, // Return the deleted customer data
	}

	// Send the response back to the client
	json.NewEncoder(w).Encode(response)
}

func main() {
	ConnectDB()
	router := httprouter.New()
	router.GET("/customers", getAllCustomers)
	router.GET("/customers/:id", getCustomerDetail)
	router.POST("/customers", addCustomer)
	router.PUT("/customers/:id", editCustomer)
	router.DELETE("/customers/:id", deleteCustomer)

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, c interface{}) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Got panic error : %v", c)
	}

	http.ListenAndServe(":8080", router)
}
