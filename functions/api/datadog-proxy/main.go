package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
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
		// Read the body (handle gzip if needed)
		var reader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer gzReader.Close()
			reader = gzReader
		}

		body, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}

		// Perform URL replacements
		content := string(body)
		content = strings.ReplaceAll(content, "https://static.datadoghq.com", "https://datadog-proxy.corecheck.dev")
		content = strings.ReplaceAll(content, "http://static.datadoghq.com", "https://datadog-proxy.corecheck.dev")
		content = strings.ReplaceAll(content, "//static.datadoghq.com", "//datadog-proxy.corecheck.dev")

		modifiedBody := []byte(content)

		// Re-compress if original was gzipped
		if resp.Header.Get("Content-Encoding") == "gzip" {
			var buf bytes.Buffer
			gzWriter := gzip.NewWriter(&buf)
			if _, err := gzWriter.Write(modifiedBody); err != nil {
				return err
			}
			if err := gzWriter.Close(); err != nil {
				return err
			}
			modifiedBody = buf.Bytes()
		} else {
			// If not gzipped, remove Content-Encoding header
			resp.Header.Del("Content-Encoding")
		}

		// Update Content-Length
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))
		resp.Body = ioutil.NopCloser(bytes.NewReader(modifiedBody))

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
	r.Any("/:route", proxy)
	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
