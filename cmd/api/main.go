package main

import (
	"context"
	"database/sql"
	"encoding/json" // <-- NUEVO: Para convertir struct a JSON texto
	"log"
	"os"

	"apigolang/internal/adapters/handlers"
	"apigolang/internal/adapters/repository"
	"apigolang/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"         // <-- NUEVO: Librería oficial de AWS
	"github.com/aws/aws-sdk-go/aws/session" // <-- NUEVO: Para crear sesiones con AWS
	"github.com/aws/aws-sdk-go/service/sns" // <-- NUEVO: Para interactuar con SNS
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Variable global para el traductor de Lambda
var ginLambda *ginadapter.GinLambda

// NUEVO: Estructura obligatoria para recibir la notificación desde la App móvil
type NotificationReq struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// La función init() se ejecuta una sola vez cuando el contenedor de Lambda se enciende
func init() {
	// 1. Conexión a la Base de Datos (Neon en la Nube)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgresql://neondb_owner:npg_9Xif8TNzQJrG@ep-wild-heart-apfawq27-pooler.c-7.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error al conectar con la base de datos: ", err)
	}

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

	os.MkdirAll("/tmp/uploads", os.ModePerm)
	router.Static("/uploads", "/tmp/uploads")

	// 4. Definir las rutas públicas
	router.POST("/register", httpHandler.Register)
	router.POST("/login", httpHandler.Login)

	// NUEVO: Endpoint solicitado por la asignación para enviar la notificación
	router.POST("/notifications/send", sendNotification)

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

// NUEVO: Función encargada de agarrar los datos del correo y publicarlos en AWS SNS
func sendNotification(c *gin.Context) {
	var req NotificationReq
	// Validamos que el JSON enviado por el cliente sea correcto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Datos inválidos"})
		return
	}

	// 1. Convertimos la estructura a un string JSON limpio
	msgBytes, err := json.Marshal(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error al procesar el mensaje"})
		return
	}
	msgString := string(msgBytes)

	// 2. Iniciamos sesión segura usando los permisos nativos de la Lambda
	sess := session.Must(session.NewSession())
	svc := sns.New(sess)

	// Esta variable la creará Terraform automáticamente más adelante
	topicArn := os.Getenv("SNS_TOPIC_ARN")

	// 3. Publicamos el mensaje directamente en el "Altoparlante" (SNS Topic)
	_, err = svc.Publish(&sns.PublishInput{
		Message:  aws.String(msgString),
		TopicArn: aws.String(topicArn),
	})

	if err != nil {
		// Si falla, imprimimos el error real en los logs de CloudWatch
		log.Printf("Error al publicar en SNS: %v", err)
		c.JSON(500, gin.H{"error": "Error al enviar mensaje"})
		return
	}

	// Respuesta exitosa obligatoria para la App móvil
	c.JSON(200, gin.H{"message": "Mensaje enviado correctamente."})
}

// Handler es la función que AWS Lambda ejecutará cada vez que llegue una petición
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp, err := ginLambda.ProxyWithContext(ctx, req)
	if err != nil {
		log.Printf("ERROR CRÍTICO EN LAMBDA: %v", err)
	}
	return resp, err
}

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(Handler)
	} else {
		log.Println("API configurada para ejecutarse en AWS Lambda. El modo local estándar está desactivado.")
	}
}
