## About

This SDK instruments web frameworks to capture http requests and auto-generates Postman Live Collections.

## Installation Process

```
go get github.com/postmanlabs/postman-go-sdk
```

## Initializing the SDK

```golang
import (
    pm "github.com/postmanlabs/postman-go-sdk/postmansdk"
)

func main() {
	router := gin.Default()
	cleanup, err := pm.Initialize("<POSTMAN-COLLECTION-ID>", "<POSTMAN-API-KEY>")

	if err == nil {
	    defer cleanup(context.Background())
            // Registers postman SDK middleware
	    pm.InstrumentGin(router)
	}
}

```

For full working example see: [Gin instrumented example](https://github.com/postmanlabs/postman-go-sdk/tree/master/postmansdk/example/testgo)

## Configuration

#### Required Params

- **CollectionId**: Postman collectionId where requests will be added. This is the id for your live collection.

  - Type: `string`

- **ApiKey**: Postman api key needed for authentication.

  - Type: `string`

#### Configuration example

```golang
import (
	pm "github.com/postmanlabs/postman-go-sdk/postmansdk"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
)

cleanup, err := pm.Initialize(
    "<POSTMAN-COLLECTION-ID>",
    "<POSTMAN-API-KEY>",
    pminterfaces.WithDebug(false),
    pminterfaces.WithEnable(true),
    // ...Other configuration options
)

```

#### Configuration Options

- **WithDebug**: Enable/Disable debug logs.

  - Type: `func(bool)`
  - Default: `false`

- **WithEnable**: Enable/Disable the SDK.

  - Disabled SDK does not capture any new traces, nor does it use up system resources.
  - Type: `func(bool)`
  - Default: `true`

- **WithTruncateData**: Truncate the request and response body so that no PII data is sent to Postman.

  - Disabling it sends actual request and response payloads.
  - Type: `func(bool)`
  - Default: `true`
  - Example:

    > Sample payload or non-truncated payload:

    ```JSON
    {
        "first_name": "John",
        "age": 30
    }
    ```

    > Truncated payload:

    ```JSON
    {
        "first_name": {
            "type": "string"
        },
        "age": {
            "type": "float64"
        }
    }
    ```

- **WithRedactSensitiveData**: Redact sensitive data such as api_keys and auth tokens, before they leave the sdk.

  - **enabled** by default.
  - Default regex rules applied are

    ```golang
    "pmPostmanAPIKey":    `PMAK-[a-f0-9]{24}-[a-f0-9]{34}`,
    "pmPostmanAccessKey": `PMAT-[0-9a-z]{26}`,
    "pmBasicAuth":        `Basic [a-zA-Z0-9]{3,1000}(?:[^a-z0-9+({})!@#$%^&|*=]{0,2})`,
    "pmBearerToken":      `Bearer [a-z0-9A-Z-._~+/]{15,1000}`,
    ```

  - Example:
    ```golang
    WithRedactSensitiveData(
        true,
        map[string]string{
            "<rule name>": "<regex to match the rule>",
            "key": `PMAT-[0-9a-z]{26}`,
        }
    )
    ```
  - Type: `func(bool, map[string][string])`

- **WithIgnoreIncomingRequests**: List of regexes to be ignored from instrumentation.

  - This rule only applies to endpoints that are **served** by the application/server.

  - Example:
    ```golang
        WithIgnoreIncomingRequests(
          []string{"knockknock", "^get.*"}
        )
    ```
    Ignore any incoming request endpoints matching the two regexes.
  - Type: `func([]string)`

- **WithBufferIntervalInMilliseconds**: Interval between SDK data push to backend
  - Type: `func(int)`
  - Default: `5000` milliseconds
