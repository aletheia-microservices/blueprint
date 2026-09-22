// Package Frontend implements the SockShop Frontend service, typically deployed via HTTP
package frontend

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/carts"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/catalogue"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/orders"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"
)

type Frontend interface {
	// cart
	GetCart(ctx context.Context, sessionID string) ([]carts.Item, error)
	AddItem(ctx context.Context, sessionID string, itemID string) (newSessionID string, err error)
	DeleteCart(ctx context.Context, sessionID string) error
	RemoveItem(ctx context.Context, sessionID string, itemID string) error
	UpdateItem(ctx context.Context, sessionID string, itemID string, quantity int) (newSessionID string, err error)

	// items
	ListItems(ctx context.Context, tags []string, order string, pageNum, pageSize int) ([]catalogue.Sock, error)
	GetSock(ctx context.Context, itemID string) (catalogue.Sock, error)

	// tags
	ListTags(ctx context.Context) ([]string, error)

	// orders
	NewOrder(ctx context.Context, userID, addressID, cardID, cartID string) (orders.Order, error)
	GetOrders(ctx context.Context, userID string) ([]orders.Order, error)
	GetOrder(ctx context.Context, orderID string) (orders.Order, error)

	// users
	Login(ctx context.Context, sessionID, username, password string) (newSessionID string, u user.User, err error)
	Register(ctx context.Context, sessionID, username, password, email, first, last string) (newSessionID string, err error)
	GetUser(ctx context.Context, userID string) (user.User, error)
	GetAddress(ctx context.Context, addressID string) (user.Address, error)
	PostAddress(ctx context.Context, userID string, address user.Address) (string, error)
	GetCard(ctx context.Context, cardID string) (user.Card, error)
	PostCard(ctx context.Context, userID string, card user.Card) (string, error)

	// extra endpoints
	// catalogue
	LoadCatalogueTags(ctx context.Context) (string, error)
	LoadCatalogueSocks(ctx context.Context) (string, error)
}

type FrontendImpl struct {
	user      user.UserService
	catalogue catalogue.CatalogueService
	cart      carts.CartService
	order     orders.OrderService
}

func NewFrontendImpl(ctx context.Context, user user.UserService, catalogue catalogue.CatalogueService, cart carts.CartService, order orders.OrderService) (Frontend, error) {
	f := &FrontendImpl{
		user:      user,
		catalogue: catalogue,
		cart:      cart,
		order:     order,
	}
	return f, nil
}

func (f *FrontendImpl) AddItem(ctx context.Context, sessionID string, itemID string) (string, error) {
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	sock, err := f.catalogue.Get(ctx, itemID)
	if err != nil {
		return sessionID, err
	}

	_, err = f.cart.AddItem(ctx, sessionID, carts.Item{ID: sock.SockID, Quantity: 1, UnitPrice: sock.Price})
	return sessionID, err
}

func (f *FrontendImpl) RemoveItem(ctx context.Context, sessionID string, itemID string) error {
	if sessionID == "" {
		return nil
	}

	return f.cart.RemoveItem(ctx, sessionID, itemID)
}

func (f *FrontendImpl) GetCart(ctx context.Context, sessionID string) ([]carts.Item, error) {
	if sessionID == "" {
		return nil, nil
	}

	return f.cart.GetCart(ctx, sessionID)
}

func (f *FrontendImpl) DeleteCart(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}

	return f.cart.DeleteCart(ctx, sessionID)
}

func (f *FrontendImpl) GetUser(ctx context.Context, userID string) (user.User, error) {
	if userID == "" {
		return user.User{}, errors.Errorf("no userID specified")
	}

	users, err := f.user.GetUsers(ctx, userID)
	if err != nil {
		return user.User{}, err
	} else if len(users) == 0 {
		return user.User{}, errors.Errorf("invalid userID %v", userID)
	} else {
		return users[0], nil
	}
}

func (f *FrontendImpl) GetAddress(ctx context.Context, addressID string) (user.Address, error) {
	if addressID == "" {
		return user.Address{}, errors.Errorf("no addressID specified")
	}
	addrs, err := f.user.GetAddresses(ctx, addressID)
	if err != nil {
		return user.Address{}, err
	} else if len(addrs) == 0 {
		return user.Address{}, errors.Errorf("invalid addressID %v", addressID)
	} else {
		return addrs[0], nil
	}
}

func (f *FrontendImpl) GetCard(ctx context.Context, cardID string) (user.Card, error) {
	if cardID == "" {
		return user.Card{}, errors.Errorf("no cardID specified")
	}
	cards, err := f.user.GetCards(ctx, cardID)
	if err != nil {
		return user.Card{}, err
	} else if len(cards) == 0 {
		return user.Card{}, errors.Errorf("invalid cardID %v", cardID)
	} else {
		return cards[0], nil
	}
}

func (f *FrontendImpl) GetOrder(ctx context.Context, orderID string) (orders.Order, error) {
	return f.order.GetOrder(ctx, orderID)
}

func (f *FrontendImpl) GetOrders(ctx context.Context, userID string) ([]orders.Order, error) {
	if userID == "" {
		return nil, errors.Errorf("no userID specified")
	}
	return f.order.GetOrders(ctx, userID)
}

func (f *FrontendImpl) GetSock(ctx context.Context, itemID string) (catalogue.Sock, error) {
	return f.catalogue.Get(ctx, itemID)
}

func (f *FrontendImpl) ListItems(ctx context.Context, tags []string, order string, pageNum int, pageSize int) ([]catalogue.Sock, error) {
	return f.catalogue.List(ctx, tags, order, pageNum, pageSize)
}

