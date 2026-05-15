package datadogexport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	defaultDatadogSite = "datadoghq.eu"
	defaultChunk       = 24 * time.Hour
)

type Config struct {
	EnvFile         string
	InventoryDir    string
	OutputDir       string
	DatadogSite     string
	From            time.Time
	To              time.Time
	Chunk           time.Duration
	RouteFilter     []string
	DashboardFilter []string
	MaxQueries      int
}

type Result struct {
	OutputDir      string
	DashboardCount int
	QueryCount     int
	RecordCount    int
}

type routeDashboard struct {
	Route       string `json:"route"`
	Title       string `json:"title"`
	ID          string `json:"id"`
	LayoutType  string `json:"layout_type"`
	WidgetCount int    `json:"widget_count"`
}

type dashboard struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Widgets []widget `json:"widgets"`
}

type widget struct {
	Definition widgetDefinition `json:"definition"`
}

type widgetDefinition struct {
	Type     string    `json:"type"`
	Title    string    `json:"title"`
	Widgets  []widget  `json:"widgets"`
	Requests []request `json:"requests"`
}

type request struct {
	Queries  []metricQuery `json:"queries"`
	Q        string        `json:"q"`
	Formulas []formula     `json:"formulas"`
}

type metricQuery struct {
	Name       string `json:"name"`
	DataSource string `json:"data_source"`
	Query      string `json:"query"`
}

type formula struct {
	Formula string `json:"formula"`
}

type queryDescriptor struct {
	Route          string   `json:"route"`
	DashboardID    string   `json:"dashboardId"`
	DashboardTitle string   `json:"dashboardTitle"`
	LayoutType     string   `json:"layoutType"`
	WidgetTitle    string   `json:"widgetTitle,omitempty"`
	WidgetType     string   `json:"widgetType"`
	GroupPath      []string `json:"groupPath,omitempty"`
	RequestIndex   int      `json:"requestIndex"`
	QueryIndex     int      `json:"queryIndex"`
	QueryName      string   `json:"queryName"`
	Query          string   `json:"query"`
	ResolvedQuery  string   `json:"resolvedQuery"`
	DataSource     string   `json:"dataSource"`
	Formulas       []string `json:"formulas,omitempty"`
}

type manifest struct {
	GeneratedAt        string              `json:"generatedAt"`
	Inventory          string              `json:"inventory"`
	Dashboards         []dashboardManifest `json:"dashboards"`
	QueryCount         int                 `json:"queryCount"`
	SelectedQueryCount int                 `json:"selectedQueryCount,omitempty"`
}

type dashboardManifest struct {
	Route          string            `json:"route"`
	DashboardID    string            `json:"dashboardId"`
	DashboardTitle string            `json:"dashboardTitle"`
	LayoutType     string            `json:"layoutType"`
	WidgetCount    int               `json:"widgetCount"`
	RawPath        string            `json:"rawPath"`
	Queries        []queryDescriptor `json:"queries"`
}

type datadogClient struct {
	baseURL    string
	apiKey     string
	appKey     string
	httpClient *http.Client
}

type queryResponse struct {
	Status string        `json:"status"`
	Series []querySeries `json:"series"`
}

type querySeries struct {
	Metric     string       `json:"metric"`
	Scope      string       `json:"scope"`
	Expression string       `json:"expression"`
	TagSet     []string     `json:"tag_set"`
	Interval   int64        `json:"interval"`
	Pointlist  [][]*float64 `json:"pointlist"`
}

type exportRecord struct {
	Route          string           `json:"route"`
	DashboardID    string           `json:"dashboardId"`
	DashboardTitle string           `json:"dashboardTitle"`
	WidgetTitle    string           `json:"widgetTitle,omitempty"`
	WidgetType     string           `json:"widgetType"`
	GroupPath      []string         `json:"groupPath,omitempty"`
	QueryName      string           `json:"queryName"`
	Query          string           `json:"query"`
	ResolvedQuery  string           `json:"resolvedQuery"`
	Formulae       []string         `json:"formulae,omitempty"`
	SeriesMetric   string           `json:"seriesMetric"`
	SeriesScope    string           `json:"seriesScope"`
	SeriesTags     []string         `json:"seriesTags,omitempty"`
	Interval       int64            `json:"intervalSeconds,omitempty"`
	Timestream     timestreamRecord `json:"timestreamRecord"`
}

type timestreamRecord struct {
	Time             string            `json:"time"`
	TimeUnit         string            `json:"timeUnit"`
	MeasureName      string            `json:"measureName"`
	MeasureValue     string            `json:"measureValue"`
	MeasureValueType string            `json:"measureValueType"`
	Dimensions       map[string]string `json:"dimensions"`
}

