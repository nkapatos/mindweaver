package services

import (
	"context"
	"log/slog"

	"github.com/nkapatos/mindweaver/internal/store"
)

type UserService struct {
	userStore store.Querier
	logger    *slog.Logger
}

func NewUserService(userStore store.Querier) *UserService {
	return &UserService{
		userStore: userStore,
		logger:    slog.Default(),
	}
}

// CreateUser creates a new user with the given name
func (s *UserService) CreateUser(ctx context.Context, name string) error {
	s.logger.Info("Creating new user", "name", name)

	if err := s.userStore.CreateUser(ctx, name); err != nil {
		s.logger.Error("Failed to create user", "name", name, "error", err)
		return err
	}

	s.logger.Info("User created successfully", "name", name)
	return nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(ctx context.Context, id int64) (store.User, error) {
	s.logger.Debug("Getting user by ID", "id", id)

	user, err := s.userStore.GetUserById(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user by ID", "id", id, "error", err)
		return store.User{}, err
	}

	s.logger.Debug("User retrieved successfully", "id", id, "name", user.Name)
	return user, nil
}

// GetUserByName retrieves a user by their name
func (s *UserService) GetUserByName(ctx context.Context, name string) (store.User, error) {
	s.logger.Debug("Getting user by name", "name", name)

	user, err := s.userStore.GetUserByName(ctx, name)
	if err != nil {
		s.logger.Error("Failed to get user by name", "name", name, "error", err)
		return store.User{}, err
	}

	s.logger.Debug("User retrieved successfully", "name", name, "id", user.ID)
	return user, nil
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers(ctx context.Context) ([]store.User, error) {
	s.logger.Debug("Getting all users")

	users, err := s.userStore.GetAllUsers(ctx)
	if err != nil {
		s.logger.Error("Failed to get all users", "error", err)
		return nil, err
	}

	s.logger.Debug("All users retrieved successfully", "count", len(users))
	return users, nil
}

// UpdateUser updates a user's name by their ID
func (s *UserService) UpdateUser(ctx context.Context, id int64, name string) error {
	s.logger.Info("Updating user", "id", id, "new_name", name)

	params := store.UpdateUserParams{
		ID:   id,
		Name: name,
	}

	if err := s.userStore.UpdateUser(ctx, params); err != nil {
		s.logger.Error("Failed to update user", "id", id, "new_name", name, "error", err)
		return err
	}

	s.logger.Info("User updated successfully", "id", id, "new_name", name)
	return nil
}

// DeleteUser deletes a user by their ID
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	s.logger.Info("Deleting user", "id", id)

	if err := s.userStore.DeleteUser(ctx, id); err != nil {
		s.logger.Error("Failed to delete user", "id", id, "error", err)
		return err
	}

	s.logger.Info("User deleted successfully", "id", id)
	return nil
}
