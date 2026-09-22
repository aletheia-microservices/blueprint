package simpleshop

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type Inventory struct {
	ID     string
	Amount int
}

type InventoryService interface {
	AddInventory(ctx context.Context, id string, amount int) (Inventory, error)
	DeleteInventory(ctx context.Context, id string) error
}

type InventoryServiceImpl struct {
	inventoryDB backend.NoSQLDatabase
}

func NewBarServiceImpl(ctx context.Context, inventoryDB backend.NoSQLDatabase) (InventoryService, error) {
	d := &InventoryServiceImpl{inventoryDB: inventoryDB}
	return d, nil
}

func (s *InventoryServiceImpl) AddInventory(ctx context.Context, id string, amount int) (Inventory, error) {
	bar := Inventory{
		ID:     id,
		Amount: amount,
	}

	collection, err := s.inventoryDB.GetCollection(ctx, "inventory_db", "inventory")
	if err != nil {
		return Inventory{}, err
	}

	err = collection.InsertOne(ctx, bar)
	if err != nil {
		return Inventory{}, err
	}

	return bar, nil
}

func (s *InventoryServiceImpl) DeleteInventory(ctx context.Context, id string) error {
	collection, err := s.inventoryDB.GetCollection(ctx, "inventory_db", "inventory")
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "ID", Value: id}}
	return collection.DeleteOne(ctx, filter)
}
