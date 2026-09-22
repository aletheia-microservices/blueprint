package user

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// The format of a User stored in the database
type dbUser struct {
	User       `bson:",inline"`
	ID         primitive.ObjectID   `bson:"_id"`
	AddressIDs []primitive.ObjectID `bson:"Addresses"`
	CardIDs    []primitive.ObjectID `bson:"Cards"`
}

// Sets the user's ID to be the hex string of the database ObjectID.
// Also constructs (empty) Address and Card objects containing the IDs
// of the user's addresses and cards.
func (u *dbUser) AddUserIDs() {
	u.User.UserID = u.ID.Hex()
	u.User.Addresses = nil
	u.User.Cards = nil
	for _, id := range u.AddressIDs {
		u.Addresses = append(u.Addresses, Address{ID: id.Hex()})
	}
	for _, id := range u.CardIDs {
		u.Cards = append(u.Cards, Card{ID: id.Hex()})
	}
}

// Generates database IDs for the user then adds to the database
func (s *UserServiceImpl) userdb_CreateUser(ctx context.Context, user *User) error {
	collection, err := s.db.GetCollection(ctx, "user_db", "user")
	if err != nil {
		return err
	}

	u := dbUser{
		User:       *user,
		ID:         primitive.NewObjectID(),
		AddressIDs: []primitive.ObjectID{},
		CardIDs:    []primitive.ObjectID{},
	}
	if u.CardIDs, err = s.carddb_CreateCards(ctx, user.Cards); err != nil {
		return err
	}
	if u.AddressIDs, err = s.addressdb_CreateAddresses(ctx, user.Addresses); err != nil {
		return err
	}
	_, err = collection.UpsertID(ctx, u.ID, u)
	if err != nil {
		// Gonna clean up if we can, ignore error
		// because the user save error takes precedence.
		s.addressdb_RemoveAddresses(ctx, u.AddressIDs)
		s.carddb_RemoveCards(ctx, u.CardIDs)
		return err
	}
	u.UserID = u.ID.Hex()
	*user = u.User
	return nil
}

// Get user by their name
func (s *UserServiceImpl) userdb_GetUserByName(ctx context.Context, username string) (User, error) {
	collection, err := s.db.GetCollection(ctx, "user_db", "user")
	if err != nil {
		return User{}, err
	}

	// Execute query
	cursor, err := collection.FindOne(ctx, bson.D{{Key: "Username", Value: username}})
	if err != nil {
		return newUser(), err
	}

	// Extract query result
	u := dbUser{}
	if _, err := cursor.One(ctx, &u); err != nil {
		return newUser(), err
	}

	// Set the hex string IDs for the user, cards, and addresses before returning
	u.AddUserIDs()
	return u.User, nil
}

// Get user by their object id
func (s *UserServiceImpl) userdb_GetUser(ctx context.Context, userid string) (User, error) {
	collection, err := s.db.GetCollection(ctx, "user_db", "user")
	if err != nil {
		return User{}, err
	}

	// Convert user ID to bson object ID
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return newUser(), errors.New("invalid Id Hex")
	}

	// Execute query
	cursor, err := collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return newUser(), err
	}

	// Extract query result
	u := dbUser{}
	if _, err := cursor.One(ctx, &u); err != nil {
		return newUser(), err
	}

	// Set the hex string IDs for the user, cards, and addresses before returning
	u.AddUserIDs()
	return u.User, nil
}

// Get all users
func (s *UserServiceImpl) userdb_GetUsers(ctx context.Context) ([]User, error) {
	collection, err := s.db.GetCollection(ctx, "user_db", "user")
	if err != nil {
		return nil, err
	}

	// Execute query
	cursor, err := collection.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	// Extract query results
	dbUsers := []dbUser{}
	if err := cursor.All(ctx, &dbUsers); err != nil {
		return nil, err
	}

	// Convert from database users to user objects
	users := []User{}
	for _, dbUser := range dbUsers {
		dbUser.AddUserIDs()
		users = append(users, dbUser.User)
	}
	return users, nil
}

// Given a user, load all cards and addresses connected to that user
func (s *UserServiceImpl) userdb_GetUserAttributes(ctx context.Context, u *User) error {
	// Query the address store
	addresses, err := s.addressdb_GetAddresses(ctx, u.addressIDs())
	if err != nil {
		return err
	}

	// Query the card store
	cards, err := s.carddb_GetCards(ctx, u.cardIDs())
	if err != nil {
		return err
	}

	// Set the complete address and card data on the user
	u.Addresses = addresses
	u.Cards = cards
	return nil
}

