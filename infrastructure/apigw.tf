resource "aws_apigatewayv2_api" "maroon_api" {
  name          = "MaroonApi"
  protocol_type = "HTTP"
  target        = aws_lambda_function.maroon_api_lambda.arn
}

resource "aws_apigatewayv2_domain_name" "maroon_api_domain_name" {
  domain_name = "${local.dns_name}.maroon.gtech.dev"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.api_cert_validation.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
  depends_on = [aws_acm_certificate_validation.api_cert_validation]
}

resource "aws_apigatewayv2_api_mapping" "maroon_api_domain_mapping" {
  api_id      = aws_apigatewayv2_api.maroon_api.id
  domain_name = aws_apigatewayv2_domain_name.maroon_api_domain_name.id
  stage       = "$default"
}

resource "aws_apigatewayv2_authorizer" "maroon_api_authorizer" {
  name             = "CognitoJWTAuth"
  api_id           = aws_apigatewayv2_api.maroon_api.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  jwt_configuration {
    audience = var.audiences
    issuer   = "https://cognito-idp.${data.aws_region.current.name}.amazonaws.com/${var.cognito_user_pool_id}"
  }
}

resource "aws_lambda_permission" "api" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.maroon_api_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.maroon_api.execution_arn}/*/*"
}