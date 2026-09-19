#!/usr/bin/env bash
#
# Local development startup: Postgres, LocalStack (Lambdas + API Gateway), and frontend.
# Requires: Docker, Go 1.21+, Node/npm, AWS CLI. Exits with clear messages on failure.
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AWS_ARGS="--endpoint-url=http://localhost:4566 --region us-east-1"
LOCALSTACK_URL="http://localhost:4566"
API_LAMBDAS=(get-pull list-pulls get-report get-mutation)
S3_BUCKET="corecheck-api-lambdas-local"
STAGE_NAME="api"

# --- Helpers ---
die() {
  echo "Error: $1" >&2
  echo "Fix the issue above and run this script again." >&2
  exit 1
}

check_port() {
  local port=$1
  if command -v ss &>/dev/null; then
    ss -tln | awk '{print $4}' | grep -q ":$port$" && return 0
  elif command -v netstat &>/dev/null; then
    netstat -tln | awk '{print $4}' | grep -q ":$port$" && return 0
  else
    return 1
  fi
}

# --- 1. Prerequisites ---
echo "Checking prerequisites..."

if ! docker info &>/dev/null; then
  die "Docker is not running. Start Docker (or the Docker daemon) and try again."
fi

if ! command -v go &>/dev/null; then
  die "Go is not installed or not in PATH. Install Go 1.21+ and try again."
fi
if ! go version | grep -q go1\.; then
  die "Go version could not be determined. Install Go 1.21+ and try again."
fi

if ! command -v node &>/dev/null; then
  die "Node.js is not installed. Install Node.js and try again."
fi
if ! command -v npm &>/dev/null; then
  die "npm is not installed. Install Node.js (includes npm) and try again."
fi

if ! command -v aws &>/dev/null; then
  die "AWS CLI is not installed. Install it (e.g. pip install awscli, or see https://aws.amazon.com/cli/) and try again."
fi

if check_port 5432; then
  if docker ps --filter "name=corecheck-postgres-local" --format '{{.Names}}' 2>/dev/null | grep -q corecheck-postgres-local; then
    echo "Port 5432 is in use by existing corecheck Postgres container; script will use it."
  else
    die "Port 5432 is in use by something other than corecheck Postgres. Stop that process (or use a different port in this script) and try again."
  fi
fi
if check_port 4566; then
  if docker ps --filter "name=corecheck-localstack" --format '{{.Names}}' 2>/dev/null | grep -q corecheck-localstack; then
    echo "Port 4566 is in use by existing corecheck LocalStack container; script will stop and replace it."
  else
    die "Port 4566 is in use by something other than corecheck LocalStack. Stop that process or use a different port and try again."
  fi
fi

echo "Prerequisites OK."
echo ""

# --- 2. Postgres ---
echo "Starting Postgres..."
PG_CONTAINER="corecheck-postgres-local"
if docker ps -a --format '{{.Names}}' | grep -q "^${PG_CONTAINER}$"; then
  if ! docker ps --format '{{.Names}}' | grep -q "^${PG_CONTAINER}$"; then
    docker start "$PG_CONTAINER" || die "Failed to start existing Postgres container."
  fi
else
  docker run -d \
    --name "$PG_CONTAINER" \
    -e POSTGRES_USER=corecheck \
    -e POSTGRES_PASSWORD=corecheck \
    -e POSTGRES_DB=corecheck \
    -p 5432:5432 \
    postgres:15-alpine \
    || die "Failed to create Postgres container."
fi

echo -n "Waiting for Postgres to be ready"
for i in {1..30}; do
  if docker exec "$PG_CONTAINER" pg_isready -U corecheck -q 2>/dev/null; then
    echo " OK."
    break
  fi
  echo -n "."
  sleep 1
  if [[ $i -eq 30 ]]; then
    die "Postgres did not become ready in time. Check: docker logs $PG_CONTAINER"
  fi
done
echo ""

# --- 3. LocalStack ---
echo "Starting LocalStack..."
LS_CONTAINER="corecheck-localstack"
if docker ps -a --format '{{.Names}}' | grep -q "^${LS_CONTAINER}$"; then
  echo "Stopping and removing existing LocalStack container..."
  docker stop "$LS_CONTAINER" 2>/dev/null || true
  docker rm "$LS_CONTAINER" 2>/dev/null || true