func Run(ctx context.Context, cfg Config) (Result, error) {
	if err := cfg.validate(); err != nil {
		return Result{}, err
	}

	if err := loadDotEnv(cfg.EnvFile); err != nil {
		return Result{}, err
	}

	client, err := newDatadogClient(cfg.DatadogSite)
	if err != nil {
		return Result{}, err
	}

	routes, err := loadRouteMap(filepath.Join(cfg.InventoryDir, "route-map.json"))
	if err != nil {
		return Result{}, err
	}
	routes = filterRoutes(routes, cfg.RouteFilter, cfg.DashboardFilter)
	if len(routes) == 0 {
		return Result{}, fmt.Errorf("no dashboards matched the provided filters")
	}

	if err := prepareOutputDir(cfg.OutputDir); err != nil {
		return Result{}, err
	}

	man := manifest{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Inventory:   cfg.InventoryDir,
		Dashboards:  make([]dashboardManifest, 0, len(routes)),
	}

	allQueries := make([]queryDescriptor, 0)
	dashboardDir := filepath.Join(cfg.OutputDir, "dashboards")
	if err := os.MkdirAll(dashboardDir, 0o755); err != nil {
		return Result{}, err
	}

	for _, route := range routes {
		rawDashboard, err := client.getDashboard(ctx, route.ID)
		if err != nil {
			return Result{}, fmt.Errorf("fetch dashboard %s: %w", route.ID, err)
		}

		rawPath := filepath.Join(dashboardDir, route.ID+".json")
		if err := os.WriteFile(rawPath, rawDashboard, 0o644); err != nil {
			return Result{}, err
		}

		var parsed dashboard
		if err := json.Unmarshal(rawDashboard, &parsed); err != nil {
			return Result{}, fmt.Errorf("decode dashboard %s: %w", route.ID, err)
		}

		queries := extractQueries(route, parsed)
		man.Dashboards = append(man.Dashboards, dashboardManifest{
			Route:          route.Route,
			DashboardID:    route.ID,
			DashboardTitle: route.Title,
			LayoutType:     route.LayoutType,
			WidgetCount:    route.WidgetCount,
			RawPath:        filepath.ToSlash(filepath.Join("dashboards", route.ID+".json")),
			Queries:        queries,
		})
		allQueries = append(allQueries, queries...)
	}

	man.QueryCount = len(allQueries)
	if cfg.MaxQueries > 0 && cfg.MaxQueries < len(allQueries) {
		allQueries = allQueries[:cfg.MaxQueries]
		man.SelectedQueryCount = len(allQueries)
	}

	if err := writeJSON(filepath.Join(cfg.OutputDir, "manifest.json"), man); err != nil {
		return Result{}, err
	}

	result := Result{
		OutputDir:      cfg.OutputDir,
		DashboardCount: len(man.Dashboards),
		QueryCount:     len(allQueries),
	}

	if cfg.From.IsZero() && cfg.To.IsZero() {
		return result, nil
	}

	recordCount, err := exportHistory(ctx, client, cfg, allQueries)
	if err != nil {
		return Result{}, err
	}
	result.RecordCount = recordCount

	return result, nil
}

func (cfg Config) validate() error {
	if cfg.InventoryDir == "" {
		return fmt.Errorf("inventory directory is required")
	}
	if cfg.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if (cfg.From.IsZero()) != (cfg.To.IsZero()) {
		return fmt.Errorf("both from and to must be set together")
	}
	if !cfg.From.IsZero() && !cfg.To.After(cfg.From) {
		return fmt.Errorf("to must be after from")
	}

	return nil
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open env file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan env file %s: %w", path, err)
	}

	return nil
}

