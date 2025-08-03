TODO


### deepstack

* deepstack: when err.Error() is called, just print the message, not the other information
* deepstack:
    * in console logs, there is at the moment just the file name, but I want also the relative path (duplications make the IDE not recognize the file when clicking on it)
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

* when logging an error which is not a deepstack error, it should be logged -> I think that is already the case, but better re-check
* add deepstack log in all "shared" module errors
* deepstack: make the "signal: killed" message green
* close other todos in code