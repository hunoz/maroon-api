resource "aws_iam_access_key" "maroon_api_user_key" {
  user = aws_iam_user.maroon_api_user.name
}

resource "aws_iam_user" "maroon_api_user" {
  name = "maroon-api"
}

resource "aws_secretsmanager_secret" "maroon_api_user_key_secret" {
  name = "MaroonApiIamUser"
}

resource "aws_secretsmanager_secret_version" "marron_api_user_key_secret_version" {
  secret_id     = aws_secretsmanager_secret.maroon_api_user_key_secret.id
  secret_string = jsonencode({ "AccessKeyId" : aws_iam_access_key.maroon_api_user_key.id, "SecretAccessKey" : aws_iam_access_key.maroon_api_user_key.secret })
}

resource "aws_iam_policy" "maroon_api_user_sts_policy" {
  name        = "AllowAssumeRole"
  description = "Allows the user to assume any role anywhere"
  policy      = jsonencode({
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "sts:AssumeRole",
            "Resource": "*"
        }
    ]
  })
}

resource "aws_iam_user_policy_attachment" "maroon_api_user_sts_policy_attachment" {
  user       = aws_iam_user.maroon_api_user.name
  policy_arn = aws_iam_policy.maroon_api_user_sts_policy.arn
}