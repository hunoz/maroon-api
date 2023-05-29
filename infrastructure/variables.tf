variable "cloudflare_api_token" {
  type = string
}

variable "cloudflare_zone_id" {
  type = string
}

variable "cognito_user_pool_id" {
  type = string
}

variable "cognito_region" {
  type = string
}

variable "audiences" {
  type = list(any)
}