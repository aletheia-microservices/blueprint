package digota

import (
	"context"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	orderTTL                   = time.Minute * 2
	defaultTaxDescription      = "Tax"
	defaultDiscountDescription = "Discount"
	defaultShippingDescription = "Shipping"
)

type OrderService interface {
	New(ctx context.Context, currency int32, items []*OrderItem, metadata map[string]string, email string, shipping *Shipping) (*Order, error)
	Get(ctx context.Context, id string) (*Order, error)
	List(ctx context.Context, page int64, limit int64, sort int32) (*OrderList, error)
	Pay(ctx context.Context, id string, card *Card, paymentProviderID int32) (*Order, error)
	Return(ctx context.Context, id string) (*Order, error)
}

type OrderServiceImpl struct {
	skuService     SkuService
	paymentService PaymentService
	db             backend.NoSQLDatabase
}

func NewOrderServiceImpl(ctx context.Context, skuService SkuService, paymentService PaymentService, db backend.NoSQLDatabase) (OrderService, error) {
	return &OrderServiceImpl{skuService: skuService, paymentService: paymentService, db: db}, nil
}

func (s *OrderServiceImpl) New(ctx context.Context, currency int32, items []*OrderItem, metadata map[string]string, email string, shipping *Shipping) (*Order, error) {
	order := &Order{
		Id:       uuid.NewString(),
		Currency: currency,
		Items:    items,
		Metadata: metadata,
		Email:    email,
		Shipping: shipping,
		Created:  time.Now().Unix(),
	}

	var orderItems []*OrderItem

	orderItems, err := s.getUpdatedOrderItems(ctx, items)
	if err != nil {
		return nil, fmt.Errorf("[OrderService] error getting updated order items: %v", err)
	}
	order.Items = orderItems

	amount, err := calculateTotal(order.Currency, order.Items)
	if err != nil {
		return nil, fmt.Errorf("[OrderService] error calculating total amount: %v", err)
	}
	order.Amount = amount

	collection, err := s.db.GetCollection(ctx, "orders_db", "orders")
	if err != nil {
		return nil, fmt.Errorf("[OrderService] error getting orders collection: %v", err)
	}
	err = collection.InsertOne(ctx, order)
	return order, err
}

func (s *OrderServiceImpl) Get(ctx context.Context, id string) (*Order, error) {
	collection, err := s.db.GetCollection(ctx, "orders_db", "orders")
	if err != nil {
		return nil, err
	}

	query := bson.D{{Key: "Id", Value: id}}
	result, err := collection.FindOne(ctx, query)
	if err != nil {
		return nil, err
	}

	var order Order
	found, err := result.One(ctx, &order)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("order not found for id (%s)", id)
	}

	return &order, nil
}

func (s *OrderServiceImpl) List(ctx context.Context, page int64, limit int64, sort int32) (*OrderList, error) {
	collection, err := s.db.GetCollection(ctx, "orders_db", "orders")
	if err != nil {
		return nil, err
	}

	result, err := collection.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var orders []Order
	err = result.All(ctx, &orders)
	if err != nil {
		return nil, err
	}

	orderList := &OrderList{
		Orders: orders,
		Total:  int32(len(orders)),
	}

	return orderList, nil
}

