package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}

type AuctionTimer struct {
	EndTime time.Time
	EndChan *chan bool
}
type AuctionRepository struct {
	Collection    *mongo.Collection
	AuctionTimers map[string]AuctionTimer
	TimerMutex    sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection:    database.Collection("auctions"),
		AuctionTimers: make(map[string]AuctionTimer),
		TimerMutex:    sync.Mutex{},
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	// Create timer for ending auction based on a forced end by a channel or a timer
	// Timer can be passed by AUCTION_DURATION env var
	ar.TimerMutex.Lock()
	endChan := make(chan bool)
	autionDuration := getAuctionInterval()
	// Saves this info if needed to manually delete or to be accessed during runtime
	ar.AuctionTimers[auctionEntity.Id] = AuctionTimer{
		EndTime: time.Now().Add(autionDuration),
		EndChan: &endChan,
	}
	go func() {
		select {
		case <-endChan:
			ar.EndAuction(auctionEntity.Id)
			break
		case <-time.After(autionDuration):
			ar.EndAuction(auctionEntity.Id)
			break
		}
	}()
	ar.TimerMutex.Unlock()

	return nil
}

func (ar *AuctionRepository) EndAuction(auctionID string) {
	update := bson.M{
		"$set": bson.M{
			"status": auction_entity.Completed,
		},
	}
	if _, err := ar.Collection.UpdateByID(context.Background(), auctionID, update); err != nil {
		logger.Error(fmt.Sprintf("Error ending auction %s", auctionID), err)
	}
	ar.TimerMutex.Lock()
	delete(ar.AuctionTimers, auctionID)
	ar.TimerMutex.Unlock()
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}
