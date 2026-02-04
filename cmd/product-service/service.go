package main

// Product service demo implementation.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Product represents a product in the catalog.
type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	Tags        []string `json:"tags,omitempty"`
}

// ProductStore is an in-memory product store.
type ProductStore struct {
	mu       sync.RWMutex
	products map[string]*Product
}

// NewProductStore creates a new product store with sample data.
func NewProductStore() *ProductStore {
	store := &ProductStore{
		products: make(map[string]*Product),
	}

	// Add sample products
	sampleProducts := []*Product{
		{
			ID:          "prod-001",
			Name:        "iPhone 15",
			Description: "Apple iPhone 15，A16仿生芯片，超瓷晶面板",
			Category:    "手机",
			Price:       7999.00,
			Stock:       100,
			Tags:        []string{"Apple", "智能手机", "5G"},
		},
		{
			ID:          "prod-002",
			Name:        "MacBook Pro",
			Description: "14英寸MacBook Pro，M3 Pro芯片，18GB内存",
			Category:    "电脑",
			Price:       14999.00,
			Stock:       50,
			Tags:        []string{"Apple", "笔记本", "办公"},
		},
		{
			ID:          "prod-003",
			Name:        "AirPods Pro",
			Description: "AirPods Pro 第二代，主动降噪，空间音频",
			Category:    "配件",
			Price:       1999.00,
			Stock:       200,
			Tags:        []string{"Apple", "耳机", "无线"},
		},
		{
			ID:          "prod-004",
			Name:        "iPad Air",
			Description: "10.9英寸iPad Air，M1芯片，支持Apple Pencil",
			Category:    "平板",
			Price:       4999.00,
			Stock:       80,
			Tags:        []string{"Apple", "平板电脑", "创作"},
		},
		{
			ID:          "prod-005",
			Name:        "Apple Watch Ultra",
			Description: "Apple Watch Ultra 2，钛金属表壳，精准GPS",
			Category:    "穿戴",
			Price:       6499.00,
			Stock:       30,
			Tags:        []string{"Apple", "智能手表", "运动"},
		},
	}

	for _, product := range sampleProducts {
		store.products[product.ID] = product
	}

	return store
}

// QueryProductArgs are the arguments for querying products.
type QueryProductArgs struct {
	ProductID string  `json:"product_id,omitempty" description:"Product ID to search for"`
	Name      string  `json:"name,omitempty" description:"Product name to search for (partial match)"`
	Category  string  `json:"category,omitempty" description:"Category to filter by"`
	MaxPrice  float64 `json:"max_price,omitempty" description:"Maximum price filter"`
}

// QueryProduct queries products by various criteria.
func (s *ProductStore) QueryProduct(ctx context.Context, args *QueryProductArgs) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Query by product ID
	if args.ProductID != "" {
		if product, ok := s.products[args.ProductID]; ok {
			return s.formatProduct(product), nil
		}
		return "未找到商品ID为 " + args.ProductID + " 的商品", nil
	}

	// Filter by criteria
	var results []*Product
	for _, product := range s.products {
		match := true

		if args.Name != "" && !strings.Contains(strings.ToLower(product.Name), strings.ToLower(args.Name)) {
			match = false
		}
		if args.Category != "" && product.Category != args.Category {
			match = false
		}
		if args.MaxPrice > 0 && product.Price > args.MaxPrice {
			match = false
		}

		if match {
			results = append(results, product)
		}
	}

	if len(results) == 0 {
		return "未找到符合条件的商品", nil
	}

	return s.formatProducts(results), nil
}

// CheckStockArgs are the arguments for checking stock.
type CheckStockArgs struct {
	ProductID string `json:"product_id" description:"Product ID to check stock for"`
}

// CheckStock checks the stock of a product.
func (s *ProductStore) CheckStock(ctx context.Context, args *CheckStockArgs) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, ok := s.products[args.ProductID]
	if !ok {
		return "未找到商品ID为 " + args.ProductID + " 的商品", nil
	}

	status := "有货"
	if product.Stock == 0 {
		status = "缺货"
	} else if product.Stock < 10 {
		status = "库存紧张"
	}

	return fmt.Sprintf("商品：%s\n当前库存：%d\n状态：%s", product.Name, product.Stock, status), nil
}

// ListCategoriesArgs are the arguments for listing categories.
type ListCategoriesArgs struct{}

// ListCategories lists all product categories.
func (s *ProductStore) ListCategories(ctx context.Context, args *ListCategoriesArgs) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	categories := make(map[string]int)
	for _, product := range s.products {
		categories[product.Category]++
	}

	var result strings.Builder
	result.WriteString("商品分类：\n")
	for category, count := range categories {
		result.WriteString(fmt.Sprintf("- %s: %d 件商品\n", category, count))
	}

	return result.String(), nil
}

// ListProductsArgs are the arguments for listing products.
type ListProductsArgs struct {
	Category string `json:"category,omitempty" description:"Category to filter by"`
	Limit    int    `json:"limit,omitempty" description:"Maximum number of products to return"`
}

// ListProducts lists products.
func (s *ProductStore) ListProducts(ctx context.Context, args *ListProductsArgs) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := args.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var products []*Product
	count := 0
	for _, product := range s.products {
		if args.Category != "" && product.Category != args.Category {
			continue
		}
		if count >= limit {
			break
		}
		products = append(products, product)
		count++
	}

	if len(products) == 0 {
		return "暂无商品数据", nil
	}

	return fmt.Sprintf("共 %d 件商品：\n%s", len(products), s.formatProducts(products)), nil
}

func (s *ProductStore) formatProduct(product *Product) string {
	data, _ := json.MarshalIndent(product, "", "  ")
	return string(data)
}

func (s *ProductStore) formatProducts(products []*Product) string {
	var result string
	for i, product := range products {
		if i > 0 {
			result += "\n---\n"
		}
		result += s.formatProduct(product)
	}
	return result
}