func (s *OrderServiceImpl) Pay(ctx context.Context, id string, card *Card, paymentProviderID int32) (*Order, error) {
	order := &Order{
		Id: id,
	}
	// get order
	if err := s.storageGetOne(ctx, order); err != nil {
		return nil, err
	}
	// check if order is payable
	if err := order.IsPayable(); err != nil {
		return nil, err
	}
	// lock all inventory order items (inventory objects)
	lockedItems := s.getLockedOrderItems(ctx, order)
	// Free all locks at func return
	defer func() {
		for _, v1 := range lockedItems {
			v1.Unlock()
		}
	}()
	// Check for errors and oversell
	for _, item := range lockedItems {
		if item.Err != nil {
			// additional logic to confirm inconsistencies between order items and inventory items
			// referential integrity: absence of cascading deletes
			return nil, fmt.Errorf("[WARNING] dangling reference found: order (%s) references sku (%s) which could not be retrieved: %v", order.Id, item.OrderItem.Parent, item.Err)
		}
		// check for oversell
		if item.Sku.Inventory != nil && item.Sku.Inventory.Type == int32(Inventory_Finite) && item.Sku.Inventory.Quantity < item.OrderItem.Quantity {
			return nil, fmt.Errorf("Oversell %s", item.Sku.Id)
		}
	}
	// charge full amount for the order order
	c, err := s.paymentService.NewCharge(ctx, order.Currency, uint64(order.Amount), card, order.Email, fmt.Sprintf("Order %s", order.Id), paymentProviderID, nil)
	if err != nil {
		return nil, err
	}
	order.ChargeId = c.Id
	order.Status = int32(Order_Paid)
	updateErr := s.storageUpdate(ctx, order)
	if updateErr != nil {
		r, err := s.paymentService.RefundCharge(ctx, c.Id, uint64(order.Amount), 0)
		if err != nil {
			return nil, fmt.Errorf("could not update order {%s} object and could not refund the charge {%s}!", order.Id, order.ChargeId)
		}
		return nil, fmt.Errorf("could not update order {%s} object, order has been refunded {%s}!", order.Id, r.Id)
	}
	// update all inventories
	for _, item := range lockedItems {
		if item.Sku.Inventory != nil && item.Sku.Inventory.Type == int32(Inventory_Finite) {
			// update inventory Quantity
			item.Sku.Inventory.Quantity -= item.OrderItem.Quantity
			item.Update()
		}
	}
	return order, nil
}

func (s *OrderServiceImpl) Return(ctx context.Context, id string) (*Order, error) {
	order := &Order{
		Id: id,
	}

	// get order
	if err := s.storageGetOne(ctx, order); err != nil {
		return nil, err
	}

	// calculate returns amount
	amount, err := calculateTotal(order.GetCurrency(), order.GetItems())
	if err != nil {
		return nil, err
	}

	// check if order can be refunded
	if err = order.IsReturnable(amount); err != nil {
		return nil, err
	}
	// lock all inventory order items (inventory objects)
	lockedItems := s.getLockedOrderItems(ctx, order)
	// Free all locks at func return
	defer func() {
		for _, item := range lockedItems {
			item.Unlock()
		}
	}()
	for _, item := range lockedItems {
		if item.Err != nil {
			// additional logic to confirm inconsistencies between order items and inventory items
			// referential integrity: absence of cascading deletes
			return nil, fmt.Errorf("[WARNING] dangling reference found: order (%s) references sku (%s) which could not be retrieved: %v", order.Id, item.OrderItem.Parent, item.Err)
		}
	}
	// refund the order
	if _, err := s.paymentService.RefundCharge(ctx, order.GetChargeId(), uint64(amount), 0); err != nil {
		return nil, err
	}
	// update order status
	switch order.Status {
	case int32(Order_Paid):
		order.Status = int32(Order_Canceled)
		for _, item := range lockedItems {
			if item.Sku.Inventory != nil && item.Sku.Inventory.Type == int32(Inventory_Finite) {
				item.Sku.Inventory.Quantity += item.OrderItem.Quantity
				item.Update()
			}
		}
	case int32(Order_Fulfilled):
		order.Status = int32(Order_Returned)
	}

	updateErr := s.storageUpdate(ctx, order)
	if updateErr != nil {
		return nil, updateErr
	}
	return order, nil
}

func (s *OrderServiceImpl) storageGetOne(ctx context.Context, order *Order) error {
	collection, err := s.db.GetCollection(ctx, "orders_db", "orders")
	if err != nil {
		return err
	}

	query := bson.D{{Key: "Id", Value: order.Id}}
	result, err := collection.FindOne(ctx, query)
	if err != nil {
		return err
	}

	found, err := result.One(ctx, order)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("order not found for id (%s)", order.Id)
	}

	return nil
}

