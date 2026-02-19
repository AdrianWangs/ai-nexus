package main

// cmd entrypoint for product-service

import (
	"flag"
	"log"

	"github.com/AdrianWangs/ai-nexus/pkg/agent"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "Path to config file")
	flag.Parse()

	// Create product store
	productStore := NewProductStore()

	// Run agent with tool registration
	err := agent.RunAgent(*configPath, func(ag *agent.Agent) error {
		// Register tools
		if err := ag.RegisterToolFunc("query_product", "Query products by product_id, name, category, or max_price", productStore.QueryProduct); err != nil {
			return err
		}

		if err := ag.RegisterToolFunc("check_stock", "Check the stock/inventory of a product", productStore.CheckStock); err != nil {
			return err
		}

		if err := ag.RegisterToolFunc("list_categories", "List all product categories", productStore.ListCategories); err != nil {
			return err
		}

		if err := ag.RegisterToolFunc("list_products", "List products, optionally filtered by category", productStore.ListProducts); err != nil {
			return err
		}

		log.Printf("Product service tools registered: query_product, check_stock, list_categories, list_products")
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to run agent: %v", err)
	}
}
