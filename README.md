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
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
)

	router := gin.Default()
	cleanup, err := pm.Initialize("<POSTMAN-COLLECTION-ID>", "<POSTMAN-API-KEY>")

	if err == nil {
		defer cleanup(context.Background())
		pm.InstrumentGin(router)
	}

```
For full working example see[Gin instrumented example](https://github.com/postmanlabs/postman-go-sdk/tree/master/postmansdk/example/testgo)

## Configuration

**For initialization the SDK, the following values can be configured -**
```golang
import (

	pm "github.com/postmanlabs/postman-go-sdk/postmansdk"
	pminterfaces "github.com/postmanlabs/postman-go-sdk/postmansdk/interfaces"
)

	cleanup, err := pm.Initialize(
    "<POSTMAN-COLLECTION-ID>", 
    "<POSTMAN-API-KEY>",
    pminterfaces.WithDebug(true),
    pminterfaces.With
  )

```


- **CollectionId**: Postman collectionId where requests will be added. This is the id for your live collection.
  - Type: ```string```

- **ApiKey**: Postman api key needed for authentication. 
  - Type: ```string```

- **WithDebug**: Enable/Disable debug logs.
  - Type: ```func(bool)```
  - Default: ```false```

- **WithEnable**: Enable/Disable the SDK. Disabled SDK does not capture any new traces, nor does it use up system resources.
  - Type: ```func(bool)```
  - Default: ```true```

- **WithTruncateData**: Truncate the request and response body so that no PII data is sent to Postman. 
- Disabling it sends actual request and response payloads.
  - Type: ```func(bool)```
  - Default: ```true```
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
            "type": "str"
        },
        "age": {
            "type": "int"
        }
    }
    ```
  - Type: ```boolean```
  - Default: ```True```

- **WithRedactSensitiveData**: Redact sensitive data such as api_keys and auth tokens, before they leave the sdk. This is **enabled** by default. But **NO** rules are set.
  - Example:
    ```
    {
    "redact_sensitive_data": {
        "enable": True(default),
        "rules": {
            "<rule name>": "<regex to match the rule>", # such as -
            "basic_auth": r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,7}\b",
            },
        }
    }
    ```
  - Type: ```object```

- **WithIgnoreIncomingRequests**: List of regexes to be ignored from instrumentation. This rule only applies to endpoints that are **served** by the application/server.
  - Example:
      ```golang
      {
          "ignore_incoming_requests": ["knockknock", "^get.*"]
      }
      ```
      The above example, will ignore any endpoint that contains the word "knockknock" in it, and all endpoints that start with get, and contain any characters after that.
  - Type: ```dict```

- **WithIgnoreOutgoingRequests**: List of regexes to be ignored from instrumentation. This rule only applies to endpoints that are **called** by the application/server.
  - Example:
      ```golang
      {
          "ignore_outgoing_requests": ["knockknock", "^get.*"]
      }
      ```
      The above example, will ignore any endpoint that contains the word "knockknock" in it, and all endpoints that start with get, and contain any characters after that.
  - Type: ```dict```

- **WithBufferIntervalInMilliseconds**: The interval in milliseconds that the SDK waits before sending data to Postman. The default interval is 5000 milliseconds. This interval can be tweaked for lower or higher throughput systems.
  - Type: ```int```
  - Default: ```5000```



