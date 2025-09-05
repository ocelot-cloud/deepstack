# DeepStack

## Overview

DeepStack is a structured logging and error management library for the Ocelot Ecosystem based on the Go SDK library [log/slog](https://go.dev/blog/slog).

### Error Design

Here is the DeepStack error data structure:

```go
type DeepStackError struct {
    message      string
    stackTrace   string
    context      map[string]interface{}
}
```

* **Log Or Return Principle**: To avoid duplication, either log or return an error, but not both. Therefore, errors are created in a low-level function and passed up to a higher-level function, where they are logged once. However, when an error is logged in a high-level function, it must contain information about which function caused it.
* Therefore, a **stack trace is included in the DeepStack error** upon creation.
* **DeepStack error implements Go `error` interface**, allowing it to be used seamlessly with Go's error handling mechanisms and reducing its coupling with the code in which it is used.
* **DeepStackError Structure**: Other logging libraries often encode context and stack trace information in a single error string, adding encoding complexity. By contrast, DeepStack errors are rich data structures that contain additional fields for context and stack traces. This makes them easy to understand and avoids unnecessary complexity.
* **Adding Error Context**: DeepStack error data structures have a context field that can store key-value pairs. These pairs can be added to extend the context during DeepStack error creation or by intermediate functions passing up the DeepStack error. As these operations are performed directly on the error data structure, the process is much lighter than the costly encoding operations performed by other logging libraries.

### Logging Design

* **Structured Logging** is the general use case of the DeepStack library that allows for easy filtering and searching of logs.
* **Error Logging** is a special case in which the DeepStack logger reflects on the error type. If it is a DeepStack error, the library prints all of this information to the console and the log file in a readable manner. This can be extended later to send logs to a server.

## Usage Overview

### Basic Structured Logging

```
func main() {
    logger := deepstack.NewDeepStackLogger("debug")
    // The design of the usage aimed for minimal overhead and simplicity, ideally with one-liners.
    logger.Info("user logged in", "name", "john", "age", 23)
}
```

Output:

```text
2025-08-31 16:58:52.135 INFO main.go:10 "user logged in" age=23 name=john
```

### Error Management

Use a simple **create → propagate → handle** lifecycle with DeepStack. Stack traces are captured **once** at creation. Logging happens **once** at a boundary, e.g. an HTTP handler.

#### 1) Error creation

Create a DeepStack error at the first failure point so the stack trace points to the origin.
- New failure: `deepstack.NewError("resource not found")`
- From a Go error: `deepstack.NewError(err.Error())`

#### 2) Error propagation

Return the error upward. Intermediate layers **do not log**, but they may enrich the error with additional context via `AddContext()` and return it.

#### 3) Error handling

Log the error at the boundary, by passing the error via `deepstack.ErrorField` to the logger. If a non-DeepStack error is logged this way, the logger emits a warning to help you find places where wrapping to DeepStack errors was missed.

#### Context

At any stage you can add context to a DeepStack error as key–value pairs. This context travels with the error and is included when it’s finally logged.

### Full Logging Example

```go
package main

import (
    "os/exec"

    "github.com/ocelot-cloud/deepstack"
)

var logger = deepstack.NewDeepStackLogger("debug")

const (
    // good practice to hard code field names for reusability
    AccessDeniedField = "access_denied"
)

func main() {
    err := func1()
    logger.Error("resource access operation failed", deepstack.ErrorField, err)
}

// intermediate function passing up the error
func func1() error {
    err := func2()
    if err != nil {
        return err
    }
    // do some other stuff here
    return nil
}

// intermediate function adding context and then passing up the error
func func2() error {
    return logger.AddContext(func3(), AccessDeniedField, "access token not found")
}

var someCondition = true

// the function where the error occurs the first time
func func3() error {
    if someCondition {
        // create own error
        return logger.NewError("access token not found", AccessDeniedField, "no access token provided")
    } else {
        // wrap error from external library
        err := exec.Command("not-existing-command").Run()
        return logger.NewError(err.Error())
    }
}
```

Output:

```text
main.func3
    /home/user/GolandProjects/playground/main.go:42
main.func2
    /home/user/GolandProjects/playground/main.go:33
main.func1
    /home/user/GolandProjects/playground/main.go:23
main.main
    /home/user/GolandProjects/playground/main.go:17
runtime.main
    /home/user/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.4.linux-amd64/src/runtime/proc.go:283
runtime.goexit
    /home/user/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.4.linux-amd64/src/runtime/asm_amd64.s:1700

2025-08-31 16:53:38.022 ERROR main.go:18 "resource access operation failed" error_cause="access token not found" access_denied="access token not found"
```

### Asserting DeepStack Errors in Tests

Use the provided assertion helper to verify DeepStack errors in tests. It requires an exact match of the error message and the context.

```go
func TestStuff(t *testing.T) {
    err := someFunction()
    deepstack.AssertDeepStackError(t, err, "some error message", "name", "john", "age", "23")
}
```

### Handlers

A common pattern is for an HTTP handler to receive a request and pass its data to the business logic, which may return an error. The handler is typically responsible for logging the error, mapping it to the relevant HTTP status code and generating an appropriate message for the client.

In standard Go, the business logic returns typed errors and the handlers use type switches to determine the response. However, DeepStack errors use a single error type, which renders this approach incompatible.

Instead, we define a set of error messages that can safely be exposed to clients. If the error matches one of these strings, it is returned unchanged. Otherwise, a default message is used to ensure that no sensitive error messages are returned. For example:

```text

var expectedErrors = deepstack.MapOf("service not configured", "operation not allowed")

func SomeHandler(w http.ResponseWriter, r *http.Request) {
    if err := BusinessLogicServer.DoStuff(); err != nil {
        deepstack.WriteResponseError(w, expectedErrors, err)
        return
    }
}
```

### Register New Log Handlers

The logger created by `NewDeepStackLogger()` comes with a console handler by default. Additional log handlers can be registered there to perform actions such as writing logs to files or sending them to a logging server.

### Contributing

Please read the [Community](https://ocelot-cloud.org/docs/community/) articles for more information on how to contribute to the project.

### License

This project is licensed under a permissive open source license, the [0BSD License](https://opensource.org/license/0bsd/). See the [LICENSE](LICENSE) file for details.