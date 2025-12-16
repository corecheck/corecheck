locals {
  bench_array_size = 10

  state_machine_lambdas = [
    "github-sync",
    "migrate",
    "handle-coverage",
    "handle-benchmarks",
    "handle-mutation",
    "rerun-all",
    "stats",
  ]
  # create a map of lambdas and their environment variables
  lambda_overrides = {
    "github-sync" = {
      timeout     = 900
      memory_size = 128
      ephemeral_storage_size = 512
      environment = {
        variables = {
          DATABASE_HOST     = var.db_host
          DATABASE_PORT     = var.db_port
          DATABASE_USER     = var.db_user
          DATABASE_PASSWORD = var.db_password
          DATABASE_NAME     = var.db_database

          GITHUB_ACCESS_TOKEN = var.github_token
          STATE_MACHINE_ARN  = aws_sfn_state_machine.state_machine.arn
          MUTATION_STATE_MACHINE_ARN = aws_sfn_state_machine.mutation_state_machine.arn
        }
      }
    },
    "migrate" = {
      timeout     = 60
      memory_size = 128
      ephemeral_storage_size = 512
      environment = {
        variables = {
          DATABASE_HOST     = var.db_host
          DATABASE_PORT     = var.db_port
          DATABASE_USER     = var.db_user
          DATABASE_PASSWORD = var.db_password
          DATABASE_NAME     = var.db_database
        }
      }
    },
    "handle-coverage" = {
      timeout     = 900
      memory_size = 2048
      ephemeral_storage_size = 2048
      environment = {
        variables = {
          DATABASE_HOST     = var.db_host
          DATABASE_PORT     = var.db_port
          DATABASE_USER     = var.db_user
          DATABASE_PASSWORD = var.db_password
          DATABASE_NAME     = var.db_database

          GITHUB_ACCESS_TOKEN = var.github_token

          BUCKET_DATA_URL = var.corecheck_data_bucket_url
        }
      }
    },
    "handle-benchmarks" = {
      timeout     = 900
      memory_size = 128
      ephemeral_storage_size = 512
      environment = {
        variables = {
          DATABASE_HOST     = var.db_host
          DATABASE_PORT     = var.db_port
          DATABASE_USER     = var.db_user
          DATABASE_PASSWORD = var.db_password
          DATABASE_NAME     = var.db_database

          BENCH_ARRAY_SIZE = local.bench_array_size
          BUCKET_DATA_URL = var.corecheck_data_bucket_url
          DD_API_KEY = var.datadog_api_key
        }
      }
    },
    "handle-mutation" = {
      timeout     = 300
      memory_size = 128
      ephemeral_storage_size = 512
      environment = {
        variables = {
          DATABASE_HOST     = var.db_host
          DATABASE_PORT     = var.db_port
          DATABASE_USER     = var.db_user
          DATABASE_PASSWORD = var.db_password
          DATABASE_NAME     = var.db_database
        }
      }
    },
    "rerun-all" = {
      timeout     = 900
      memory_size = 128
      ephemeral_storage_size = 512
      environment = {
        variables = {
          DATABASE_HOST     = var.db_host
          DATABASE_PORT     = var.db_port
          DATABASE_USER     = var.db_user
          DATABASE_PASSWORD = var.db_password
          DATABASE_NAME     = var.db_database

          BUCKET_DATA = var.corecheck_data_bucket
          BUCKET_DATA_REGION = var.corecheck_data_bucket_region
          STATE_MACHINE_ARN  = aws_sfn_state_machine.state_machine.arn
        }
      }
    },
    "stats" = {
      timeout     = 900
      memory_size = 1024
      ephemeral_storage_size = 10240
      
      environment = {
        variables = {
          DD_API_KEY = var.datadog_api_key
        }
      }
    }
  }
}

data "aws_s3_object" "lambda_statemachine_zip" {
  provider = aws.compute_region
  for_each = toset(local.state_machine_lambdas)
  bucket   = var.lambda_bucket
  key      = "${each.value}.zip"
}

