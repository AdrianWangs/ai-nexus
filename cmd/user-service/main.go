package main

// cmd entrypoint for user-service

import (
	"flag"
	"log"

	"github.com/AdrianWangs/ai-nexus/pkg/agent"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "Path to config file")
	flag.Parse()

	// Create user store
	userStore := NewUserStore()

	// Run agent with tool registration
	err := agent.RunAgent(*configPath, func(ag *agent.Agent) error {
		// Register tools
		if err := ag.RegisterToolFunc("query_user", "Query user by username, user_id, or name", userStore.QueryUser); err != nil {
			return err
		}

		if err := ag.RegisterToolFunc("create_user", "Create a new user account", userStore.CreateUser); err != nil {
			return err
		}

		if err := ag.RegisterToolFunc("list_users", "List all users", userStore.ListUsers); err != nil {
			return err
		}

		log.Printf("User service tools registered: query_user, create_user, list_users")
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to run agent: %v", err)
	}
}
