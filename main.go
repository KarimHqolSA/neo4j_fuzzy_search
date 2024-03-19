package main

import (
	"encoding/json"
	"fmt"
	"fuzzy_search/internal"
	"fuzzy_search/internal/repo"
	"fuzzy_search/internal/services"
	"log"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver *neo4j.DriverWithContext
)

type Query struct {
	Query string `json:"query"`
}

func main() {

	// initialize neo4j driver
	tempDriver, err := neo4j.NewDriverWithContext(
		"bolt://localhost:7689",
		neo4j.BasicAuth("neo4j", "neo4jneo4j", ""))

	if err != nil {
		log.Fatal(err)
	}

	driver = &tempDriver

	http.HandleFunc("/addProducts", handleAddProducts)
	http.HandleFunc("/search", searchHandler)
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	var query Query
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request body: %s", err), http.StatusBadRequest)
		return
	}

	prodRepo := repo.NewNeo4jRepository(*driver)

	productService := services.NewProductService(prodRepo)

	products, err := productService.SearchProduct(r.Context(), query.Query)
	fmt.Println("query.Query", query.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func handleAddProducts(w http.ResponseWriter, r *http.Request) {

	prodRepo := repo.NewNeo4jRepository(*driver)

	productService := services.NewProductService(prodRepo)

	var products []internal.Product
	err := json.NewDecoder(r.Body).Decode(&products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = productService.SaveProduct(r.Context(), products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Products added successfully"))
}
