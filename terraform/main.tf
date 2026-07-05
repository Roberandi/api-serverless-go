# 1. Crear un Rol de IAM para que Lambda tenga permisos de ejecutarse
resource "aws_iam_role" "lambda_exec" {
  name = "serverless_api_role_v8"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# 2. Definir la Función Lambda Principal (Backend)
resource "aws_lambda_function" "api_lambda" {
  function_name    = "apigolang-serverless-v8"
  filename         = "../bootstrap.zip"
  handler          = "bootstrap"
  source_code_hash = filebase64sha256("../bootstrap.zip")
  role             = aws_iam_role.lambda_exec.arn
  runtime          = "provided.al2"
  timeout          = 30

  environment {
    variables = {
      DATABASE_URL  = var.database_url
      JWT_SECRET    = var.jwt_secret
      GIN_MODE      = "release"
      # NUEVO: Le inyectamos la dirección del Altoparlante SNS
      SNS_TOPIC_ARN = aws_sns_topic.notif_topic.arn
    }
  }
}

# 3. Crear el API Gateway
resource "aws_apigatewayv2_api" "http_api" {
  name          = "apigolang-http-api-v8"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id             = aws_apigatewayv2_api.http_api.id
  integration_type   = "AWS_PROXY"
  integration_uri    = aws_lambda_function.api_lambda.invoke_arn
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "default_route" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

# 4. Dar permiso al API Gateway
resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}

# 5. Crear el Grupo de Logs en CloudWatch (Lambda Principal)
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.api_lambda.function_name}"
  retention_in_days = 14
}

# ==============================================================================
# 6. INFRAESTRUCTURA DE MENSAJERÍA DESACOPLADA (NUEVO)
# ==============================================================================

# A. Crear el "Altoparlante" (SNS Topic)
resource "aws_sns_topic" "notif_topic" {
  name = "email-notifications-topic-v8"
}

# Dar permiso a la Lambda Principal para publicar en el Altoparlante
resource "aws_iam_role_policy" "lambda_sns_policy" {
  name   = "lambda_sns_publish_policy_v8"
  role   = aws_iam_role.lambda_exec.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "sns:Publish"
      Effect   = "Allow"
      Resource = aws_sns_topic.notif_topic.arn
    }]
  })
}

# B. Crear la "Barra de Pedidos / Cola" (SQS Queue)
resource "aws_sqs_queue" "notif_queue" {
  name = "email-notifications-queue-v8"
}

# Dar permiso para que el Altoparlante SNS deje mensajes en la Cola SQS
resource "aws_sqs_queue_policy" "sns_to_sqs" {
  queue_url = aws_sqs_queue.notif_queue.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "sns.amazonaws.com" }
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.notif_queue.arn
      Condition = {
        ArnEquals = { "aws:SourceArn" = aws_sns_topic.notif_topic.arn }
      }
    }]
  })
}

# Conectar el Altoparlante SNS con la Cola SQS (Suscripción)
resource "aws_sns_topic_subscription" "notif_sub" {
  topic_arn = aws_sns_topic.notif_topic.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.notif_queue.arn
}

# C. Crear el "Cocinero" (Lambda de Notificaciones)

# Rol para que la Lambda de notificaciones pueda ejecutarse y leer de SQS
resource "aws_iam_role" "notif_lambda_exec" {
  name = "notification_lambda_role_v8"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}

# Permiso oficial de AWS para leer colas SQS
resource "aws_iam_role_policy_attachment" "notif_sqs_policy" {
  role       = aws_iam_role.notif_lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole"
}

# La Función Lambda que envía los correos
resource "aws_lambda_function" "notification_lambda" {
  function_name    = "notification-lambda-v8"
  filename         = "../notification.zip"
  handler          = "bootstrap"
  source_code_hash = filebase64sha256("../notification.zip")
  role             = aws_iam_role.notif_lambda_exec.arn
  runtime          = "provided.al2"
  timeout          = 30

  environment {
    variables = {
      SMTP_USER = var.smtp_user
      SMTP_PASS = var.smtp_pass
    }
  }
}

# Conectar la Cola SQS como "gatillo" (Trigger) para despertar a la Lambda
resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.notif_queue.arn
  function_name    = aws_lambda_function.notification_lambda.arn
  batch_size       = 1
}

# Grupo de Logs en CloudWatch exclusivo para los correos enviados
resource "aws_cloudwatch_log_group" "notif_logs" {
  name              = "/aws/lambda/${aws_lambda_function.notification_lambda.function_name}"
  retention_in_days = 14
}

# ==============================================================================
# 7. AUTOMATIZACIÓN CON AMAZON EVENTBRIDGE SCHEDULER
# ==============================================================================

# Rol de IAM para que EventBridge Scheduler pueda publicar en SNS
resource "aws_iam_role" "eventbridge_sns_role" {
  name = "eventbridge_sns_publish_role_v7"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "scheduler.amazonaws.com"
      }
    }]
  })
}

# Política para permitir la publicación en el Tópico SNS específico
resource "aws_iam_role_policy" "eventbridge_sns_policy" {
  name   = "eventbridge_sns_publish_policy_v7"
  role   = aws_iam_role.eventbridge_sns_role.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "sns:Publish"
      Effect   = "Allow"
      Resource = aws_sns_topic.notif_topic.arn
    }]
  })
}

# Creación de la regla del Scheduler utilizando una expresión rate
resource "aws_scheduler_schedule" "five_minute_sns_trigger" {
  name       = "trigger-sns-every-5-minutes"
  group_name = "default"

  # Expresión RATE utilizada: Ejecutar cada 5 minutos
  schedule_expression = "rate(5 minutes)"
  
  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = aws_sns_topic.notif_topic.arn
    role_arn = aws_iam_role.eventbridge_sns_role.arn
    
    # Mensaje que se enviará automáticamente
    input = jsonencode({
      email   = "admin@tu-sistema.com",
      subject = "Reporte Automatizado - EventBridge",
      message = "Este es un evento programado por EventBridge ejecutándose cada 5 minutos."
    })
  }
}