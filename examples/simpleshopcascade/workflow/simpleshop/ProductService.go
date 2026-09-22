package simpleshop

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type Product struct {
	ID          string
	Description string
}

type ProductService interface {
	RegisterProduct(ctx context.Context, id string, description string, amount int) (Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

type ProductServiceImpl struct {
	productDB        backend.NoSQLDatabase
	inventoryService InventoryService
}

func NewProductServiceImpl(ctx context.Context, database backend.NoSQLDatabase, inventoryService InventoryService) (ProductService, error) {
	d := &ProductServiceImpl{productDB: database, inventoryService: inventoryService}
	return d, nil
}

func (s *ProductServiceImpl) RegisterProduct(ctx context.Context, id string, description string, amount int) (Product, error) {
	product := Product{
		ID:          id,
		Description: description,
	}

	collection, err := s.productDB.GetCollection(ctx, "product_db", "product")
	if err != nil {
		return Product{}, err
	}

	err = collection.InsertOne(ctx, product)
	if err != nil {
		return Product{}, err
	}

	_, err = s.inventoryService.AddInventory(ctx, id, amount)

	return product, err
}

func (s *ProductServiceImpl) DeleteProduct(ctx context.Context, id string) error {
	collection, err := s.productDB.GetCollection(ctx, "product_db", "product")
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "ID", Value: id}}
	if err := collection.DeleteOne(ctx, filter); err != nil {
		return err
	}

	return s.inventoryService.DeleteInventory(ctx, id)
}
