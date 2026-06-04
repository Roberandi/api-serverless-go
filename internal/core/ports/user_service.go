package ports

import "apigolang/internal/core/domain"

// UserService define las acciones que un usuario puede realizar en nuestro sistema.
type UserService interface {
	RegisterUser(user *domain.User) error
	Login(matricula, password string) (string, error)
	GetAllUsers() ([]domain.User, error)
	GetUserByID(id int) (*domain.User, error)
	UpdateUser(id int, user *domain.User) error
	DeleteUser(id int) error
}