// Adds a card to the cards DB and saves it for a user if there is a user
func (s *UserServiceImpl) userdb_CreateCard(ctx context.Context, userid string, card *Card) error {
	collection, err := s.db.GetCollection(ctx, "user_db", "user")
	if err != nil {
		return err
	}

	if userid == "" {
		// An anonymous user; simply insert the card to the DB
		_, err := s.carddb_CreateCard(ctx, card)
		return err
	}

	// A userid is provided; first check it's valid
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return errors.New("invalid ID Hex")
	}

	// Insert the card to the DB
	cardID, err := s.carddb_CreateCard(ctx, card)
	if err != nil {
		return err
	}

	// Update the user
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "Cards", Value: cardID}}}}
	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

// Adds an address to the address DB and saves it for a user if there is a user
func (s *UserServiceImpl) userdb_CreateAddress(ctx context.Context, userid string, address *Address) error {
	collection, err := s.db.GetCollection(ctx, "user_db", "user")
	if err != nil {
		return err
	}

	if userid == "" {
		// An anonymous user; simply insert the address to the DB
		_, err := s.addressdb_CreateAddress(ctx, address)
		return err
	}

	// A userid is provided; first check it's valid
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return errors.New("invalid ID Hex")
	}

	// Insert the address to the DB
	addressID, err := s.addressdb_CreateAddress(ctx, address)
	if err != nil {
		return err
	}

	// Update the user
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "Addresses", Value: addressID}}}}
	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *UserServiceImpl) userdb_Delete(ctx context.Context, entity string, id string) error {
	switch entity {
	case "customers":
		return s.userdb_DeleteUser(ctx, id)
	case "addresses":
		return s.userdb_DeleteAddress(ctx, id)
	case "cards":
		return s.userdb_DeleteCard(ctx, id)
	default:
		return errors.New("invalid entity " + entity)
	}
}

func (s *UserServiceImpl) userdb_DeleteUser(ctx context.Context, userid string) error {
	collection, err := s.db.GetCollection(ctx, "user_db", "user")
	if err != nil {
		return err
	}

	// Check valid user ID
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return errors.New("invalid Id Hex")
	}

	// Get user details
	u, err := s.userdb_GetUser(ctx, userid)
	if err != nil {
		return err
	}

	// Delete addresses
	addressIds, err := hexToObjectIds(u.addressIDs())
	if err != nil {
		return err
	}
	if err := s.addressdb_RemoveAddresses(ctx, addressIds); err != nil {
		return err
	}

	// Delete cards
	cardIds, err := hexToObjectIds(u.cardIDs())
	if err != nil {
		return err
	}
	if err := s.carddb_RemoveCards(ctx, cardIds); err != nil {
		return err
	}

	// Delete user
	return collection.DeleteMany(ctx, bson.D{{Key: "_id", Value: id}})
}

func (s *UserServiceImpl) userdb_DeleteAddress(ctx context.Context, addressid string) error {
	// Remove from customers db from any customers that have this address
	if err := s.userdb_DeleteAttr(ctx, "Addresses", addressid); err != nil {
		return err
	}

	// Remove from addresses db
	return s.addressdb_RemoveAddress(ctx, addressid)
}

func (s *UserServiceImpl) userdb_DeleteCard(ctx context.Context, cardid string) error {
	// Remove from customers db from any customers that have this card
	if err := s.userdb_DeleteAttr(ctx, "Cards", cardid); err != nil {
		return err
	}

	// Remove from addresses db
	return s.carddb_RemoveCard(ctx, cardid)
}

func (s *UserServiceImpl) userdb_DeleteAttr(ctx context.Context, attr, idhex string) error {
	collection, err := s.db.GetCollection(ctx, "user_db", "user")
	if err != nil {
		return err
	}

	// Check valid ID
	id, err := primitive.ObjectIDFromHex(idhex)
	if err != nil {
		return errors.New("invalid Id Hex")
	}

	// Remove customer attr
	filter := bson.D{{Key: attr, Value: id}}
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: attr, Value: id}}}}
	_, err = collection.UpdateMany(ctx, filter, update)
	return err
}

// Converts bson object ids from hex strings to object representations
func hexToObjectIds(hexes []string) ([]primitive.ObjectID, error) {
	ids := make([]primitive.ObjectID, 0, len(hexes))
	for _, hex := range hexes {
		objectId, err := primitive.ObjectIDFromHex(hex)
		if err != nil {
			return nil, errors.New("invalid Id Hex")
		}
		ids = append(ids, objectId)
	}
	return ids, nil
}
