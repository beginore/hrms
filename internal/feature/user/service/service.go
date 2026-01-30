package service

import (
	"context"
	domain "hrms/internal/domain/user"
	"hrms/internal/feature/user/repository/postgres"
	postgresModel "hrms/internal/feature/user/repository/postgres/model"
)

//go:generate mockgen -source=service.go -package=mock -destination=mock/service_mock.go -mock_names=UserService=MockUserService
type UserService interface {
	CreateUser(ctx context.Context, userDomain *domain.User) error
}

type userService struct {
	// Service can take different services inside.
	// But repositories that are injected here can be only from this module.
	// If you need repository method from different module, call service method of that module
	// in some service method here.
	// If the logic is complex, for example you need to call 2 other service methods, there is no problem with that,
	// you just write one aggregating service method here that calls service methods from different methods.
	userPostgresRepository postgres.UserRepository
}

func NewUserService(userPostgresRepository postgres.UserRepository) UserService {
	return &userService{userPostgresRepository: userPostgresRepository}
}

func (s *userService) CreateUser(ctx context.Context, userDomain *domain.User) error {
	// 0) Do some validations
	if userDomain.Email == "taken_email@gmail.com" {
		return domain.NewUserAlreadyExistsError()
	}

	// 1) Prepare Models from Domain
	postgresUserModel := &postgresModel.User{
		Email: userDomain.Email,
	}
	postgresUserInfoModel := &postgresModel.UserInfo{
		Firstname: userDomain.FirstName,
		Lastname:  userDomain.LastName,
	}
	// 2) Do required logic

	// These 2 operations should be in TX, but for the example I skipped it
	if err := s.userPostgresRepository.CreateUser(ctx, postgresUserModel); err != nil {
		return err
	}
	postgresUserInfoModel.UserID = postgresUserModel.ID
	if err := s.userPostgresRepository.CreateUserInfo(ctx, postgresUserInfoModel); err != nil {
		return err
	}

	// 3) Update Domain with other fields
	userDomain.ID = postgresUserModel.ID

	return nil
}
