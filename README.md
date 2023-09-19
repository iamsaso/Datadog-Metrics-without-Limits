# DataDog Metrics Management Tool

This tool is designed to help you manage and configure metrics in your DataDog environment efficiently. It leverages the DataDog API and Go programming language to automate tasks related to metric configuration and management.

## Prerequisites

Before using this tool, make sure you have the following prerequisites in place:

- Datadog API Key: You'll need an API key to authenticate with the DataDog API. Set it as an environment variable named `DATADOG_API_KEY`.

- Datadog Application Key: Similarly, set your application key as an environment variable named `DATADOG_APPLICATION_KEY`.

## Usage

1. Clone this repository to your local machine:

   ```shell
   git clone https://github.com/iamsaso/DataDog.git
   ```

2. Navigate to the project directory:

   ```shell
   cd DataDog
   ```

3. Build and run the application:

   ```shell
   go run main.go
   ```

4. The application will fetch metrics data and perform the following tasks for each metric:

   - Retrieve active metric configurations.
   - Retrieve metric metadata (type).
   - List tag configurations by metric name.
   - Update or create tag configurations as needed.

## Configuration

You can modify the behavior of the tool by adjusting the following optional parameters in the `main.go` file:

- `optionalParams.WithDay`: Configure the time range for metric retrieval.
- `optionalParams.WithLimit`: Set the limit for the number of metrics to process.
- `listActiveMetricConfigurationsOptionalParams.WithWindowSeconds`: Define the time window for active metric configurations.

## Contributing

Contributions are welcome! If you have any improvements or feature suggestions, please feel free to open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Disclaimer:** This tool is provided as-is and may have specific requirements and limitations depending on your Datadog environment. Use it responsibly and ensure it aligns with your organization's policies and practices.
