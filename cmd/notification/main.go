package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Estructura de lo que recibiremos
type NotificationReq struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// Esta es la función que se despierta cuando llega un mensaje a la cola (SQS)
func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	// Leemos tu correo y contraseña desde AWS
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// SQS nos puede mandar varios mensajes a la vez, los leemos uno por uno
	for _, message := range sqsEvent.Records {

		var req NotificationReq
		// SNS envuelve el mensaje en un JSON grande, aquí lo sacamos
		var snsMsg struct {
			Message string `json:"Message"`
		}
		json.Unmarshal([]byte(message.Body), &snsMsg)
		json.Unmarshal([]byte(snsMsg.Message), &req)

		// Preparamos el correo
		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		to := []string{req.Email}
		msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", req.Email, req.Subject, req.Message))

		// ¡Enviamos el correo!
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, to, msg)
		if err != nil {
			log.Printf("Error enviando correo: %v", err)
			return err
		}
		log.Printf("¡Correo enviado a %s!", req.Email)
	}
	return nil
}

func main() {
	// Iniciamos el "Cocinero"
	lambda.Start(handler)
}