resource "aws_cloudwatch_log_group" "function_statemachine_logs" {
  for_each = toset(local.state_machine_lambdas)
  name     = "/aws/lambda/${each.value}-${terraform.workspace}"
  provider = aws.compute_region

  retention_in_days = 7

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
}

resource "aws_lambda_function" "function" {
  for_each = toset(local.state_machine_lambdas)

  provider      = aws.compute_region
  function_name = "${each.value}-${terraform.workspace}"
  role          = aws_iam_role.lambda.arn
  handler       = "${each.value}-${terraform.workspace}"
  memory_size   = local.lambda_overrides[each.value].memory_size
  ephemeral_storage {
    size = local.lambda_overrides[each.value].ephemeral_storage_size
  }
  architectures = ["arm64"]
  timeout       = local.lambda_overrides[each.value].timeout

  s3_key            = data.aws_s3_object.lambda_statemachine_zip[each.value].key
  s3_object_version = data.aws_s3_object.lambda_statemachine_zip[each.value].version_id
  s3_bucket         = var.lambda_bucket

  environment {
    variables = local.lambda_overrides[each.value].environment.variables
  }

  runtime = "provided.al2"
}


resource "aws_lambda_invocation" "run_migrations" {
  provider      = aws.compute_region
  function_name = aws_lambda_function.function["migrate"].function_name
  input         = jsonencode({ action = "up" })
  triggers = {
    always_run = timestamp()
  }
  depends_on = [
    null_resource.configure_db,
  ]
}

resource "aws_cloudwatch_event_rule" "github_sync" {
  provider = aws.compute_region
  name        = "github-sync-rule-${terraform.workspace}"
  description = "github-sync"
  schedule_expression = "rate(10 minutes)"
  is_enabled = terraform.workspace == "default"
}

# target
resource "aws_cloudwatch_event_target" "github_sync" {
  provider = aws.compute_region
  rule      = aws_cloudwatch_event_rule.github_sync.name
  target_id = "github-sync-${terraform.workspace}"
  arn       = aws_lambda_function.function["github-sync"].arn
}


resource "aws_lambda_permission" "allow_eventbridge" {
  provider = aws.compute_region
  statement_id  = "AllowExecutionFromEventBridge-${terraform.workspace}"
  action        = "lambda:InvokeFunction"
  function_name = "github-sync-${terraform.workspace}"
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.github_sync.arn
}

resource "aws_cloudwatch_event_rule" "stats" {
  provider = aws.compute_region
  name        = "stats-rule-${terraform.workspace}"
  description = "stats"
  schedule_expression = "rate(1 hour)"
  is_enabled = terraform.workspace == "default"
}

# target
resource "aws_cloudwatch_event_target" "stats" {
  provider = aws.compute_region
  rule      = aws_cloudwatch_event_rule.stats.name
  target_id = "stats-${terraform.workspace}"
  arn       = aws_lambda_function.function["stats"].arn
}

resource "aws_lambda_permission" "allow_eventbridge_stats" {
  provider = aws.compute_region
  statement_id  = "AllowExecutionFromEventBridgeStats-${terraform.workspace}"
  action        = "lambda:InvokeFunction"
  function_name = "stats-${terraform.workspace}"
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.stats.arn
}

