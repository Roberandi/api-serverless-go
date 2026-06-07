# 1. Crear un Rol de IAM para que Lambda tenga permisos de ejecutarse
resource "aws_iam_role" "lambda_exec" {
  name = "serverless_api_role_v4" # <-- CAMBIO AQUÍ
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
  function_name    = "apigolang-serverless-v4" # <-- CAMBIO AQUÍ
  filename         = "../bootstrap.zip"
  handler          = "bootstrap"
  source_code_hash = filebase64sha256("../bootstrap.zip")
  role             = aws_iam_role.lambda_exec.arn
  runtime          = "provided.al2"
  timeout          = 30

  environment {
    variables = {
      DATABASE_URL = var.database_url
      JWT_SECRET   = var.jwt_secret
      GIN_MODE     = "release"
    }
  }
}

# 3. Crear el API Gateway
resource "aws_apigatewayv2_api" "http_api" {
  name          = "apigolang-http-api-v4" # <-- CAMBIO AQUÍ
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

# 5. Crear el Grupo de Logs en CloudWatch
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.api_lambda.function_name}"
  retention_in_days = 14
}