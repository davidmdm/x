# xhttp

## xhttp.TimeoutHandler

xhttp.TimeoutHandler is an alternative implementation of the http.TimeoutHandler supporting different uses cases.

### Issues with the standard http.TimeoutHandler

The standard http.TimeoutHandler is simple and well suited for small response payloads. It works by buffereing the response into memory. If the handler exits before the timeout expires the buffered response is served. Should the timeout occur first the timeout response is served.

This implies that if you are serving large payloads such as streaming files, the standard timeout Middleware causes you to load the entire file into memory.

The standard TimeoutMiddleware also fails to detect and guard against degraded connections. If the server managed to load the response into memory before the timeout is met, it attempts to serve the content to the connection regardless of the speed of the connection. An attacker could potentially stop reading (hang the connection) and take up server resources.

### The altenative approach

xhttp.TimeoutMiddleware takes a different approach. You can provide two durations as options: initial and rolling. The initial timeout works similarly to the standard timeout handler. If the initial timeout occurs before the first write the timeout response is written to the connection. Where it differs is that if a write occurs before the timeout, the payload is immediately sent over the connection. At that point the rolling timeout comes into effect. If the time elapsed between any two writes is longer than the rolling timeout, the connection is killed.

This approach supports long-lived connections, response streaming, and minimal buffering.

### Recommendations for working with Rolling Timeouts.

It is most people's first instinct to think that connections are fast, and that the initial write will be more slower than subsequent writes. In my own experience I have used configurations with high Initial and low Rolling timeouts such as:

```yaml
Initial: 5s
Rolling: 500ms
```

However that was a mistake. Localhost behaves much differently than real networks. Real networks are unsteady and multi-second pauses between writes are not uncommon. The rolling timeout should be used to kill connections that you can reasonably assume are hung.

TLDR;

```yaml
Initial: The reasonable timeout for your service to respond. If the timeout is exceeded it is unreasonable for your service and want to timeout instead of continue trying.
Rolling: The connection can reasonably be assumed to be hung, or so slow that it is not worth trying to send the entire payload.
```

### Left todo

The following improvements are on the roadmap:

- MaxTimeout -> would create a deadline after which the connection would be killed
- RollingAttempts -> allow multiple times for the rolling timeout to be missed before the connection is killed

I
