// Package mongo implements ports.HistoryStore: one document per symbol
// holding its daily-close series, fed by the initial backfill and the
// daily history_daily worker, and read by the scrubber / trend layer.
package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"investment-map/internal/ports"
)

const historyCollection = "history"

type Adapter struct {
	client   *mongo.Client
	database *mongo.Database
}

func New(ctx context.Context, uri, database string) (*Adapter, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo: connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo: ping: %w", err)
	}
	return &Adapter{client: client, database: client.Database(database)}, nil
}

// EnsureIndexes creates the unique index on symbol needed for upserts.
// Safe to call on every startup.
func (a *Adapter) EnsureIndexes(ctx context.Context) error {
	_, err := a.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "symbol", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("mongo: ensure indexes: %w", err)
	}
	return nil
}

func (a *Adapter) collection() *mongo.Collection {
	return a.database.Collection(historyCollection)
}

func (a *Adapter) Close(ctx context.Context) error {
	return a.client.Disconnect(ctx)
}

var _ ports.HistoryStore = (*Adapter)(nil)
