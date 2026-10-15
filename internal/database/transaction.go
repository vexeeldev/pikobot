package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TransactionType string

const (
	TypeIncome     TransactionType = "income"
	TypeExpense    TransactionType = "expense"
	TypeSetBalance TransactionType = "set_balance"
	TypeAddBalance TransactionType = "add_balance"
)

type Transaction struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TelegramID   int64              `bson:"telegram_id" json:"telegram_id"`
	Type         TransactionType    `bson:"type" json:"type"`
	Amount       int64              `bson:"amount" json:"amount"`
	BalanceAfter int64              `bson:"balance_after" json:"balance_after"`
	Description  string             `bson:"description" json:"description"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

type TransactionRepository struct {
	collection *mongo.Collection
}

func NewTransactionRepository(db *mongo.Database) *TransactionRepository {
	return &TransactionRepository{
		collection: db.Collection("transactions"),
	}
}

func (r *TransactionRepository) Record(ctx context.Context, t *Transaction) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, t)
	return err
}

func (r *TransactionRepository) GetRecentTransactions(ctx context.Context, telegramID int64, limit int64) ([]Transaction, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, bson.M{"telegram_id": telegramID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []Transaction
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
