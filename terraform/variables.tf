variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "us-east-1"
}

variable "database_url" {
  description = "URL de la base de datos PostgreSQL en Neon"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "Secreto para la firma de JWT"
  type        = string
  sensitive   = true
}

variable "smtp_user" {
  description = "Usuario SMTP para correos"
  type        = string
  sensitive   = true
}

variable "smtp_pass" {
  description = "Contraseña SMTP para correos"
  type        = string
  sensitive   = true
}