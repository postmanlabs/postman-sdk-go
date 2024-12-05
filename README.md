> [!WARNING]
> **This package is DEPRECATED! It is no longer maintained**

![Postman](https://user-images.githubusercontent.com/117167853/230871188-0b05ff7c-8b61-401b-9d9a-4c1cb79ade88.jpg)

## About

This SDK instruments web frameworks to capture http requests and auto-generates Postman Live Collections.

## Installation Process

```
go get github.com/postmanlabs/postman-sdk-go
```

## Initializing the SDK

```golang
import (
    pm "github.com/postmanlabs/postman-sdk-go/postmansdk"
)

func main() {
	router := gin.Default()
	sdk, err := pm.Initialize("<POSTMAN-COLLECTION-ID>", "<POSTMAN-API-KEY>")

	if err == nil {
      // Registers a custom middleware
	    sdk.Integrations.Gin(router)
	}
}

```

For full working example see: [Gin instrumented example](https://github.com/postmanlabs/postman-sdk-go/blob/bbba6b5060e098fb25d601077769e1084729f5fe/postmansdk/example/testgo/main.go#L16)

## Configuration

#### Required Params

- **CollectionId**: Postman collectionId where requests will be added. This is the id for your live collection.

  - Type: `string`

- **ApiKey**: Postman api key needed for authentication.

  - Type: `string`

#### Configuration example

```golang
import (
	pm "github.com/postmanlabs/postman-sdk-go/postmansdk"
)

sdk, err := pm.Initialize(
    "<POSTMAN-COLLECTION-ID>",
    "<POSTMAN-API-KEY>",
    pm.WithDebug(false),
    pm.WithEnable(true),
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
  When this is enabled, below redaction rules are applied by default (they are not case-sensitive):
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

    // To enable default redactions
    WithRedactSensitiveData(
        true,
        map[string]string{}
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
