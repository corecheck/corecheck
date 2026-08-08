package datadogexport

import (
	"reflect"
	"testing"
)

func TestNormalizeMetricQuery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "single template variable becomes wildcard",
			query: "avg:bitcoin.bitcoin.benchmarks.total_time{$Benchmark} by {benchmark_name}",
			want:  "avg:bitcoin.bitcoin.benchmarks.total_time{*} by {benchmark_name}",
		},
		{
			name:  "only dashboard variables become wildcard filter",
			query: "avg:aws.states.execution_time{$scope,$statemachinearn,$lambdafunctionarn} by {statemachinearn}",
			want:  "avg:aws.states.execution_time{*} by {statemachinearn}",
		},
		{
			name:  "literal filters are preserved",
			query: "sum:metric.name{env:prod,$region}",
			want:  "sum:metric.name{env:prod}",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeMetricQuery(testCase.query); got != testCase.want {
				t.Fatalf("normalizeMetricQuery() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestExtractQueriesIncludesNestedGroups(t *testing.T) {
	t.Parallel()

	route := routeDashboard{
		Route:      "/",
		Title:      "Overview",
		ID:         "abc-123",
		LayoutType: "ordered",
	}
	data := dashboard{
		ID:    "abc-123",
		Title: "Overview",
		Widgets: []widget{
			{
				Definition: widgetDefinition{
					Type:  "group",
					Title: "Pull requests",
					Widgets: []widget{
						{
							Definition: widgetDefinition{
								Type:  "query_value",
								Title: "Open pull requests",
								Requests: []request{
									{
										Queries: []metricQuery{
											{Name: "query1", DataSource: "metrics", Query: "avg:bitcoin.bitcoin.pulls.open{*}"},
										},
										Formulas: []formula{{Formula: "query1"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	got := extractQueries(route, data)
	if len(got) != 1 {
		t.Fatalf("expected 1 query, got %d", len(got))
	}

	if got[0].WidgetTitle != "Open pull requests" {
		t.Fatalf("unexpected widget title %q", got[0].WidgetTitle)
	}
	if !reflect.DeepEqual(got[0].GroupPath, []string{"Pull requests"}) {
		t.Fatalf("unexpected group path %#v", got[0].GroupPath)
	}
	if got[0].ResolvedQuery != "avg:bitcoin.bitcoin.pulls.open{*}" {
		t.Fatalf("unexpected resolved query %q", got[0].ResolvedQuery)
	}
}

func TestBuildExportRecordIncludesSeriesTags(t *testing.T) {
	t.Parallel()

	record, ok := buildExportRecord(queryDescriptor{
		Route:          "/tests",
		DashboardID:    "7ck-zbu-au3",
		DashboardTitle: "Bitcoin Core tests",
		WidgetTitle:    "Functional tests duration",
		WidgetType:     "timeseries",
		QueryName:      "query1",
		Query:          "avg:bitcoin.bitcoin.test.functional.duration{*} by {test_name}",
		ResolvedQuery:  "avg:bitcoin.bitcoin.test.functional.duration{*} by {test_name}",
		GroupPath:      []string{"Pull requests"},
	}, querySeries{
		Metric:    "bitcoin.bitcoin.test.functional.duration",
		Scope:     "test_name:wallet_hd.py",
		TagSet:    []string{"test_name:wallet_hd.py"},
		Interval:  300,
		Pointlist: [][]*float64{{floatPtr(1700000000000), floatPtr(42)}},
	}, []*float64{floatPtr(1700000000000), floatPtr(42)})
	if !ok {
		t.Fatal("expected record to be built")
	}

	if got := record.Timestream.Dimensions["tag_test_name"]; got != "wallet_hd.py" {
		t.Fatalf("unexpected tag dimension %q", got)
	}
	if got := record.Timestream.Time; got != "1700000000000" {
		t.Fatalf("unexpected time %q", got)
	}
	if got := record.Timestream.MeasureValue; got != "42" {
		t.Fatalf("unexpected measure value %q", got)
	}
}

func floatPtr(value float64) *float64 {
	return &value
}
