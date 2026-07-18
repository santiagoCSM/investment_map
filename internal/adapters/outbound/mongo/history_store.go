package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"investment-map/internal/domain"
)

type historyPointDoc struct {
	Date  time.Time `bson:"date"`
	Close float64   `bson:"close"`
}

type historyDoc struct {
	Symbol string            `bson:"symbol"`
	Points []historyPointDoc `bson:"points"`
}

// Append adds one day's close to symbol's series. It first removes any
// existing point for the same date so re-running the daily job (retries,
// backfill overlap) is idempotent instead of producing duplicate entries.
func (a *Adapter) Append(ctx context.Context, symbol string, point domain.HistoryPoint) error {
	coll := a.collection()
	day := truncateToDay(point.Date)

	_, err := coll.UpdateOne(ctx,
		bson.M{"symbol": symbol},
		bson.M{"$pull": bson.M{"points": bson.M{"date": day}}},
	)
	if err != nil {
		return fmt.Errorf("mongo: append %s: pull existing point: %w", symbol, err)
	}

	_, err = coll.UpdateOne(ctx,
		bson.M{"symbol": symbol},
		bson.M{
			"$push":        bson.M{"points": historyPointDoc{Date: day, Close: point.Close}},
			"$setOnInsert": bson.M{"symbol": symbol},
		},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("mongo: append %s: push point: %w", symbol, err)
	}
	return nil
}

// Query returns symbol's daily-close series within rng, oldest first.
func (a *Adapter) Query(ctx context.Context, symbol string, rng domain.HistoryRange) ([]domain.HistoryPoint, error) {
	var doc historyDoc
	err := a.collection().FindOne(ctx, bson.M{"symbol": symbol}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mongo: query %s: %w", symbol, err)
	}

	cutoff, hasCutoff := rangeCutoff(rng)
	points := make([]domain.HistoryPoint, 0, len(doc.Points))
	for _, p := range doc.Points {
		if hasCutoff && p.Date.Before(cutoff) {
			continue
		}
		points = append(points, domain.HistoryPoint{Date: p.Date, Close: p.Close})
	}
	return points, nil
}

func rangeCutoff(rng domain.HistoryRange) (time.Time, bool) {
	now := time.Now().UTC()
	switch rng {
	case domain.HistoryRange1Month:
		return now.AddDate(0, -1, 0), true
	case domain.HistoryRange3Month:
		return now.AddDate(0, -3, 0), true
	case domain.HistoryRange1Year:
		return now.AddDate(-1, 0, 0), true
	case domain.HistoryRangeMax:
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}

func truncateToDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
