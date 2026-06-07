package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"apigolang/internal/adapters/handlers"
	"apigolang/internal/adapters/repository"
	"apigolang/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Variable global para el traductor de Lambda
var ginLambda *ginadapter.GinLambda

// La función init() se ejecuta una sola vez cuando el contenedor de Lambda se enciende
func init() {
	// 1. Conexión a la Base de Datos (Neon en la Nube)
	// Intentamos leer la URL de las variables de entorno (GitHub Secrets nos dará esto)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Si no hay variable, usamos tu base de datos de Neon por defecto
		connStr = "postgresql://neondb_owner:npg_9Xif8TNzQJrG@ep-wild-heart-apfawq27-pooler.c-7.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error al conectar con la base de datos: ", err)
	}
	// IMPORTANTE: En Lambda no ponemos defer db.Close() porque queremos reusar la conexión

	// 2. Armando las capas (Inyección de Dependencias)
	userRepo := repository.NewPostgresUserRepository(db)
	userService := services.NewUserService(userRepo)
	httpHandler := handlers.NewHTTPHandler(userService)

	// 3. Configurar el Servidor Web (Router)
	router := gin.Default()

	// Agregar middleware de CORS manual
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	// En AWS Lambda, la única carpeta donde tenemos permiso de escritura es /tmp
	os.MkdirAll("/tmp/uploads", os.ModePerm)
	router.Static("/uploads", "/tmp/uploads")

	// 4. Definir las rutas públicas
	router.POST("/register", httpHandler.Register)
	router.POST("/login", httpHandler.Login)

	// 5. Definir las rutas protegidas
	protected := router.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.GET("/users", httpHandler.GetAll)
		protected.GET("/users/:id", httpHandler.GetByID)
		protected.PUT("/users/:id", httpHandler.Update)
		protected.DELETE("/users/:id", httpHandler.Delete)

		protected.POST("/upload", httpHandler.UploadFile)
	}

	// 6. Envolver el router de Gin con el adaptador de Lambda
	ginLambda = ginadapter.New(router)
}

// Handler es la función que AWS Lambda ejecutará cada vez que llegue una petición
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Intentamos procesar la petición
	resp, err := ginLambda.ProxyWithContext(ctx, req)

	// Si hay un error, lo registramos en CloudWatch para poder verlo
	if err != nil {
		log.Printf("ERROR CRÍTICO EN LAMBDA: %v", err)
	}

	return resp, err
}

func main() {
	// AWS inyecta esta variable automáticamente. Si existe, corremos en modo Lambda.
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(Handler)
	} else {
		// Mensaje por si intentas correrlo localmente
		log.Println("API configurada para ejecutarse en AWS Lambda. El modo local estándar está desactivado.")
	}
}
