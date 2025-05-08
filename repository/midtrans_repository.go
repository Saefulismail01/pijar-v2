package repository

import (
	"context"
	"database/sql"
	"errors"
	"pijar/model"
	"time"
)

// ProductRepository adalah interface untuk repository produk
type ProductRepository interface {
	GetProductByID(id int) (model.Product, error)
}

// productRepository adalah implementasi dari ProductRepository
type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetProductByID(id int) (model.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var product model.Product
	query := `SELECT id, name, description, price, created_at, updated_at FROM products WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Product{}, errors.New("product not found")
		}
		return model.Product{}, err
	}
	
	return product, nil
}

// TransactionRepository adalah interface untuk repository transaksi
type TransactionRepository interface {
	CreateTransaction(transaction model.Transaction) (model.Transaction, error)
	UpdateTransactionStatus(id int, status string) error
	GetTransactionByID(id int) (model.Transaction, error)
	UpdateTransactionStatusByOrderID(orderID string, status string, midtransID string) error
}

// transactionRepository adalah implementasi dari TransactionRepository
type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) CreateTransaction(transaction model.Transaction) (model.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO transactions 
		(user_id, product_id, amount, status, order_id, payment_url, midtrans_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id
	`
	
	// Set current time for created_at and updated_at
	now := time.Now()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		transaction.UserID,
		transaction.ProductID,
		transaction.Amount,
		transaction.Status,
		transaction.OrderID,
		transaction.PaymentURL,
		transaction.MidtransID,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	).Scan(&transaction.ID)

	if err != nil {
		return model.Transaction{}, err
	}
	
	return transaction, nil
}

func (r *transactionRepository) UpdateTransactionStatus(id int, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE transactions SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("transaction not found")
	}
	
	return nil
}

func (r *transactionRepository) GetTransactionByID(id int) (model.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var transaction model.Transaction
	query := `
		SELECT 
			id, user_id, product_id, amount, status, order_id, payment_url, midtrans_id, created_at, updated_at 
		FROM transactions 
		WHERE id = $1
	`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.ProductID,
		&transaction.Amount,
		&transaction.Status,
		&transaction.OrderID,
		&transaction.PaymentURL,
		&transaction.MidtransID,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Transaction{}, errors.New("transaction not found")
		}
		return model.Transaction{}, err
	}
	
	return transaction, nil
}

func (r *transactionRepository) UpdateTransactionStatusByOrderID(orderID string, status string, midtransID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE transactions SET status = $1, midtrans_id = $2, updated_at = $3 WHERE order_id = $4`
	result, err := r.db.ExecContext(ctx, query, status, midtransID, time.Now(), orderID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("transaction not found")
	}
	
	return nil
}