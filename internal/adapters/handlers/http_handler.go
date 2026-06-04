package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"apigolang/internal/core/domain"
	"apigolang/internal/core/ports"

	"github.com/gin-gonic/gin"
)

// HTTPHandler es el "recepcionista" de nuestra API web.
type HTTPHandler struct {
	userService ports.UserService
}

func NewHTTPHandler(userService ports.UserService) *HTTPHandler {
	return &HTTPHandler{
		userService: userService,
	}
}

// Register maneja la ruta POST /register
func (h *HTTPHandler) Register(c *gin.Context) {
	var user domain.User

	// Leemos el JSON que envía el usuario y lo convertimos a nuestra estructura
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Llamamos a la lógica de negocio
	err := h.userService.RegisterUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado exitosamente"})
}

// Login maneja la ruta POST /login
func (h *HTTPHandler) Login(c *gin.Context) {
	// Estructura temporal solo para leer las credenciales
	var creds struct {
		Matricula string `json:"matricula"`
		Password  string `json:"password"`
	}

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	token, err := h.userService.Login(creds.Matricula, creds.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetAll devuelve la lista de todos los usuarios
func (h *HTTPHandler) GetAll(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetByID devuelve un solo usuario
func (h *HTTPHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id")) // Convierte el ID de la URL a número
	user, err := h.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Update actualiza un usuario
func (h *HTTPHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	if err := h.userService.UpdateUser(id, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Usuario actualizado exitosamente"})
}

// Delete borra un usuario
func (h *HTTPHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.userService.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Usuario eliminado"})
}

// UploadFile recibe un archivo desde la app móvil y lo guarda en el servidor
func (h *HTTPHandler) UploadFile(c *gin.Context) {
	// 1. Buscamos el archivo en la petición (la app lo enviará con el nombre "file")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se recibió ningún archivo"})
		return
	}

	// 2. Creamos un nombre único usando la fecha y hora para que no se sobrescriban
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))

	// Cambia esto:
	// savePath := fmt.Sprintf("uploads/%s", filename)

	// Por esto:
	savePath := fmt.Sprintf("/tmp/uploads/%s", filename)

	// 3. Guardamos el archivo físicamente
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el archivo"})
		return
	}

	// 4. Respondemos con éxito y le damos a la app la URL para ver el archivo
	c.JSON(http.StatusOK, gin.H{
		"message": "Archivo subido exitosamente",
		"url":     fmt.Sprintf("http://localhost:8080/%s", savePath),
	})
}
