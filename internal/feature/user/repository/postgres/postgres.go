package postgres

import (
	"context"
	"hrms/internal/feature/user/repository/postgres/model"
	"hrms/internal/infrastructure/storage/postgres"
)

//go:generate mockgen -source=interface.go -package=mock -destination=mock/repository_mock.go -mock_names=UserRepository=MockUserRepository
type UserRepository interface {
	CreateUser(ctx context.Context, userModel *model.User) error
	CreateUserInfo(ctx context.Context, userInfo *model.UserInfo) error
}

type userRepository struct {
	// Here you pass only interface to work with DB usually

	// To work with transactions, you should use some transaction manager. It will take DB pointer in the bootstrap of the app.
	// Then you will inject transaction manager in needed service.
	// After that, you will declare methods with postfix TX only (e.g. CreateUserTX) It means this method is used inside existing TX.
	// So you shouldn't create TXs there and work with passed TX that is usually created by called transaction manager.
	database postgres.Database
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) CreateUser(ctx context.Context, userModel *model.User) error {
	return nil
}

func (r *userRepository) CreateUserInfo(ctx context.Context, userInfoModel *model.UserInfo) error {
	return nil
}
