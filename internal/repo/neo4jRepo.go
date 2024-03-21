package repo

import (
	"context"
	"fmt"
	"fuzzy_search/internal"
	"strconv"
	"strings"
	"unicode"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/db"
)

type Neo4jRepository interface {
	Save(context.Context, internal.Product) error
	Search(context.Context, string, float64) ([]internal.Product, error)
}

type neo4jRepo struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jRepository(driver neo4j.DriverWithContext) *neo4jRepo {
	return &neo4jRepo{
		driver: driver,
	}
}

func (n *neo4jRepo) Save(ctx context.Context, product internal.Product) error {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(
			ctx,
			`CREATE (p:Product {id: $id, title: $title, description: $description, price: $price,fullTextWithSpaces:$fullTextWithSpaces, fullTextWithoutSpaces:$fullTextWithoutSpaces })`,
			map[string]interface{}{
				"id":                    product.Id,
				"title":                 product.Title,
				"description":           product.Description,
				"price":                 product.Price,
				"fullTextWithSpaces":    product.FullTextWithSpaces,
				"fullTextWithoutSpaces": product.FullTextWithoutSpaces,
			},
		)

		if err != nil {
			return nil, err
		}
		return nil, nil

	})

	if err != nil {
		return err
	}
	return nil
}

func (n *neo4jRepo) Search(ctx context.Context, query string, percentage float64) ([]internal.Product, error) {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Preprocess the search query to insert spaces between each character
	// modifiedQuery := strings.Join(strings.Split(query, ""), " ")
	// fmt.Println("modified query", modifiedQuery)
	// analyzer := "standard-folding"
	// modifiedQuery = fmt.Sprint(modifiedQuery, "~", percentage)

	analyzer := "standard-folding"

	// query = fmt.Sprint("\"", query, "\"", "~", 100000) // proximity
	// fuzzyQuery := fmt.Sprintf("fullTextWithSpaces: %s~%f", query, percentage)
	percentageStr := strconv.FormatFloat(percentage, 'f', -1, 64)

	fuzzyQuery := "fullTextWithSpaces: " + query + "~" + percentageStr

	result, err := session.Run(
		ctx, `
			CALL db.index.fulltext.queryNodes("product_full_text_index", $query, {analyzer: $analyzer}) YIELD node
			RETURN node.title AS title, node.description AS description, node.price AS price, node.id AS id
		`,
		map[string]interface{}{
			"query":    fuzzyQuery,
			"analyzer": analyzer,
		},
	)
	if err != nil {
		return nil, err
	}
	fuzzyHasResult := false

	products := make([]internal.Product, 0)

	for result.Next(ctx) {
		fuzzyHasResult = true
		record := result.Record()
		product := parseQueryResult(record)
		products = append(products, product)
	}
	wildHasResult := false
	if !fuzzyHasResult {
		wildCardQuery := "fullTextWithoutSpaces: " + splitQuery(query, percentage)
		result, err = session.Run(
			ctx, `
					CALL db.index.fulltext.queryNodes("product_full_text_index", $query, {analyzer: $analyzer}) YIELD node
					RETURN node.title AS title, node.description AS description, node.price AS price, node.id AS id
				`,
			map[string]interface{}{
				"query":    wildCardQuery,
				"analyzer": analyzer,
			},
		)
		if err != nil {
			return nil, err
		}

		for result.Next(ctx) {
			wildHasResult = true
			record := result.Record()
			product := parseQueryResult(record)
			products = append(products, product)
		}
		fmt.Println("wildHasResult", wildHasResult)
	}
	// if !wildHasResult {
	// 	orQuery := splitQuery(query, percentage)
	// 	result, err = session.Run(
	// 		ctx, `
	// 				CALL db.index.fulltext.queryNodes("product_full_text_index", $query, {analyzer: $analyzer}) YIELD node
	// 				RETURN node.title AS title, node.description AS description, node.price AS price, node.id AS id
	// 			`,
	// 		map[string]interface{}{
	// 			"query":    orQuery,
	// 			"analyzer": analyzer,
	// 		},
	// 	)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	for result.Next(ctx) {
	// 		record := result.Record()
	// 		product := parseQueryResult(record)
	// 		products = append(products, product)
	// 	}
	// }

	return products, nil
}

func splitQuery(query string, percentage float64) string {
	// Split the query string on spaces or special characters
	tokens := strings.FieldsFunc(query, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	if len(tokens) == 1 {
		return query + "*"
	}
	for i := 0; i < len(tokens); i++ {
		tokens[i] = tokens[i] + "*"
	}
	finalQuery := strings.Join(tokens, " OR ")
	return finalQuery
}

func parseQueryResult(record *db.Record) internal.Product {
	productNode := record.Values
	// product := internal.NewProduct(productNode[3].(string), productNode[0].(string), productNode[1].(string), productNode[2].(float64))
	product := internal.Product{
		Id:          productNode[3].(string),
		Title:       productNode[0].(string),
		Description: productNode[1].(string),
		Price:       productNode[2].(float64),
	}

	return product
}

/*
// this returns a good result
	result, err := session.Run(
		ctx, `
			CALL db.index.fulltext.queryNodes("product_full_text_index", $query) YIELD node
			RETURN node.title AS title, node.description AS description, node.price AS price, node.id AS id
		`,
		map[string]interface{}{
			"query": query,
		},
	)


*/