func (f *FrontendImpl) ListTags(ctx context.Context) ([]string, error) {
	return f.catalogue.Tags(ctx)
}

func (f *FrontendImpl) Login(ctx context.Context, sessionID string, username string, password string) (string, user.User, error) {
	u, err := f.user.Login(ctx, username, password)
	if err != nil {
		return sessionID, user.User{}, err
	}

	if sessionID != "" {
		if err := f.cart.MergeCarts(ctx, u.UserID, sessionID); err != nil {
			return u.UserID, u, err
		}
	}

	return u.UserID, u, nil
}

func (f *FrontendImpl) NewOrder(ctx context.Context, userID string, addressID string, cardID string, cartID string) (orders.Order, error) {
	return f.order.NewOrder(ctx, userID, addressID, cardID, cartID)
}

func (f *FrontendImpl) PostAddress(ctx context.Context, userID string, address user.Address) (string, error) {
	return f.user.PostAddress(ctx, userID, address)
}

func (f *FrontendImpl) PostCard(ctx context.Context, userID string, card user.Card) (string, error) {
	return f.user.PostCard(ctx, userID, card)
}

func (f *FrontendImpl) Register(ctx context.Context, sessionID string, username string, password string, email string, first string, last string) (string, error) {
	userID, err := f.user.Register(ctx, username, password, email, first, last)
	if err != nil {
		return sessionID, err
	}

	if sessionID != "" {
		return userID, f.cart.MergeCarts(ctx, userID, sessionID)
	} else {
		return userID, nil
	}
}

func (f *FrontendImpl) UpdateItem(ctx context.Context, sessionID string, itemID string, quantity int) (string, error) {
	item, err := f.catalogue.Get(ctx, itemID)
	if err != nil {
		return sessionID, err
	}

	return sessionID, f.cart.UpdateItem(ctx, sessionID, carts.Item{ID: item.SockID, Quantity: quantity, UnitPrice: item.Price})
}

func (f *FrontendImpl) LoadCatalogueTags(ctx context.Context) (string, error) {
	var alltags = []string{"brown", "geek", "formal", "blue", "skin", "red", "action", "sport", "black", "magic", "green"}

	if err := f.catalogue.AddTags(ctx, alltags); err != nil {
		return "Failed to load catalogue tags", err
	}

	return "Load catalogue tags successful", nil
}

func (f *FrontendImpl) LoadCatalogueSocks(ctx context.Context) (string, error) {
	sock := func(name, description string, price float32, qty int, url1, url2 string, tags ...string) catalogue.Sock {
		return catalogue.Sock{Name: name, Description: description,
			Price: price, Quantity: qty, ImageURL1: url1, ImageURL2: url2, Tags: tags}
	}

	var socks = []catalogue.Sock{
		sock("Weave special", "Limited issue Weave socks.", 17.15, 33, "/catalogue/images/weave1.jpg", "/catalogue/images/weave2.jpg", "geek", "black"),
		sock("Nerd leg", "For all those leg lovers out there. A perfect example of a swivel chair trained calf. Meticulously trained on a diet of sitting and Pina Coladas. Phwarr...", 7.99, 115, "/catalogue/images/bit_of_leg_1.jpeg", "/catalogue/images/bit_of_leg_2.jpeg", "blue", "skin"),
		sock("Crossed", "A mature sock, crossed, with an air of nonchalance.", 17.32, 738, "/catalogue/images/cross_1.jpeg", "/catalogue/images/cross_2.jpeg", "formal", "blue", "red", "action"),
		sock("SuperSport XL", "Ready for action. Engineers: be ready to smash that next bug! Be ready, with these super-action-sport-masterpieces. This particular engineer was chased away from the office with a stick.", 15.00, 820, "/catalogue/images/puma_1.jpeg", "/catalogue/images/puma_2.jpeg", "formal", "sport", "black"),
		sock("Holy", "Socks fit for a Messiah. You too can experience walking in water with these special edition beauties. Each hole is lovingly proggled to leave smooth edges. The only sock approved by a higher power.", 99.99, 1, "/catalogue/images/holy_1.jpeg", "/catalogue/images/holy_2.jpeg", "action", "magic"),
		sock("YouTube.sock", "We were not paid to sell this sock. It's just a bit geeky.", 10.99, 801, "/catalogue/images/youtube_1.jpeg", "/catalogue/images/youtube_2.jpeg", "geek", "formal"),
		sock("Figueroa", "enim officia aliqua excepteur esse deserunt quis aliquip nostrud anim", 14, 808, "/catalogue/images/WAT.jpg", "/catalogue/images/WAT2.jpg", "formal", "blue", "green"),
		sock("Classic", "Keep it simple.", 12, 127, "/catalogue/images/classic.jpg", "/catalogue/images/classic2.jpg", "brown", "green"),
		sock("Colourful", "proident occaecat irure et excepteur labore minim nisi amet irure", 18, 438, "/catalogue/images/colourful_socks.jpg", "/catalogue/images/colourful_socks.jpg", "brown", "blue"),
		sock("Cat socks", "consequat amet cupidatat minim laborum tempor elit ex consequat in", 15, 175, "/catalogue/images/catsocks.jpg", "/catalogue/images/catsocks2.jpg", "brown", "formal", "green"),
	}

	for _, s := range socks {
		if _, err := f.catalogue.AddSock(ctx, s); err != nil {
			return "Failed to load catalogue socks", err
		}
	}

	return "Load catalogue socks successful", nil
}
