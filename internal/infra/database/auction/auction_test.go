// File: internal/infra/database/auction/create_auction_test.go
package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	mongo_options "go.mongodb.org/mongo-driver/mongo/options"
)

// -- Mocks --

type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*mongo_options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollection) UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...*mongo_options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, id, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*mongo_options.FindOneOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockCollection) Find(ctx context.Context, filter interface{}, opts ...*mongo_options.FindOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

// MockDatabase implements IMongoDatabase
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Collection(name string) IMongoCollection {
	args := m.Called(name)
	return args.Get(0).(IMongoCollection)
}

func TestCreateAuction_AutoClosesAfterDuration(t *testing.T) {
	mockCollection := new(MockCollection)
	mockDB := new(MockDatabase)

	mockDB.On("Collection", "auctions").Return(mockCollection)

	repo := NewAuctionRepository(&mongo.Database{}) // Now takes IMongoDatabase interface

	auction, _ := auction_entity.CreateAuction("Product Test", "Test", "Test Description", auction_entity.New)

	// Setup expectations
	mockCollection.
		On("InsertOne", mock.Anything, mock.Anything).
		Return(&mongo.InsertOneResult{}, nil)

	mockCollection.
		On("UpdateByID", mock.Anything, auction.Id, mock.Anything).
		Return(&mongo.UpdateResult{}, nil)

	// Set short auction duration
	t.Setenv("AUCTION_DURATION", "1s")

	err := repo.CreateAuction(context.Background(), auction)
	assert.Nil(t, err)

	// Wait for the auction to close automatically
	time.Sleep(1500 * time.Millisecond)

	mockCollection.AssertCalled(t, "InsertOne", mock.Anything, mock.Anything)
	mockCollection.AssertCalled(t, "UpdateByID", mock.Anything, auction.Id, mock.Anything)
}