func (s *OrderServiceImpl) storageUpdate(ctx context.Context, order *Order) error {
	collection, err := s.db.GetCollection(ctx, "orders_db", "orders")
	if err != nil {
		return err
	}

	query := bson.D{{Key: "Id", Value: order.Id}}
	updated, err := collection.ReplaceOne(ctx, query, order)
	if err != nil {
		return err
	}
	if updated == 0 {
		return fmt.Errorf("order not found for id (%s)", order.Id)
	}
	return nil
}

func calculateTotal(currency int32, orderItems []*OrderItem) (int64, error) {
	var err error
	currencyString := Currency_name[int32(currency)]
	m := money.New(0, currencyString)
	for _, v := range orderItems {
		if v.Quantity <= 0 {
			v.Quantity = 1
		}
		vCurrencyString := Currency_name[int32(v.Currency)]
		m, err = m.Add(money.New(v.Quantity*v.Amount, vCurrencyString))
		if err != nil {
			return 0, fmt.Errorf("[OrderService] error calculating total amount (currency: %s, vCurrency: %s): %v", currencyString, vCurrencyString, err)
		}
	}
	return m.Amount(), nil
}

func (s *OrderServiceImpl) getUpdatedOrderItems(ctx context.Context, reqItems []*OrderItem) (orderItems []*OrderItem, err error) {
	var skuMap = make(map[string]*OrderItem)
	var errs []error

	// merge duplicated items
	for _, v := range reqItems {
		if v.IsTypeSku() {
			if skuItem, ok := skuMap[v.Parent]; ok {
				skuItem.Quantity += v.Quantity
				continue
			} else {
				skuMap[v.Parent] = v
			}
		} else if v.IsTypeDiscount() {
			if v.Quantity <= 0 {
				v.Quantity = 1
			}
		} else if v.IsTypeShipping() {
			if v.Quantity <= 0 {
				v.Quantity = 1
			}
		} else if v.IsTypeTax() {
			if v.Quantity <= 0 {
				v.Quantity = 1
			}
		}
		orderItems = append(orderItems, v)
	}

	// update order item data
	for _, v := range orderItems {
		if v.IsTypeSku() {
			orderItem := v
			item, err := s.skuService.Get(ctx, v.Parent)
			if err != nil {
				return nil, err
			} else {
				orderItem.Amount = int64(item.Price)
				orderItem.Currency = item.Currency
				orderItem.Description = item.Name
			}
		} else if v.IsTypeDiscount() {
			// nothing to fetch yet
			if v.Description == "" {
				v.Description = defaultDiscountDescription
			}
		} else if v.IsTypeShipping() {
			// nothing to fetch yet
			if v.Description == "" {
				v.Description = defaultShippingDescription
			}
		} else if v.IsTypeTax() {
			// nothing to fetch yet
			if v.Description == "" {
				v.Description = defaultTaxDescription
			}
		}
	}

	// return the first error
	if errs != nil {
		return nil, errs[0]
	}

	return
}

type lockedOrderItem struct {
	OrderItem *OrderItem
	Sku       *Sku
	Unlock    func() error
	Update    func() error
	Err       error
}

func (s *OrderServiceImpl) getLockedOrderItems(ctx context.Context, order *Order) (items []*lockedOrderItem) {
	for _, orderItem := range order.GetItems() {
		if orderItem.IsTypeSku() {
			item, err := s.skuService.Get(ctx, orderItem.Parent)
			items = append(items, &lockedOrderItem{
				OrderItem: orderItem,
				Sku:       item,
				Unlock:    func() error { return nil },
				Update:    func() error { return nil },
				Err:       err,
			})
		}
	}
	return
}