fi
# Lambda containers need to reach Postgres on the host; pass host.docker.internal into Lambda containers.
# On Linux, also add host gateway to the LocalStack container itself.
EXTRA_LS_ARGS=()
if [[ "$(uname)" == "Linux" ]]; then
  EXTRA_LS_ARGS=(--add-host=host.docker.internal:host-gateway)
fi

docker run -d \
  --name "$LS_CONTAINER" \
  -e SERVICES=lambda,s3,apigateway,iam \
  -e LAMBDA_RUNTIME_ENVIRONMENT_TIMEOUT=60 \
  -e LAMBDA_DOCKER_FLAGS="--add-host=host.docker.internal:host-gateway" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -p 4566:4566 \
  "${EXTRA_LS_ARGS[@]}" \
  localstack/localstack:latest \
  || die "Failed to create LocalStack container."

echo -n "Waiting for LocalStack to be ready"
for i in {1..60}; do
  if curl -sSf "$LOCALSTACK_URL/_localstack/health" 2>/dev/null | grep -q '"lambda": "available"'; then
    echo " OK."
    break
  fi
  echo -n "."
  sleep 1
  if [[ $i -eq 60 ]]; then
    die "LocalStack did not become ready in time. Check: docker logs $LS_CONTAINER"
  fi
done
echo ""

# --- 4. Build Lambdas ---
echo "Building Lambda binaries..."
BUILD_DIR="${REPO_ROOT}/.build/local-lambdas"
mkdir -p "$BUILD_DIR"
for name in "${API_LAMBDAS[@]}"; do
  (cd "$REPO_ROOT" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=readonly -ldflags='-s -w' -o "$BUILD_DIR/bootstrap" "./functions/api/$name") \
    || die "Failed to build Lambda $name. Check Go and repo state."
  (cd "$BUILD_DIR" && zip -j -q "$BUILD_DIR/$name.zip" bootstrap) \
    || die "Failed to zip Lambda $name."
done
echo "Lambdas built."
echo ""

# --- 5. S3 + IAM + Lambda ---
# LocalStack accepts dummy credentials; set them so AWS CLI does not use real AWS creds
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-test}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-test}"

echo "Creating S3 bucket and uploading zips..."
aws $AWS_ARGS s3 mb "s3://$S3_BUCKET" 2>/dev/null || true
for name in "${API_LAMBDAS[@]}"; do
  aws $AWS_ARGS s3 cp "$BUILD_DIR/$name.zip" "s3://$S3_BUCKET/$name.zip" \
    || die "Failed to upload $name.zip to S3."
done

echo "Creating IAM role for Lambda..."
ROLE_NAME="corecheck-lambda-role-local"
ROLE_ARN=""
if aws $AWS_ARGS iam get-role --role-name "$ROLE_NAME" &>/dev/null; then
  ROLE_ARN=$(aws $AWS_ARGS iam get-role --role-name "$ROLE_NAME" --query 'Role.Arn' --output text)
else
  aws $AWS_ARGS iam create-role \
    --role-name "$ROLE_NAME" \
    --assume-role-policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}' \
    --output text >/dev/null \
    || die "Failed to create IAM role."
  ROLE_ARN=$(aws $AWS_ARGS iam get-role --role-name "$ROLE_NAME" --query 'Role.Arn' --output text)
  # LocalStack: inline policy for CloudWatch Logs (no managed policies required)
  aws $AWS_ARGS iam put-role-policy \
    --role-name "$ROLE_NAME" \
    --policy-name "logs" \
    --policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["logs:CreateLogStream","logs:PutLogEvents"],"Resource":"*"}]}' \
    || die "Failed to attach logs policy to IAM role."
fi
echo "IAM role: $ROLE_ARN"
echo ""

