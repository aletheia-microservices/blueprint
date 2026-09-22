package simpleshop

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type Inventory struct {
	ID     string `bson:"ID"`
	Amount int    `bson:"Amount"`
}

type InventoryService interface {
	AddInventory(ctx context.Context, id string, amount int) (Inventory, error)
	DeleteInventory(ctx context.Context, id string) error
	GetInventory(ctx context.Context, id string) (Inventory, error)
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

func (s *InventoryServiceImpl) GetInventory(ctx context.Context, id string) (Inventory, error) {
	collection, err := s.inventoryDB.GetCollection(ctx, "inventory_db", "inventory")
	if err != nil {
		return Inventory{}, err
	}

	filter := bson.D{{Key: "ID", Value: id}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return Inventory{}, err
	}

	var inventory Inventory
	ok, err := cursor.One(ctx, &inventory)
	if err != nil {
		return Inventory{}, err
	}
	if !ok {
		return Inventory{}, fmt.Errorf("could not find inventory for id (%s)", id)
	}

	return inventory, nil
}