func newDatadogClient(site string) (*datadogClient, error) {
	apiKey := strings.TrimSpace(os.Getenv("DATADOG_API_KEY"))
	appKey := strings.TrimSpace(os.Getenv("DATADOG_APP_KEY"))
	if apiKey == "" || appKey == "" {
		return nil, fmt.Errorf("DATADOG_API_KEY and DATADOG_APP_KEY must be set")
	}

	site = strings.TrimSpace(site)
	if site == "" {
		site = strings.TrimSpace(os.Getenv("DATADOG_SITE"))
	}
	if site == "" {
		site = defaultDatadogSite
	}
	site = strings.TrimPrefix(site, "https://")
	site = strings.TrimPrefix(site, "http://")

	return &datadogClient{
		baseURL: "https://api." + site,
		apiKey:  apiKey,
		appKey:  appKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (c *datadogClient) getDashboard(ctx context.Context, id string) ([]byte, error) {
	endpoint := c.baseURL + "/api/v1/dashboard/" + url.PathEscape(id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (c *datadogClient) queryMetrics(ctx context.Context, from, to time.Time, metricQuery string) (queryResponse, error) {
	params := url.Values{}
	params.Set("from", strconv.FormatInt(from.Unix(), 10))
	params.Set("to", strconv.FormatInt(to.Unix(), 10))
	params.Set("query", metricQuery)

	endpoint := c.baseURL + "/api/v1/query?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return queryResponse{}, err
	}

	body, err := c.do(req)
	if err != nil {
		return queryResponse{}, err
	}

	var response queryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return queryResponse{}, err
	}
	if response.Status != "ok" {
		return queryResponse{}, fmt.Errorf("datadog query returned status %q", response.Status)
	}

	return response, nil
}

func (c *datadogClient) do(req *http.Request) ([]byte, error) {
	req.Header.Set("DD-API-KEY", c.apiKey)
	req.Header.Set("DD-APPLICATION-KEY", c.appKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %s: %s", resp.Status, bytes.TrimSpace(body))
	}

	return body, nil
}

func loadRouteMap(path string) ([]routeDashboard, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var routes []routeDashboard
	if err := json.Unmarshal(body, &routes); err != nil {
		return nil, err
	}

	return routes, nil
}

func filterRoutes(routes []routeDashboard, routeFilter, dashboardFilter []string) []routeDashboard {
	if len(routeFilter) == 0 && len(dashboardFilter) == 0 {
		return routes
	}

	filtered := make([]routeDashboard, 0, len(routes))
	for _, route := range routes {
		if len(routeFilter) > 0 && slices.Contains(routeFilter, route.Route) {
			filtered = append(filtered, route)
			continue
		}
		if len(dashboardFilter) > 0 && slices.Contains(dashboardFilter, route.ID) {
			filtered = append(filtered, route)
		}
	}

	return filtered
}

func prepareOutputDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("output directory %s must be empty", path)
	}

	return nil
}

func writeJSON(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o644)
}

func extractQueries(route routeDashboard, data dashboard) []queryDescriptor {
	queries := make([]queryDescriptor, 0)
	walkWidgets(data.Widgets, nil, route, data, &queries)
	return queries
}

func walkWidgets(widgets []widget, groupPath []string, route routeDashboard, data dashboard, queries *[]queryDescriptor) {
	for _, item := range widgets {
		definition := item.Definition
		nextGroupPath := groupPath
		if definition.Type == "group" {
			title := strings.TrimSpace(definition.Title)
			if title != "" {
				nextGroupPath = append(append([]string{}, groupPath...), title)
			}
			walkWidgets(definition.Widgets, nextGroupPath, route, data, queries)
			continue
		}

		formulasByRequest := make([][]string, len(definition.Requests))
		for requestIndex, req := range definition.Requests {
			formulasByRequest[requestIndex] = collectFormulas(req.Formulas)

			for queryIndex, q := range req.Queries {
				if q.Query == "" {
					continue
				}
				if q.DataSource != "" && q.DataSource != "metrics" {
					continue
				}

				*queries = append(*queries, buildQueryDescriptor(route, data, definition, nextGroupPath, requestIndex, queryIndex, q.Name, q.Query, q.DataSource, formulasByRequest[requestIndex]))
			}

			if req.Q != "" {
				*queries = append(*queries, buildQueryDescriptor(route, data, definition, nextGroupPath, requestIndex, len(req.Queries), "legacy_query", req.Q, "metrics", formulasByRequest[requestIndex]))
			}
		}
	}
}

func buildQueryDescriptor(route routeDashboard, data dashboard, definition widgetDefinition, groupPath []string, requestIndex, queryIndex int, queryName, query, dataSource string, formulas []string) queryDescriptor {
	return queryDescriptor{
		Route:          route.Route,
		DashboardID:    data.ID,
		DashboardTitle: route.Title,
		LayoutType:     route.LayoutType,
		WidgetTitle:    strings.TrimSpace(definition.Title),
		WidgetType:     definition.Type,
		GroupPath:      append([]string{}, groupPath...),
		RequestIndex:   requestIndex + 1,
		QueryIndex:     queryIndex + 1,
		QueryName:      fallbackQueryName(queryName, queryIndex),
		Query:          query,
		ResolvedQuery:  normalizeMetricQuery(query),
		DataSource:     dataSource,
		Formulas:       formulas,
	}
}

func fallbackQueryName(name string, index int) string {
	if strings.TrimSpace(name) != "" {
		return name
	}

	return fmt.Sprintf("query%d", index+1)
}

func collectFormulas(formulas []formula) []string {
	values := make([]string, 0, len(formulas))
	for _, item := range formulas {
		if strings.TrimSpace(item.Formula) == "" {
			continue
		}
		values = append(values, item.Formula)
	}
	return values
}

