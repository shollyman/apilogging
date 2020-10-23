# apilogging

This is a simplistic logging utility that can help users working with a Google API client library that uses an underlying client based on google.golang.org/api/

Within cloud.google.com/go/ there are several submodules that also work with this, as they
are based on HTTP REST clients rather than GRPC transports:

* cloud.google.com/go/bigquery
* cloud.google.com/go/storage

See integration_test.go for examples of usage.