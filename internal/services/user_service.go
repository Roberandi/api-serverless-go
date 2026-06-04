package services

import (
	"errors"

	"apigolang/internal/adapters/auth" // Asegúrate de que esta ruta sea correcta
	"apigolang/internal/core/domain"
	"apigolang/internal/core/ports"

	"golang.org/x/crypto/bcrypt"
)

// userService es la estructura que implementará nuestra interfaz UserService.
type userService struct {
	repo ports.UserRepository
}

// NewUserService es el constructor de nuestro servicio.
func NewUserService(repo ports.UserRepository) ports.UserService {
	return &userService{
		repo: repo,
	}
}

// RegisterUser contiene la lógica para registrar un nuevo usuario.
func (s *userService) RegisterUser(user *domain.User) error {
	// 1. Encriptar la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error al encriptar la contraseña")
	}

	user.Password = string(hashedPassword)

	// 2. Guardar en la DB
	return s.repo.Create(user)
}

// Login verifica las credenciales y genera un token JWT real.
func (s *userService) Login(matricula, password string) (string, error) {
	// 1. Buscar al usuario
	user, err := s.repo.GetByMatricula(matricula)
	if err != nil {
		return "", errors.New("usuario no encontrado")
	}

	// 2. Comparar contraseña
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// 3. Generar el Token JWT real
	token, err := auth.GenerateToken(user.Matricula)
	if err != nil {
		return "", errors.New("error al generar el token")
	}

	return token, nil
}

func (s *userService) GetAllUsers() ([]domain.User, error) {
	return s.repo.GetAll()
}

func (s *userService) GetUserByID(id int) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) UpdateUser(id int, user *domain.User) error {
	user.ID = id
	return s.repo.Update(user)
}

func (s *userService) DeleteUser(id int) error {
	return s.repo.Delete(id)
}
