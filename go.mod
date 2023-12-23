module github.com/corecheck/corecheck

go 1.21.1

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go v1.48.13
	github.com/awslabs/aws-lambda-go-api-proxy v0.16.0
	github.com/davecgh/go-spew v1.1.1
	github.com/google/go-github/v57 v57.0.0
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/labstack/echo/v4 v4.11.3
	github.com/waigani/diffparser v0.0.0-20190828052634-7391f219313d
	go.uber.org/zap v1.26.0
	golang.org/x/oauth2 v0.11.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/waigani/diffparser => github.com/aureleoules/diffparser v0.0.0-20231221115932-4781aa7ec7a0
