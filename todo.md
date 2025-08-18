TODO

* add DeepStackError function "Unwrap() error" so that "join" function works?

* enable to log entire data structures, maybe reuse the slog.Group feature?

* in console logs, there is at the moment just the file name, but I want also the relative path (duplications make the IDE not recognize the file when clicking on it) -> should work, to be tested

* odd logs sometimes in cloud:

2025-08-02 13:44:37.528 DEBUG security_api.go:41 "Request path"2025-08-02 13:44:37.529 DEBUG security_api.go:46 "request to ocelot-cloud API is called"
url_path=/api/settings/maintenance/read

-> where the bottom structure is white font color, not the expected log level color; might have problems with line breaks in console.

* strange output:
  2025-08-02 11:10:39.209 DEBUG security.go:106 "checking if request is addressed to an app" database_host= request_host=localhost:8080
  2025-08-02 11:10:39.210 DEBUG security.go:110 "is request addressed to an app"2025-08-02 11:10:39.210 DEBUG security.go:110 "is request addressed to an app" is_request_addressed_to_an_app=false
  2025-08-02 11:10:39.210 DEBUG security_api.go:46 "request to ocelot-cloud API is called"
  2025-08-02 11:10:39.210 DEBUG handlers.go:19 "app list handler called"
  is_request_addressed_to_an_app=false

* task-runner: make the "signal: killed" message green

### low prio:

* add the software version to the log so that "source" attribute deterministally references its origin; another interesting field would be an application id
  * realization idea: add the option to define global attributes, which are contained in every log message; they are contained in the log file, but not in console to not clutter it
  * related todo: "add the possibility to create a random ID when application starts so that I can group donated logs by deployed instances -> maybe the feature above could be combined with this; this is one global ID for whole application, while feature above needs a unique ID for each operation usually triggered by a user request"
* feature: a function like AddRoutineContext("unique_operation_id") that takes values from the context.Context and adds it as field to the structured log. Not sure if that should be printed to console? but definitely should be printed to the log file; if the value is empty, I should maybe do a warning log, as this is a hint that I forgot to set the value in the context.Context somewhere in the code (maybe add a flag in the NewDeepStackLogger to dis-/enable this feature?)
