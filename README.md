# apilogging

This is a simple logging utility that can help users working with a Google API client library
based on google.golang.org/api/ to log HTTP-level interactions.

Within cloud.google.com/go/ there are several submodules that also work with this, as they
are based on underlying discovery-based HTTP clients rather than gRPC.  Notable libraries include:

* cloud.google.com/go/bigquery
* cloud.google.com/go/storage

See integration_test.go for examples of usage.