package services

import (
	"context"
	"fuzzy_search/internal"
	"fuzzy_search/internal/repo"
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
	for _, product := range products {
		err := p.repo.Save(ctx, product)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *ProductService) SearchProduct(ctx context.Context, query string) ([]internal.Product, error) {
	return p.repo.Search(ctx, query)
}
