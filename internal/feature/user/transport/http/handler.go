package http

import (
	"encoding/json"
	"hrms/internal/feature/user/service"
	"hrms/internal/feature/user/transport/http/dto"
	"net/http"
)

// Create interface and store mocked handler in http/mock/handler_mock.go
//
//go:generate mockgen -source=handler.go -package=mock -destination=mock/handler_mock.go -mock_names=UserHandler=MockUserHandler
type UserHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	// Handler takes only different services inside.
	// You can't inject repositories in it.
	// Additionally, you can put here needed utility managers (clock, logger, transaction manager).
	userService service.UserService
}

func NewUserHandler(userService service.UserService) UserHandler {
	return &userHandler{userService: userService}
}

func (h *userHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 1) Validate request DTO
	// If you need very strong validation, you can create Validate() method on DTO struct and use some external package for validation
	userCreateRequestDTO := &dto.UserCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(userCreateRequestDTO); err != nil {
		// Log error
		// Pass error message as the response
		return
	}

	// 2) Convert request DTO to Domain
	userDomain := userCreateRequestDTO.ToDomain()

	// 3) Call service method, userDomain will be fulfilled with needed fields
	err := h.userService.CreateUser(ctx, &userDomain)
	if err != nil {
		// Log error
		// Pass error message as the response
		return
	}

	userCreateResponseDTO := &dto.UserCreateResponse{
		Message: "User was created successfully.",
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(userCreateResponseDTO); err != nil {
		// Log error
		// Pass error message as the response
		return
	}

}
