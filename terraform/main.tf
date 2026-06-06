# 1. Crear un Rol de IAM para que Lambda tenga permisos de ejecutarse y escribir Logs
resource "aws_iam_role" "lambda_exec" {
  name = "serverless_api_role"
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

# 2. Definir la Función Lambda
resource "aws_lambda_function" "api_lambda" {
  function_name    = "apigolang-serverless"
  # Asumimos que GitHub Actions empaquetará tu código de Go en un zip llamado bootstrap.zip
  filename         = "../bootstrap.zip"
  handler          = "bootstrap"
  source_code_hash = filebase64sha256("../bootstrap.zip")
  role             = aws_iam_role.lambda_exec.arn
  # Usamos el entorno más moderno y rápido de AWS para Go
  runtime          = "provided.al2"

  # Variables de entorno que viviran dentro de AWS Lambda
  environment {
    variables = {
      DATABASE_URL = var.database_url
      JWT_SECRET   = var.jwt_secret
      GIN_MODE     = "release"
    }
  }
}

# 3. Crear el API Gateway (El Recepcionista web)
resource "aws_apigatewayv2_api" "http_api" {
  name          = "apigolang-http-api"
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

# Conectar todas las rutas a nuestro código en Go
resource "aws_apigatewayv2_route" "default_route" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

# 4. Dar permiso al API Gateway para que despierte a tu Lambda
resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}

# 5. Crear el Grupo de Logs en CloudWatch
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.api_lambda.function_name}"
  retention_in_days = 14
}