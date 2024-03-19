package repo

import (
	"context"
	"fuzzy_search/internal"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jRepository interface {
	Save(context.Context, internal.Product) error
	Search(context.Context, string) ([]internal.Product, error)
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
			`CREATE (p:Product {id: $id, title: $title, description: $description, price: $price})`,
			map[string]interface{}{
				"id":          product.Id,
				"title":       product.Title,
				"description": product.Description,
				"price":       product.Price,
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

func (n *neo4jRepo) Search(ctx context.Context, query string) ([]internal.Product, error) {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// percentage := 0.115556789123
	// result, err := session.Run(
	// 	ctx, `
	// 			MATCH
	// 				(p:Product)
	// 			with
	// 				p,
	// 				apoc.text.levenshteinSimilarity(p.title, $query ) as sc1,
	// 				apoc.text.levenshteinSimilarity(p.description, $query ) as sc2
	// 			where sc1 >= $percentage or sc2 >= $percentage
	// 			RETURN p
	// 			order by sc1 desc, sc2 desc
	// 	`,
	// 	map[string]interface{}{
	// 		"query":      query,
	// 		"percentage": percentage,
	// 	},
	// )

	if err != nil {
		return nil, err
	}

	products := make([]internal.Product, 0)

	for result.Next(ctx) {
		record := result.Record()
		productNode := record.Values[0].(neo4j.Node)
		product := internal.Product{
			Id:          productNode.Props["id"].(int64),
			Title:       productNode.Props["title"].(string),
			Description: productNode.Props["description"].(string),
			Price:       productNode.Props["price"].(int64),
		}
		products = append(products, product)
	}
	return products, nil
}
