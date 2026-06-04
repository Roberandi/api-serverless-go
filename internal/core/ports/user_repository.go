package ports

import "apigolang/internal/core/domain"

// UserRepository define cómo la aplicación se comunicará con la base de datos.
type UserRepository interface {
	Create(user *domain.User) error
	GetByMatricula(matricula string) (*domain.User, error)
	GetAll() ([]domain.User, error)
	GetByID(id int) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id int) error
}
