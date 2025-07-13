package integration

import (
	"context"
	"database/sql"
	"testing"

	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_ "github.com/mattn/go-sqlite3"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ActorTestSuite struct {
	suite.Suite
	db           *sql.DB
	actorService *services.ActorService
}

// Make sure that database and services are set up before each test
func (suite *ActorTestSuite) SetupSuite() {
	suite.db = SetupTestDatabase()
	querier := store.New(suite.db)
	suite.actorService = services.NewActorService(querier)
}

func (suite *ActorTestSuite) TearDownSuite() {
	// RunDownMigrations()
	suite.db.Close()
}

// All methods that begin with "Test" are run as tests within a suite.
func (suite *ActorTestSuite) TestCreateActor() {
	metadata := `{"auth_strategy": "password", "credentials": {"username": "test", "password": "test"}, "is_active": true}`
	actorParams := &store.CreateActorParams{
		Type:        "user",
		Name:        "test_user",
		DisplayName: sql.NullString{String: "Test User", Valid: true},
		AvatarUrl:   sql.NullString{String: "", Valid: false},
		IsActive:    sql.NullBool{Bool: true, Valid: true},
		Metadata:    sql.NullString{String: metadata, Valid: true},
		CreatedBy:   1, // System actor ID
		UpdatedBy:   1,
	}

	err := suite.actorService.CreateActor(context.Background(), actorParams.Type, actorParams.Name, actorParams.DisplayName.String, actorParams.AvatarUrl.String, actorParams.Metadata.String, actorParams.IsActive.Bool, actorParams.CreatedBy, actorParams.UpdatedBy)
	assert.NoError(suite.T(), err)

	actor, err := suite.actorService.GetActorByName(context.Background(), actorParams.Name, actorParams.Type)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), actorParams.Name, actor.Name)
	assert.Equal(suite.T(), actorParams.Type, actor.Type)
	assert.Equal(suite.T(), actorParams.DisplayName.String, actor.DisplayName.String)
}

func (suite *ActorTestSuite) TestGetActorByName() {
	actor, err := suite.actorService.GetActorByName(context.Background(), "test_user", "user")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test_user", actor.Name)
	assert.Equal(suite.T(), "user", actor.Type)
	assert.Equal(suite.T(), "Test User", actor.DisplayName.String)
}

func (suite *ActorTestSuite) TestGetActorByID() {
	actor, err := suite.actorService.GetActorByID(context.Background(), 1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test_user", actor.Name)
	assert.Equal(suite.T(), "user", actor.Type)
	assert.Equal(suite.T(), "Test User", actor.DisplayName.String)
}

// TODO: follow up on this for proper order of tests and the teardown
// func (suite *ActorTestSuite) TestDeleteActor() {
// 	err := suite.actorService.DeleteActor(context.Background(), 1)
// 	assert.NoError(suite.T(), err)
// }

// TODO: add tests for updating actor details and for their auth handling
func TestActorTestSuite(t *testing.T) {
	suite.Run(t, new(ActorTestSuite))
}
