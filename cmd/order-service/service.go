package main

// Order service demo implementation.

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents an item in an order.
type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// Order represents an order in the system.
type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	Username    string      `json:"username"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderStore is an in-memory order store.
type OrderStore struct {
	mu     sync.RWMutex
	orders map[string]*Order
}

// NewOrderStore creates a new order store with sample data.
func NewOrderStore() *OrderStore {
	store := &OrderStore{
		orders: make(map[string]*Order),
	}

	// Add sample orders
	sampleOrders := []*Order{
		{
			ID:       "order-001",
			UserID:   "user-001",
			Username: "zhangsan",
			Items: []OrderItem{
				{ProductID: "prod-001", ProductName: "iPhone 15", Quantity: 1, Price: 7999.00},
				{ProductID: "prod-003", ProductName: "AirPods Pro", Quantity: 1, Price: 1999.00},
			},
			TotalAmount: 9998.00,
			Status:      OrderStatusDelivered,
			CreatedAt:   time.Now().Add(-15 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-10 * 24 * time.Hour),
		},
		{
			ID:       "order-002",
			UserID:   "user-001",
			Username: "zhangsan",
			Items: []OrderItem{
				{ProductID: "prod-002", ProductName: "MacBook Pro", Quantity: 1, Price: 14999.00},
			},
			TotalAmount: 14999.00,
			Status:      OrderStatusShipped,
			CreatedAt:   time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
		},
		{
			ID:       "order-003",
			UserID:   "user-002",
			Username: "lisi",
			Items: []OrderItem{
				{ProductID: "prod-004", ProductName: "iPad Air", Quantity: 2, Price: 4999.00},
			},
			TotalAmount: 9998.00,
			Status:      OrderStatusPaid,
			CreatedAt:   time.Now().Add(-1 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	for _, order := range sampleOrders {
		store.orders[order.ID] = order
	}

	return store
}

// QueryOrderArgs are the arguments for querying orders.
type QueryOrderArgs struct {
	OrderID  string `json:"order_id,omitempty" description:"Order ID to search for"`
	UserID   string `json:"user_id,omitempty" description:"User ID to filter orders"`
	Username string `json:"username,omitempty" description:"Username to filter orders"`
	Status   string `json:"status,omitempty" description:"Order status filter (pending/paid/shipped/delivered/cancelled)"`
}

// QueryOrder queries orders by various criteria.
func (s *OrderStore) QueryOrder(ctx context.Context, args *QueryOrderArgs) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Query by order ID
	if args.OrderID != "" {
		if order, ok := s.orders[args.OrderID]; ok {
			return s.formatOrder(order), nil
		}
		return "未找到订单号为 " + args.OrderID + " 的订单", nil
	}

	// Filter by criteria
	var results []*Order
	for _, order := range s.orders {
		match := true

		if args.UserID != "" && order.UserID != args.UserID {
			match = false
		}
		if args.Username != "" && order.Username != args.Username {
			match = false
		}
		if args.Status != "" && string(order.Status) != args.Status {
			match = false
		}

		if match {
			results = append(results, order)
		}
	}

	if len(results) == 0 {
		return "未找到符合条件的订单", nil
	}

	return s.formatOrders(results), nil
}

// CreateOrderArgs are the arguments for creating an order.
type CreateOrderArgs struct {
	UserID   string      `json:"user_id" description:"User ID for the order"`
	Username string      `json:"username" description:"Username for the order"`
	Items    []OrderItem `json:"items" description:"List of items to order"`
}

// CreateOrder creates a new order.
func (s *OrderStore) CreateOrder(ctx context.Context, args *CreateOrderArgs) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(args.Items) == 0 {
		return "订单必须包含至少一个商品", nil
	}

	var totalAmount float64
	for _, item := range args.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	order := &Order{
		ID:          fmt.Sprintf("order-%03d", len(s.orders)+1),
		UserID:      args.UserID,
		Username:    args.Username,
		Items:       args.Items,
		TotalAmount: totalAmount,
		Status:      OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.orders[order.ID] = order

	return fmt.Sprintf("订单创建成功！\n%s", s.formatOrder(order)), nil
}

// UpdateOrderStatusArgs are the arguments for updating order status.
type UpdateOrderStatusArgs struct {
	OrderID string `json:"order_id" description:"Order ID to update"`
	Status  string `json:"status" description:"New status (pending/paid/shipped/delivered/cancelled)"`
}

// UpdateOrderStatus updates an order's status.
func (s *OrderStore) UpdateOrderStatus(ctx context.Context, args *UpdateOrderStatusArgs) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[args.OrderID]
	if !ok {
		return "未找到订单号为 " + args.OrderID + " 的订单", nil
	}

	order.Status = OrderStatus(args.Status)
	order.UpdatedAt = time.Now()

	return fmt.Sprintf("订单状态更新成功！\n%s", s.formatOrder(order)), nil
}

// ListOrdersArgs are the arguments for listing orders.
type ListOrdersArgs struct {
	Limit int `json:"limit,omitempty" description:"Maximum number of orders to return"`
}

// ListOrders lists all orders.
func (s *OrderStore) ListOrders(ctx context.Context, args *ListOrdersArgs) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := args.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var orders []*Order
	count := 0
	for _, order := range s.orders {
		if count >= limit {
			break
		}
		orders = append(orders, order)
		count++
	}

	if len(orders) == 0 {
		return "暂无订单数据", nil
	}

	return fmt.Sprintf("共 %d 个订单：\n%s", len(s.orders), s.formatOrders(orders)), nil
}

func (s *OrderStore) formatOrder(order *Order) string {
	data, _ := json.MarshalIndent(order, "", "  ")
	return string(data)
}

func (s *OrderStore) formatOrders(orders []*Order) string {
	var result string
	for i, order := range orders {
		if i > 0 {
			result += "\n---\n"
		}
		result += s.formatOrder(order)
	}
	return result
}
