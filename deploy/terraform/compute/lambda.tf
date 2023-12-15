data "aws_iam_policy_document" "assume_lambda_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

// create lambda role, that lambda function can assume (use)
resource "aws_iam_role" "lambda" {
  name               = "AssumeLambdaRoleForStateMachine-${terraform.workspace}"
  description        = "Role for lambda to assume lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role.json
}

data "aws_iam_policy_document" "allow_lambda_logging" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      "arn:aws:logs:*:*:*",
    ]
  }
}

resource "aws_iam_policy" "function_logging_policy" {
  name        = "AllowLambdaLoggingPolicy-${terraform.workspace}"
  description = "Policy for lambda cloudwatch logging"
  policy      = data.aws_iam_policy_document.allow_lambda_logging.json
}

// attach policy to out created lambda role
resource "aws_iam_role_policy_attachment" "lambda_logging_policy_attachment" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_logging_policy.arn
}



# allow lambda to invoke state machine
data "aws_iam_policy_document" "allow_lambda_invoke" {
  statement {
    effect = "Allow"
    actions = [
      "states:StartExecution",
      "lambda:InvokeFunction",
    ]

    resources = [
      aws_sfn_state_machine.state_machine.arn,
    ]
  }
}

resource "aws_iam_policy" "function_invoke_policy" {
  name        = "AllowLambdaInvokePolicy"
  description = "Policy for lambda to invoke state machine"
  policy      = data.aws_iam_policy_document.allow_lambda_invoke.json
}

resource "aws_iam_role_policy_attachment" "lambda_invoke_policy_attachment" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_invoke_policy.arn
}
