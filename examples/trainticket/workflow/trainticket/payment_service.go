// package payment implements ts-payment-service from the original train ticket application
package trainticket

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// PaymentService manages payments in the application
type PaymentService interface {
	// Pay `payment`
	Pay(ctx context.Context, payment Payment) error
	// Adds Money to an existing user's account
	AddMoney(ctx context.Context, payment Payment) error
	// Get all payments
	Query(ctx context.Context) ([]Payment, error)
	// Create an initial payment
	InitPayment(ctx context.Context, payment Payment) error
	// Remove all payments; Only used in testing
	Cleanup(ctx context.Context) error
}

// Implementation of PaymentService
type PaymentServiceImpl struct {
	paymentDB backend.NoSQLDatabase
	moneyDB   backend.NoSQLDatabase
}

// Creates a new PaymentService object
func NewPaymentServiceImpl(ctx context.Context, payDB backend.NoSQLDatabase, moneyDB backend.NoSQLDatabase) (PaymentService, error) {
	return &PaymentServiceImpl{paymentDB: payDB, moneyDB: moneyDB}, nil
}

func (p *PaymentServiceImpl) InitPayment(ctx context.Context, payment Payment) error {
	coll, err := p.paymentDB.GetCollection(ctx, "payment_db", "payment")
	if err != nil {
		return err
	}
	res, err := coll.FindOne(ctx, bson.D{{Key: "ID", Value: payment.ID}})
	if err != nil {
		return err
	}
	var stored Payment
	exists, err := res.One(ctx, &stored)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return coll.InsertOne(ctx, payment)
}

func (p *PaymentServiceImpl) Query(ctx context.Context) ([]Payment, error) {
	var payments []Payment
	coll, err := p.paymentDB.GetCollection(ctx, "payment_db", "payment")
	if err != nil {
		return payments, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return payments, err
	}
	return payments, res.All(ctx, &payments)
}

func (p *PaymentServiceImpl) Pay(ctx context.Context, payment Payment) error {
	coll, err := p.paymentDB.GetCollection(ctx, "payment_db", "payment")
	if err != nil {
		return err
	}
	ok, err := coll.Upsert(ctx, bson.D{{Key: "OrderID", Value: payment.OrderID}}, payment)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Payment for order with orderid " + payment.OrderID + " was not found")
	}
	return nil
}

func (p *PaymentServiceImpl) AddMoney(ctx context.Context, payment Payment) error {
	m := Money{}
	m.UserID = payment.UserID
	m.Money = payment.Price
	m.ID = uuid.New().String()

	coll, err := p.moneyDB.GetCollection(ctx, "money_db", "money")
	if err != nil {
		return err
	}
	return coll.InsertOne(ctx, m)
}

func (p *PaymentServiceImpl) Cleanup(ctx context.Context) error {
	pay_coll, err := p.paymentDB.GetCollection(ctx, "payment_db", "payment")
	if err != nil {
		return err
	}
	err = pay_coll.DeleteMany(ctx, bson.D{})
	if err != nil {
		return err
	}
	money_coll, err := p.moneyDB.GetCollection(ctx, "money_db", "money")
	if err != nil {
		return err
	}
	return money_coll.DeleteMany(ctx, bson.D{})
}
