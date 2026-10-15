package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	TelegramID int64     `bson:"telegram_id" json:"telegram_id"`
	Username   string    `bson:"username,omitempty" json:"username,omitempty"`
	FirstName  string    `bson:"first_name" json:"first_name"`
	LastName   string    `bson:"last_name,omitempty" json:"last_name,omitempty"`
	Balance    int64     `bson:"balance" json:"balance"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
	TotalMsgs  int64     `bson:"total_messages" json:"total_messages"`
}

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) UpsertUser(ctx context.Context, u *User) error {
	filter := bson.M{"telegram_id": u.TelegramID}
	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"username":   u.Username,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
			"balance":    int64(0),
		},
		"$inc": bson.M{
			"total_messages": 1,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *UserRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"telegram_id": telegramID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) SetUserBalance(ctx context.Context, telegramID int64, newBalance int64) error {
	filter := bson.M{"telegram_id": telegramID}
	update := bson.M{
		"$set": bson.M{
			"balance":    newBalance,
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *UserRepository) AddUserBalance(ctx context.Context, telegramID int64, delta int64) (int64, error) {
	filter := bson.M{"telegram_id": telegramID}
	update := bson.M{
		"$inc": bson.M{
			"balance": delta,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var updated User
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		return 0, err
	}
	return updated.Balance, nil
}

func (r *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
