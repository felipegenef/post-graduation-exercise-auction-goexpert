package auction

import (
	"context"
	"fullcycle-auction_go/configuration/database/mongodb"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func LoadEnv() {
	if err := godotenv.Load("../../../../cmd/auction/.env"); err != nil {
		log.Fatal("Error trying to load env variables %w", err)
	}
}

func CleanupTest(ctx context.Context, repo *AuctionRepository, auctionId string) {
	repo.Collection.DeleteOne(ctx, bson.M{"_id": auctionId})
}

func TestAuctionAutoClose(t *testing.T) {
	LoadEnv()

	ctx := context.Background()

	os.Setenv("AUCTION_INTERVAL", "2s")
	os.Setenv("MONGODB_URL", "mongodb://admin:admin@localhost:27017/auctions?authSource=admin")

	databaseConnection, err := mongodb.NewMongoDBConnection(ctx)

	if err != nil {
		log.Fatal(err)
	}
	repo := NewAuctionRepository(databaseConnection)

	auction, _ := auction_entity.CreateAuction(
		"Product Test", "Tests", "A test product created to check if auction is deleted", auction_entity.New,
	)

	repo.CreateAuction(ctx, auction)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)

		var result AuctionEntityMongo
		err := repo.Collection.FindOne(ctx, bson.M{"_id": auction.Id}).Decode(&result)
		if err != nil {
			t.Errorf("Failed to find auction after 1s: %v", err)
			CleanupTest(ctx, repo, auction.Id)
			return
		}

		if result.Status != auction_entity.Active {
			t.Errorf("Expected auction to be Active after 1s, got: %v", result.Status)
			CleanupTest(ctx, repo, auction.Id)
			return
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Second)

		var result AuctionEntityMongo
		err := repo.Collection.FindOne(ctx, bson.M{"_id": auction.Id}).Decode(&result)
		if err != nil {
			t.Errorf("Failed to find auction after 10s: %v", err)
			CleanupTest(ctx, repo, auction.Id)
			return
		}

		if result.Status != auction_entity.Completed {
			t.Errorf("Expected auction to be Completed after 2s, got: %v", result.Status)
			CleanupTest(ctx, repo, auction.Id)
		}
	}()

	wg.Wait()
	CleanupTest(ctx, repo, auction.Id)
}

func TestAuctionManualClose(t *testing.T) {
	LoadEnv()

	ctx := context.Background()

	os.Setenv("AUCTION_INTERVAL", "10s")
	os.Setenv("MONGODB_URL", "mongodb://admin:admin@localhost:27017/auctions?authSource=admin")

	databaseConnection, err := mongodb.NewMongoDBConnection(ctx)

	if err != nil {
		log.Fatal(err)
	}
	repo := NewAuctionRepository(databaseConnection)

	auction, _ := auction_entity.CreateAuction(
		"Product Test", "Tests", "A test product created to check if auction is deleted", auction_entity.New,
	)

	repo.CreateAuction(ctx, auction)

	var wg sync.WaitGroup
	wg.Add(2)

	time.Sleep(1 * time.Second)

	var result AuctionEntityMongo
	if err := repo.Collection.FindOne(ctx, bson.M{"_id": auction.Id}).Decode(&result); err != nil {
		t.Errorf("Failed to find auction after 1s: %v", err)
		CleanupTest(ctx, repo, auction.Id)
		return
	}

	if result.Status != auction_entity.Active {
		t.Errorf("Expected auction to be Active after 1s, got: %v", result.Status)
		CleanupTest(ctx, repo, auction.Id)
		return
	}

	*repo.AuctionTimers[auction.Id].EndChan <- true

	time.Sleep(time.Second)

	if err := repo.Collection.FindOne(ctx, bson.M{"_id": auction.Id}).Decode(&result); err != nil {
		t.Errorf("Failed to find auction after 1s: %v", err)
		CleanupTest(ctx, repo, auction.Id)
		return
	}

	if result.Status != auction_entity.Completed {
		t.Errorf("Expected auction to be Completed after manual close, got: %v", result.Status)
		CleanupTest(ctx, repo, auction.Id)
		return
	}
	CleanupTest(ctx, repo, auction.Id)

}