echo "Creating Lambda functions..."
LAMBDA_ENV='{"Variables":{"AUTO_MIGRATE":"true","DATABASE_HOST":"host.docker.internal","DATABASE_PORT":"5432","DATABASE_USER":"corecheck","DATABASE_PASSWORD":"corecheck","DATABASE_NAME":"corecheck","BUCKET_DATA_URL":"http://host.docker.internal:4566"}}'
for name in "${API_LAMBDAS[@]}"; do
  if aws $AWS_ARGS lambda get-function --function-name "$name" &>/dev/null; then
    aws $AWS_ARGS lambda update-function-code --function-name "$name" --s3-bucket "$S3_BUCKET" --s3-key "$name.zip" --output text >/dev/null \
      || die "Failed to update Lambda $name."
    aws $AWS_ARGS lambda update-function-configuration --function-name "$name" --environment "$LAMBDA_ENV" --output text >/dev/null \
      || true
  else
    aws $AWS_ARGS lambda create-function \
      --function-name "$name" \
      --runtime provided.al2 \
      --handler bootstrap \
      --role "$ROLE_ARN" \
      --code "S3Bucket=$S3_BUCKET,S3Key=$name.zip" \
      --timeout 30 \
      --memory-size 128 \
      --environment "$LAMBDA_ENV" \
      --output text >/dev/null \
      || die "Failed to create Lambda $name. Check LocalStack logs."
  fi
done
echo "Lambdas ready."
echo ""

# --- 6. API Gateway ---
echo "Creating API Gateway..."
API_ID=""
if API_ID=$(aws $AWS_ARGS apigateway get-rest-apis --query "items[?name=='corecheck-api-local'].id" --output text 2>/dev/null) && [[ -n "$API_ID" ]]; then
  echo "Using existing API: $API_ID"
else
  API_ID=$(aws $AWS_ARGS apigateway create-rest-api --name "corecheck-api-local" --query 'id' --output text) \
    || die "Failed to create REST API."
fi

ROOT_ID=$(aws $AWS_ARGS apigateway get-resources --rest-api-id "$API_ID" --query "items[?path=='/'].id" --output text) \
  || die "Failed to get root resource."

# Resource IDs (create or get)
create_or_get_resource() {
  local parent_id=$1
  local path_part=$2
  local existing
  existing=$(aws $AWS_ARGS apigateway get-resources --rest-api-id "$API_ID" --query "items[?pathPart=='$path_part' && parentId=='$parent_id'].id" --output text 2>/dev/null)
  if [[ -n "$existing" ]]; then
    echo "$existing"
    return
  fi
  aws $AWS_ARGS apigateway create-resource --rest-api-id "$API_ID" --parent-id "$parent_id" --path-part "$path_part" --query 'id' --output text
}

# /pulls
RES_PULLS=$(create_or_get_resource "$ROOT_ID" "pulls") || die "Failed to create /pulls"
# /pulls/{id}
RES_PULLS_ID=$(create_or_get_resource "$RES_PULLS" "{id}") || die "Failed to create /pulls/{id}"
# /pulls/{id}/report
RES_REPORT=$(create_or_get_resource "$RES_PULLS_ID" "report") || die "Failed to create /pulls/{id}/report"
# /mutations
RES_MUTATIONS=$(create_or_get_resource "$ROOT_ID" "mutations") || die "Failed to create /mutations"
# /mutations/meta
RES_MUTATIONS_META=$(create_or_get_resource "$RES_MUTATIONS" "meta") || die "Failed to create /mutations/meta"
# /me
RES_ME=$(create_or_get_resource "$ROOT_ID" "me") || die "Failed to create /me"

# Methods and integrations
put_method_and_integration() {
  local res_id=$1
  local method=$2
  local lambda_name=$3
  aws $AWS_ARGS apigateway put-method --rest-api-id "$API_ID" --resource-id "$res_id" --http-method "$method" --authorization-type NONE --output text >/dev/null 2>&1 || true
  local lambda_arn
  lambda_arn=$(aws $AWS_ARGS lambda get-function --function-name "$lambda_name" --query 'Configuration.FunctionArn' --output text)
  aws $AWS_ARGS apigateway put-integration \
    --rest-api-id "$API_ID" \
    --resource-id "$res_id" \
    --http-method "$method" \
    --type AWS_PROXY \
    --integration-http-method POST \
    --uri "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/$lambda_arn/invocations" \
    --output text >/dev/null \
    || die "Failed to put integration for $res_id $method"
  aws $AWS_ARGS lambda add-permission \
    --function-name "$lambda_name" \
    --statement-id "apigw-$res_id-$method" \
    --action lambda:InvokeFunction \
    --principal apigateway.amazonaws.com \
    --source-arn "arn:aws:execute-api:us-east-1:000000000000:$API_ID/*/$method/*" \
    --output text >/dev/null 2>&1 || true
}

