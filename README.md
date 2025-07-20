# DeepStack

## Overview

DeepStack is a structured logging library for the Ocelot Ecosystem based on the Go SDK library [log/slog](https://go.dev/blog/slog).

### Error Design

* **Log Or Return Principle**: To avoid duplication, either log or return an error, but not both. Therefore, the workflow should be to create DeepStack errors in a low-level function and then immediately return them. Then, they should be passed up to a higher-level function, where they are handled and logged.
* A **stack trace is automatically included in the DeepStack error**. This makes it possible to create a single log in the high-level error-handling function, which displays the full stack trace. This includes the root cause of the error in the low-level function. The `NewError()` function is provided for creating errors.
* **DeepStack error implements Go `error` interface**, allowing it to be used seamlessly with Go's error handling mechanisms and reducing its coupling with the code in which it is used.
* **DeepStackError Structure**: Other logging libraries often encode context and stack trace information in a single string to create errors. DeepStack errors, on the other hand, are more complex data structures that carry additional fields to handle the extra complexity.
* **Adding Error Context**: DeepStack errors have an additional context field that can store key-value pairs. These pairs can be added to extend the context during DeepStack error creation or by intermediate functions passing up the DeepStack error. Since these operations are performed on the error object itself, the process is lightweight and avoids string concatenation or complex error wrapping.

```go
type DeepStackError struct {
    message      string
    stackTrace   string
    context      map[string]interface{}
}
```

### Logging Design

* **Structured Logging** is the general use case of this library that allows for easy filtering and searching of logs.
* **Error Logging** is a special case in which the DeepStack logger reflects on the error type. If it is a DeepStack error, the library prints all of this information to the console and the log file in a readable manner. This can be extended later to send logs to a server.
* **Centralization** by pushing logs and metrics into a database is planned for the future, allowing for a unified view of all logs across the Ocelot ecosystem.
* A **Retention Policy** for log files is automatically included.

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

This project is licensed under the 0BSD License - see the [LICENSE](LICENSE) file for details.