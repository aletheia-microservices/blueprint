package simpleshop

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type Product struct {
	ID          string `bson:"ID"`
	Description string `bson:"Description"`
}

type ProductService interface {
	RegisterProduct(ctx context.Context, id string, description string, amount int) (Product, error)
	DeleteProduct(ctx context.Context, id string) error
	GetProduct(ctx context.Context, id string) (Product, error)
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

func (s *ProductServiceImpl) GetProduct(ctx context.Context, id string) (Product, error) {
	collection, err := s.productDB.GetCollection(ctx, "product_db", "product")
	if err != nil {
		return Product{}, err
	}

	filter := bson.D{{Key: "ID", Value: id}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return Product{}, err
	}

	var product Product
	ok, err := cursor.One(ctx, &product)
	if err != nil {
		return Product{}, err
	}
	if !ok {
		return Product{}, fmt.Errorf("could not find product for id (%s)", id)
	}

	return product, nil
}
