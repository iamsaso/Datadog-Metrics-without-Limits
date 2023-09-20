package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

type RateLimitInfo struct {
	Limit     int
	Name      string
	Period    int
	Remaining int
	Reset     int
}

func main() {
	// Read Datadog API and Application Keys from environment variables
	apiKey := os.Getenv("DATADOG_API_KEY")
	appKey := os.Getenv("DATADOG_APPLICATION_KEY")

	if apiKey == "" || appKey == "" {
		fmt.Println("Datadog API and Application keys are not set in environment variables.")
		return
	}

	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: apiKey,
			},
			"appKeyAuth": {
				Key: appKey,
			},
		},
	)

	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewUsageMeteringApi(apiClient)

	optionalParams := datadogV1.NewGetUsageTopAvgMetricsOptionalParameters()

	// Set the start date for the metrics
	optionalParams.WithDay(time.Now().AddDate(0, 0, -1))
	// Number of metrics to pull
	optionalParams.WithLimit(100)

	resp, r, err := api.GetUsageTopAvgMetrics(ctx, *optionalParams)
	handleRateLimit(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsageMeteringApi.GetUsageTopAvgMetrics`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return
	}

	// Loop through the metrics and get the active configuration for each
	for _, metric := range resp.GetUsage() {
		metricID := metric.GetMetricName()

		configuration := datadog.NewConfiguration()
		apiClient := datadog.NewAPIClient(configuration)
		metricApiV2 := datadogV2.NewMetricsApi(apiClient)
		metricApiV1 := datadogV1.NewMetricsApi(apiClient)

		listActiveMetricConfigurationsOptionalParams := datadogV2.NewListActiveMetricConfigurationsOptionalParameters()
		listActiveMetricConfigurationsOptionalParams.WithWindowSeconds(2628000)

		metricResp, r, err := metricApiV2.ListActiveMetricConfigurations(ctx, metricID, *listActiveMetricConfigurationsOptionalParams)
		handleRateLimit(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.ListActiveMetricConfigurations`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			continue
		}

		fmt.Println("-----------------------")
		fmt.Println("Metric ID: ", *metricResp.Data.Id)

		metricMetadataResp, r, err := metricApiV1.GetMetricMetadata(ctx, metricID)
		handleRateLimit(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.ListTagsByMetricName`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		metricType := datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION
		if metricMetadataResp.Type != nil {
			metricType = datadogV2.MetricTagConfigurationMetricTypes(*metricMetadataResp.Type)
			fmt.Println("Metric Type: ", metricType)
		}

		listTagConfigurationByNameResp, r, err := metricApiV2.ListTagConfigurationByName(ctx, metricID)
		handleRateLimit(r)
		if err != nil {
			fmt.Println("List Tag configuration not found. Creating new one.")
			fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.ListTagConfigurationByName`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		fmt.Println("Metric Active Tags: ", metricResp.Data.Attributes.ActiveTags)

		tagConfigurationnResp := datadogV2.MetricTagConfigurationResponse{}

		if listTagConfigurationByNameResp.Data != nil {
			body := datadogV2.MetricTagConfigurationUpdateRequest{
				Data: datadogV2.MetricTagConfigurationUpdateData{
					Type: datadogV2.METRICTAGCONFIGURATIONTYPE_MANAGE_TAGS,
					Id:   metricID,
					Attributes: &datadogV2.MetricTagConfigurationUpdateAttributes{
						Tags: metricResp.Data.Attributes.ActiveTags,
					},
				},
			}
			tagConfigurationnResp, r, err = metricApiV2.UpdateTagConfiguration(ctx, metricID, body)
			handleRateLimit(r)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.UpdateTagConfiguration`: %v\n", err)
				fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			} else {
				fmt.Println("Metric Tags: ", tagConfigurationnResp.Data.Attributes.Tags)
			}
		} else {
			tag := datadogV2.MetricTagConfigurationCreateRequest{
				Data: datadogV2.MetricTagConfigurationCreateData{
					Type: datadogV2.METRICTAGCONFIGURATIONTYPE_MANAGE_TAGS,
					Id:   metricID,
					Attributes: &datadogV2.MetricTagConfigurationCreateAttributes{
						Tags:       metricResp.Data.Attributes.ActiveTags,
						MetricType: metricType,
					},
				},
			}

			tagConfigurationnResp, r, err = metricApiV2.CreateTagConfiguration(ctx, metricID, tag)
			handleRateLimit(r)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.UpdateTagConfiguration`: %v\n", err)
				fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			} else {
				fmt.Println("Metric Tags: ", tagConfigurationnResp.Data.Attributes.Tags)
			}
		}
	}
}

func handleRateLimit(r *http.Response) {
	if r != nil {
		//limit, _ := strconv.Atoi(r.Header.Get("X-Ratelimit-Limit"))
		//name := r.Header.Get("X-Ratelimit-Name")
		//period, _ := strconv.Atoi(r.Header.Get("X-Ratelimit-Period"))
		remaining, _ := strconv.Atoi(r.Header.Get("X-Ratelimit-Remaining"))
		reset, _ := strconv.Atoi(r.Header.Get("X-Ratelimit-Reset"))

		if remaining == 0 {
			fmt.Println("Rate Limit Reached. Sleeping for ", reset, " seconds.")
			time.Sleep(time.Duration(reset) * time.Second)
		}
	}
}
