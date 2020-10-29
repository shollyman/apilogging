# apilogging


## Overview

This is a simple logging utility that can help users working with a Google API client library
based on google.golang.org/api/ to log HTTP-level interactions.

Within cloud.google.com/go/ there are several submodules that also work with this, as they
are based on underlying discovery-based HTTP clients rather than gRPC.  Notable libraries include:

* cloud.google.com/go/bigquery
* cloud.google.com/go/storage

See integration_test.go for examples of usage.


# Background

This logger relies on wrapping one `http.RoundTripper` with an instrumented roundtripper.  Thus,
any changes to the http request made by the inner RoundTripper are not observable.

The Google API discovery client provides access to it's custom roundtripper via 
`google.golang.org/api/transport/http`.  It allows the same pattern, where it accepts a wrapped
base roundtripper.  It modifies requests to add essential pieces to the request such as the
authentication token, as well as secondary values that are communicated through additional
headers.

When used, request flow ends up passing through the api client transport, then the intercepting
logging transport, and then finally a vanilla http.DefaultTransport.

# Disclaimer

This is not an officially supported Google product.
