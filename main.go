package main

import (
	"encoding/json"
	"fmt"
	"fuzzy_search/internal"
	"fuzzy_search/internal/repo"
	"fuzzy_search/internal/services"
	"log"
	"net/http"

	"github.com/tealeg/xlsx"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver *neo4j.DriverWithContext
)

type Query struct {
	Query      string  `json:"query"`
	Percentage float64 `json:"percentage"`
	Analyzer   string  `json:"analyzer"`
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
	http.HandleFunc("/parse-xls", handleParseXLS)
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

	products, err := productService.SearchProduct(r.Context(), query.Query, query.Percentage)
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
	for i := 0; i < len(products); i++ {
		products[i].CreateIndex()
	}

	err = productService.SaveProduct(r.Context(), products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Products added successfully"))
}

func handleParseXLS(w http.ResponseWriter, r *http.Request) {
	// Open the XLS file
	file, err := xlsx.OpenFile("products.xlsx")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open XLS file: %s", err), http.StatusInternalServerError)
		return
	}

	// Read data from the XLS file
	sheet := file.Sheets[0]

	var products []internal.Product
	price := 10554.47
	// Parse rows into Product struct
	for r, row := range sheet.Rows {
		if r == 0 {
			continue
		}
		var id string
		var desc string
		var name string
		price++
		for i, cell := range row.Cells {
			switch i {
			case 0:
				id = cell.Value
			case 1:
				name = cell.Value
			case 2:
				desc = cell.Value
			}
		}

		// product := internal.NewProduct(id, name, desc, price)
		product := internal.Product{
			Id:          id,
			Title:       name,
			Description: desc,
			Price:       price,
		}

		products = append(products, product)
	}

	for i := 0; i < len(products); i++ {
		products[i].CreateIndex()
	}

	prodRepo := repo.NewNeo4jRepository(*driver)

	productService := services.NewProductService(prodRepo)

	err = productService.SaveProduct(r.Context(), products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Products parsed successfully from XLS file"))
}