func normalizeMetricQuery(query string) string {
	return templateFilterPattern.ReplaceAllStringFunc(query, func(filter string) string {
		content := strings.TrimSuffix(strings.TrimPrefix(filter, "{"), "}")
		if !strings.Contains(content, "$") {
			return filter
		}

		parts := strings.Split(content, ",")
		literals := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" || strings.Contains(part, "$") {
				continue
			}
			literals = append(literals, part)
		}

		if len(literals) == 0 {
			return "{*}"
		}

		return "{" + strings.Join(literals, ",") + "}"
	})
}

var templateFilterPattern = regexp.MustCompile(`\{[^}]*\}`)

func exportHistory(ctx context.Context, client *datadogClient, cfg Config, queries []queryDescriptor) (int, error) {
	chunk := cfg.Chunk
	if chunk <= 0 {
		chunk = defaultChunk
	}

	file, err := os.Create(filepath.Join(cfg.OutputDir, "records.jsonl"))
	if err != nil {
		return 0, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	recordCount := 0
	for _, descriptor := range queries {
		for start := cfg.From; start.Before(cfg.To); start = nextChunkStart(start, chunk) {
			end := start.Add(chunk)
			if end.After(cfg.To) {
				end = cfg.To
			}

			response, err := client.queryMetrics(ctx, start, end, descriptor.ResolvedQuery)
			if err != nil {
				return recordCount, fmt.Errorf("query %q from %s to %s: %w", descriptor.ResolvedQuery, start.Format(time.RFC3339), end.Format(time.RFC3339), err)
			}

			for _, series := range response.Series {
				for _, point := range series.Pointlist {
					record, ok := buildExportRecord(descriptor, series, point)
					if !ok {
						continue
					}

					if err := encoder.Encode(record); err != nil {
						return recordCount, err
					}
					recordCount++
				}
			}
		}
	}

	return recordCount, nil
}

func nextChunkStart(start time.Time, chunk time.Duration) time.Time {
	return start.Add(chunk + time.Second)
}

func buildExportRecord(descriptor queryDescriptor, series querySeries, point []*float64) (exportRecord, bool) {
	if len(point) < 2 || point[0] == nil || point[1] == nil {
		return exportRecord{}, false
	}

	timestampMillis := int64(*point[0])
	value := strconv.FormatFloat(*point[1], 'f', -1, 64)
	dimensions := map[string]string{
		"route":           descriptor.Route,
		"dashboard_id":    descriptor.DashboardID,
		"dashboard_title": descriptor.DashboardTitle,
		"widget_type":     descriptor.WidgetType,
		"query_name":      descriptor.QueryName,
		"metric_name":     seriesMetricName(descriptor, series),
	}
	if descriptor.WidgetTitle != "" {
		dimensions["widget_title"] = descriptor.WidgetTitle
	}
	if series.Scope != "" {
		dimensions["series_scope"] = series.Scope
	}
	for index, group := range descriptor.GroupPath {
		if strings.TrimSpace(group) == "" {
			continue
		}
		dimensions[fmt.Sprintf("group_%d", index+1)] = group
	}
	for _, tag := range series.TagSet {
		key, value, ok := strings.Cut(tag, ":")
		if !ok {
			continue
		}
		dimensions["tag_"+sanitizeDimensionName(key)] = value
	}

	return exportRecord{
		Route:          descriptor.Route,
		DashboardID:    descriptor.DashboardID,
		DashboardTitle: descriptor.DashboardTitle,
		WidgetTitle:    descriptor.WidgetTitle,
		WidgetType:     descriptor.WidgetType,
		GroupPath:      descriptor.GroupPath,
		QueryName:      descriptor.QueryName,
		Query:          descriptor.Query,
		ResolvedQuery:  descriptor.ResolvedQuery,
		Formulae:       descriptor.Formulas,
		SeriesMetric:   seriesMetricName(descriptor, series),
		SeriesScope:    series.Scope,
		SeriesTags:     append([]string{}, series.TagSet...),
		Interval:       series.Interval,
		Timestream: timestreamRecord{
			Time:             strconv.FormatInt(timestampMillis, 10),
			TimeUnit:         "MILLISECONDS",
			MeasureName:      "value",
			MeasureValue:     value,
			MeasureValueType: "DOUBLE",
			Dimensions:       dimensions,
		},
	}, true
}

func seriesMetricName(descriptor queryDescriptor, series querySeries) string {
	if strings.TrimSpace(series.Metric) != "" {
		return series.Metric
	}
	return extractMetricName(descriptor.ResolvedQuery)
}

func extractMetricName(query string) string {
	beforeFilter, _, _ := strings.Cut(query, "{")
	_, metric, ok := strings.Cut(beforeFilter, ":")
	if !ok {
		return strings.TrimSpace(query)
	}
	return strings.TrimSpace(metric)
}

func sanitizeDimensionName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "tag"
	}

	var builder strings.Builder
	lastUnderscore := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}

		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}

	sanitized := strings.Trim(builder.String(), "_")
	if sanitized == "" {
		return "tag"
	}

	return sanitized
}
