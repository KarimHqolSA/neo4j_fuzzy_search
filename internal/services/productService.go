package services

import (
	"context"
	"fuzzy_search/internal"
	"fuzzy_search/internal/repo"
	"sync"
)

type ProductService struct {
	repo repo.Neo4jRepository
}

func NewProductService(repo repo.Neo4jRepository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (p *ProductService) SaveProduct(ctx context.Context, products []internal.Product) error {
	// Create a channel to receive errors from goroutines
	errCh := make(chan error)

	// Create a WaitGroup to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(len(products))

	// Iterate over the products and launch a goroutine for each product
	for _, product := range products {
		go func(prod internal.Product) {
			defer wg.Done()
			// Save the product and send any error to the error channel
			errCh <- p.repo.Save(ctx, prod)
		}(product)
	}

	// Close the error channel after all goroutines are finished
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Listen for errors from the error channel
	for err := range errCh {
		if err != nil {
			// If there's an error, return it immediately
			return err
		}
	}

	return nil
}

func (p *ProductService) SearchProduct(ctx context.Context, query string, percentage float64) ([]internal.Product, error) {
	return p.repo.Search(ctx, query, percentage)
}
