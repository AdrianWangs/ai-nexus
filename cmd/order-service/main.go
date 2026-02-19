package main

// cmd entrypoint for order-service

import (
	"flag"
	"log"

	"github.com/AdrianWangs/ai-nexus/pkg/agent"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "Path to config file")
	flag.Parse()

	// Create order store
	orderStore := NewOrderStore()

	// Run agent with tool registration
	err := agent.RunAgent(*configPath, func(ag *agent.Agent) error {
		// Register tools
		if err := ag.RegisterToolFunc("query_order", "Query orders by order_id, user_id, username, or status", orderStore.QueryOrder); err != nil {
			return err
		}

		if err := ag.RegisterToolFunc("create_order", "Create a new order", orderStore.CreateOrder); err != nil {
			return err
		}

		if err := ag.RegisterToolFunc("update_order_status", "Update an order's status", orderStore.UpdateOrderStatus); err != nil {
			return err
		}

		if err := ag.RegisterToolFunc("list_orders", "List all orders", orderStore.ListOrders); err != nil {
			return err
		}

		log.Printf("Order service tools registered: query_order, create_order, update_order_status, list_orders")
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to run agent: %v", err)
	}
}
