package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
	"log"
	"os"
)

type User struct {
	Id       int
	Username string
	Password string
	Name     string
	Email    string
}

// HandleRequest returns JSON response if HTTP method is GET, otherwise returns error
func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var resp events.APIGatewayProxyResponse
	var err error
	if request.HTTPMethod == "GET" {
		resp = events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       marshall(postgresDB()),
		}
		err = nil
	} else {
		fmt.Println(request.HTTPMethod)
		resp = events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid request",
		}
		err = errors.New("invalid request method")
	}
	return resp, err
}

// Initialize handler
func main() {
	lambda.Start(HandleRequest)
}

// Checks for error
func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// formats users to JSON, then parse to string that is compatible for APIResponse
func marshall(users *[]User) string {
	testVar, err := json.Marshal(users)
	check(err)
	return string(testVar)
}

// Connects to AWS RDS, retrieves user data from table
func postgresDB() *[]User {
	dbUsername := os.Getenv("USERNAME")
	dbPassword := os.Getenv("PASSWORD")
	dbHost := os.Getenv("HOST")
	dbName := os.Getenv("DBNAME")
	dbTableName := os.Getenv("TABLENAME")
	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s/%s", dbUsername, dbPassword, dbHost, dbName)
	db, err := sql.Open("postgres", dataSourceName)
	check(err)
	defer db.Close()
	check(db.Ping())
	fmt.Println("You connected to your database.")
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s;", dbTableName))
	check(err)
	defer rows.Close()
	users := make([]User, 0)
	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.Id, &u.Username, &u.Password, &u.Name, &u.Email) // maintain order
		check(err)
		users = append(users, u)
	}
	check(rows.Err())
	return &users
}
