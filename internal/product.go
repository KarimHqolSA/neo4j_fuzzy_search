package internal

import "strings"

type Product struct {
	Id                    string
	Title                 string
	Description           string
	Price                 float64
	FullTextWithSpaces    string
	FullTextWithoutSpaces string
}

func (p *Product) CreateIndex() {
	// Concatenate title and description with spaces
	fullTextWithSpaces := p.Title + " " + p.Description

	// Remove spaces from title and description and concatenate
	noSpaces := strings.ReplaceAll(p.Title, " ", "") + strings.ReplaceAll(p.Description, " ", "")

	p.FullTextWithSpaces = fullTextWithSpaces
	p.FullTextWithoutSpaces = noSpaces
}