# state machine role
resource "aws_iam_role" "state_machine_role" {
  name = "state_machine_role-${terraform.workspace}"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "states.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

# state machine policy (batch + lambda), create managed-rule
#   - events:PutTargets
#   - events:PutRule
#   - events:DescribeRule
resource "aws_iam_role_policy" "state_machine_policy" {
  name = "state_machine_policy"
  role = aws_iam_role.state_machine_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": [
        "batch:SubmitJob",
        "batch:TerminateJob",
        "batch:DescribeJobs",
        "batch:DescribeJobDefinitions",
        "batch:DescribeJobQueues",
        "batch:RegisterJobDefinition"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": [
        "events:PutTargets",
        "events:PutRule",
        "events:DescribeRule"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

# statemachine resource
resource "aws_sfn_state_machine" "state_machine" {
  name     = "start-jobs-${terraform.workspace}"
  role_arn = aws_iam_role.state_machine_role.arn
  provider = aws.compute_region

  definition = <<EOF
{
  "Comment": "A description of my state machine",
  "StartAt": "Start coverage",
  "States": {
    "Start coverage": {
      "Type": "Task",
      "Resource": "arn:aws:states:::batch:submitJob.sync",
      "Parameters": {
        "Parameters.$": "$.params",
        "JobDefinition": "${aws_batch_job_definition.coverage_job.arn}",
        "JobName": "coverage",
        "JobQueue": "${aws_batch_job_queue.coverage_queue.arn}"
      },
      "Next": "Handle coverage",
      "ResultPath": "$.coverage_job"
    },
    "Handle coverage": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "Parameters": {
        "FunctionName": "handle-coverage-${terraform.workspace}:$LATEST",
        "Payload.$": "$"
      },
      "Next": "Parallel",
      "ResultPath": "$.coverage_result"
    },
    "Parallel": {
      "Type": "Parallel",
      "End": true,
      "Branches": [
        {
          "StartAt": "Start Sonarcloud",
          "States": {
            "Start Sonarcloud": {
              "Type": "Task",
              "Resource": "arn:aws:states:::batch:submitJob.sync",
              "Parameters": {
                "Parameters.$": "$.params",
                "JobDefinition": "${aws_batch_job_definition.sonar_job.arn}",
                "JobName": "sonar",
                "JobQueue": "${aws_batch_job_queue.sonar_queue.arn}"
              },
              "ResultPath": "$.sonar_job",
              "End": true
            }
          }
        },
        {
          "StartAt": "Start Benchmarks",
          "States": {
            "Start Benchmarks": {
              "Type": "Task",
              "Resource": "arn:aws:states:::batch:submitJob.sync",
              "Parameters": {
                "Parameters.$": "$.params",
                "JobDefinition": "${aws_batch_job_definition.bench_job.arn}",
                "JobName": "benchmarks",
                "JobQueue": "${aws_batch_job_queue.bench_queue.arn}",
                "ArrayProperties": {
                  "Size": ${local.bench_array_size}
                }
              },
              "ResultPath": "$.benchmarks_job",
              "Next": "Handle Benchmarks"
            },
            "Handle Benchmarks": {
              "Type": "Task",
              "Resource": "arn:aws:states:::lambda:invoke",
              "Parameters": {
                "FunctionName": "handle-benchmarks-${terraform.workspace}:$LATEST",
                "Payload.$": "$"
              },
              "End": true
            }
          }
        }
      ]
    }
  }
}
EOF
}

resource "aws_sfn_state_machine" "mutation_state_machine" {
  name     = "start-mutation-jobs-${terraform.workspace}"
  role_arn = aws_iam_role.state_machine_role.arn
  provider = aws.compute_region

  definition = <<EOF
{
  "Comment": "Starts mutation batch job and then handles success with handle mutation lambda",
  "StartAt": "Start mutation",
  "States": {
    "Start mutation": {
      "Type": "Task",
      "Resource": "arn:aws:states:::batch:submitJob.sync",
      "Parameters": {
        "Parameters.$": "$.params",
        "JobDefinition": "${aws_batch_job_definition.mutation_job.arn}",
        "JobName": "mutation",
        "JobQueue": "${aws_batch_job_queue.mutation_queue.arn}"
      },
      "Next": "Handle mutation",
      "ResultPath": "$.mutation_job"
    },
    "Handle mutation": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "Parameters": {
        "FunctionName": "handle-mutation-${terraform.workspace}:$LATEST",
        "Payload.$": "$"
      },
      "End": true,
      "ResultPath": "$.mutation_result"
    }
  }
}
EOF
}
