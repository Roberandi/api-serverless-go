package domain

// User representa a nuestra entidad principal en el sistema.
type User struct {
	ID        int    `json:"id"`
	Matricula string `json:"matricula"`
	Nombre    string `json:"nombre"`
	Password  string `json:"password,omitempty"`
}
