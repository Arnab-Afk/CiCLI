package generator

import (
	"fmt"
)

type Generator struct {
	// dependencies
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(provider string) error {
	fmt.Printf("Generating pipeline for provider: %s\n", provider)
	// TODO: Load templates and write to .github/workflows
	return nil
}