put_method_and_integration "$RES_PULLS" "GET" "list-pulls"
put_method_and_integration "$RES_PULLS_ID" "GET" "get-pull"
put_method_and_integration "$RES_REPORT" "GET" "get-report"
put_method_and_integration "$RES_MUTATIONS" "GET" "get-mutation"
put_method_and_integration "$RES_MUTATIONS_META" "GET" "get-mutation"

# /me MOCK
aws $AWS_ARGS apigateway put-method --rest-api-id "$API_ID" --resource-id "$RES_ME" --http-method GET --authorization-type NONE --output text >/dev/null 2>&1 || true
aws $AWS_ARGS apigateway put-integration \
  --rest-api-id "$API_ID" \
  --resource-id "$RES_ME" \
  --http-method GET \
  --type MOCK \
  --request-templates '{"application/json":"{\"statusCode\": 200}"}' \
  --output text >/dev/null \
  || die "Failed to put MOCK integration for /me"
aws $AWS_ARGS apigateway put-method-response --rest-api-id "$API_ID" --resource-id "$RES_ME" --http-method GET --status-code 200 --output text >/dev/null 2>&1 || true
aws $AWS_ARGS apigateway put-integration-response \
  --rest-api-id "$API_ID" --resource-id "$RES_ME" --http-method GET --status-code 200 \
  --selection-pattern "" --response-templates '{"application/json":"{}"}' \
  --output text >/dev/null 2>&1 || true

# Deploy (create deployment and update stage)
aws $AWS_ARGS apigateway create-deployment --rest-api-id "$API_ID" --stage-name "$STAGE_NAME" --output text >/dev/null \
  || die "Failed to create API Gateway deployment."

# Use non-deprecated LocalStack invoke path (/_aws/execute-api/... instead of /restapis/.../_user_request_)
API_BASE_URL="${LOCALSTACK_URL}/_aws/execute-api/${API_ID}/${STAGE_NAME}"
echo "API Gateway ready. Base URL: $API_BASE_URL"
echo ""

# --- 6b. Optional: sync PRs from GitHub into the DB ---
if [[ -n "${GITHUB_ACCESS_TOKEN:-}" ]]; then
  echo "Syncing open PRs from GitHub (bitcoin/bitcoin)..."
  if (cd "$REPO_ROOT" && DATABASE_HOST=localhost DATABASE_PORT=5432 DATABASE_USER=corecheck DATABASE_PASSWORD=corecheck DATABASE_NAME=corecheck GITHUB_ACCESS_TOKEN="$GITHUB_ACCESS_TOKEN" go run ./cmd/sync-prs); then
    echo "PR sync done."
  else
    echo "PR sync failed (check GITHUB_ACCESS_TOKEN and network). Frontend will still start; DB may have no PRs."
  fi
else
  echo "Tip: set GITHUB_ACCESS_TOKEN and re-run to sync PRs from GitHub (e.g. export GITHUB_ACCESS_TOKEN=ghp_...)."
fi
echo ""

# --- 7. Frontend ---
echo "Starting frontend (pointed at local API)..."
cd "$REPO_ROOT/frontend"
if [[ ! -d node_modules ]]; then
  echo "Running npm install..."
  npm install || die "npm install failed. Check Node and network."
fi
echo "Run the frontend with: PUBLIC_ENDPOINT=$API_BASE_URL npm run dev"
echo ""
trap 'echo "To stop: docker stop $PG_CONTAINER $LS_CONTAINER"' EXIT
exec env PUBLIC_ENDPOINT="$API_BASE_URL" npm run dev
