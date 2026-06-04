package repository

import (
	"database/sql"
	"errors"

	"apigolang/internal/core/domain"
	"apigolang/internal/core/ports"
)

// postgresUserRepository es nuestro adaptador para hablar con PostgreSQL.
type postgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository crea una nueva conexión a este repositorio.
func NewPostgresUserRepository(db *sql.DB) ports.UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

// Create inserta un nuevo usuario en la base de datos.
func (r *postgresUserRepository) Create(user *domain.User) error {
	// Escribimos la consulta SQL
	query := `INSERT INTO users (matricula, nombre, password) VALUES ($1, $2, $3) RETURNING id`

	// Ejecutamos la consulta y guardamos el ID generado en nuestro usuario
	err := r.db.QueryRow(query, user.Matricula, user.Nombre, user.Password).Scan(&user.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetByMatricula busca un usuario usando su número de matrícula.
func (r *postgresUserRepository) GetByMatricula(matricula string) (*domain.User, error) {
	query := `SELECT id, matricula, nombre, password FROM users WHERE matricula = $1`
	row := r.db.QueryRow(query, matricula)

	var user domain.User
	// Leemos los datos de la fila y los metemos en nuestra variable 'user'
	err := row.Scan(&user.ID, &user.Matricula, &user.Nombre, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("usuario no encontrado")
		}
		return nil, err
	}

	return &user, nil
}

// GetAll trae todos los usuarios (sin incluir sus contraseñas por seguridad)
func (r *postgresUserRepository) GetAll() ([]domain.User, error) {
	rows, err := r.db.Query(`SELECT id, matricula, nombre FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Matricula, &u.Nombre); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

// GetByID busca un usuario específico por su ID
func (r *postgresUserRepository) GetByID(id int) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(`SELECT id, matricula, nombre FROM users WHERE id = $1`, id).Scan(&u.ID, &u.Matricula, &u.Nombre)
	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}
	return &u, nil
}

// Update modifica los datos de un usuario
func (r *postgresUserRepository) Update(user *domain.User) error {
	_, err := r.db.Exec(`UPDATE users SET matricula = $1, nombre = $2 WHERE id = $3`, user.Matricula, user.Nombre, user.ID)
	return err
}

// Delete elimina un usuario de la base de datos
func (r *postgresUserRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}
