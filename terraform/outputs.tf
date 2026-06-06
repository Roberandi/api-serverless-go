output "api_gateway_url" {
  description = "URL pública de tu API Serverless"
  value       = aws_apigatewayv2_stage.default.invoke_url
}