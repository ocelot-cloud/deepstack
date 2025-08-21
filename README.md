# DeepStack

## Overview

DeepStack is a structured logging library for the Ocelot Ecosystem based on the Go SDK library [log/slog](https://go.dev/blog/slog).

### Error Design

* **Log Or Return Principle**: To avoid duplication, either log or return an error, but not both. Therefore, the workflow should be to create DeepStack errors in a low-level function and then immediately return them. Then, they should be passed up to a higher-level function, where they are handled and logged.
* A **stack trace is automatically included in the DeepStack error**. This makes it possible to create a single log in the high-level error-handling function, which displays the full stack trace. This includes the root cause of the error in the low-level function. The `NewError()` function is provided for creating such errors.
* **DeepStack error implements Go `error` interface**, allowing it to be used seamlessly with Go's error handling mechanisms and reducing its coupling with the code in which it is used.
* **DeepStackError Structure**: Other logging libraries often encode context and stack trace information in a single error string, adding encoding complexity. In contrast, DeepStack errors are rich data structures containing extra fields for context and stack traces, thus avoiding this complexity.
* **Adding Error Context**: DeepStack error data structures have a context field that can store key-value pairs. These pairs can be added to extend the context during DeepStack error creation or by intermediate functions passing up the DeepStack error. As these operations are performed directly on the error data structure, the process is much lighter than the costly encoding operations performed by other logging libraries.

```go
type DeepStackError struct {
    message      string
    stackTrace   string
    context      map[string]interface{}
}
```

### Logging Design

* **Structured Logging** is the general use case of the DeepStack library that allows for easy filtering and searching of logs.
* **Error Logging** is a special case in which the DeepStack logger reflects on the error type. If it is a DeepStack error, the library prints all of this information to the console and the log file in a readable manner. This can be extended later to send logs to a server.
* **Centralization** by pushing logs and metrics into a database is planned for the future, allowing for a unified view of all logs across the Ocelot ecosystem.
* **Log persistence** and a **retention policy** for log files are already included.

### Logging Example

```go
const (
    NameField     = "name"
    RoleField     = "role"
)

func main() {
    logger := deepstack.ProviderLogger()
    logger.Info("user logged in", NameField, "john", RoleField, "admin")
    
    err := doAResourceAccessOperation()
    logger.Error("resource access operation failed", deepstack.ErrorField, err)
}

func doAResourceAccessOperation() error {
	return deepstack.NewError("unauthorized access", "user_id", 12345)
}
```

Output:

```text
time=2025-07-21T00:15:00.000+02:00 level=INFO source=main.go:29 msg="user logged in" name=john role=admin
...
time=2025-07-21T00:15:01.000+02:00 level=ERROR source=logger_test.go:29 msg="testing detailed error" user_id=12345
deepstack.subfunction
    /some/path/main.go:33
deepstack.TestLoggingWithStackTrace
    /some/path/main.go:29
```

### Contributing

Please read the [Community](https://ocelot-cloud.org/docs/community/) articles for more information on how to contribute to the project.

### License

This project is licensed under a permissive open source license, the [0BSD License](https://opensource.org/license/0bsd/). See the [LICENSE](LICENSE) file for details.