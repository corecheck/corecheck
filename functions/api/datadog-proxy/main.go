package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func proxy(c *gin.Context) {
	remote, err := url.Parse("https://p.datadoghq.eu")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host

		req.URL.Path = "/" + c.Param("route") + c.Param("proxyPath")
		fmt.Println(req.URL.Path)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body = ioutil.NopCloser(strings.NewReader(strings.ReplaceAll(string(body), "https://static.datadoghq.com", "https://datadog-proxy.corecheck.dev")))

		return nil
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func staticProxy(c *gin.Context) {
	remote, err := url.Parse("https://static.datadoghq.com")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = "/static" + c.Param("proxyPath")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

var ginLambda *ginadapter.GinLambda

func init() {
	r := gin.Default()
	r.RedirectTrailingSlash = false
	r.Use(cors.Default())
	r.Any("/static/*proxyPath", staticProxy)
	r.Any("/:route/*proxyPath", proxy)
	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
