## XIO

Package `xio` is a context-aware port of common io operations. This allows long running operations to be canceled.

### API

```go
xio.Copy(context.Context, io.Writer, io.Reader)

xio.CopyBuffer(context.Context, io.Writer, io.Reader)

xio.CopyN(context.Context, io.Writer, io.Reader, int64)

xio.ReadAll(context.Context, io.Reader)
```

The copy functions accept `xio.CopyOption` variadic function arguments. They are:

- `func Buffer(b []byte) CopyOption` -> Allows us to specify the buffer used for copying data
- `func BufferSize(size int) CopyOption` -> Allows us to change the size of the internal buffer used for copying (default 32Kb same as standard `io`). Not used if a Buffer is specified.
- `WaitForLastOp(value bool) CopyOption` -> Fundamentally read and write operations are synchronous, and when the context is canceled `xio` waits for any ongoing write/read to finish before returning. This allows `xio` to return the correct amount of bytes copied. When false, Copy returns immediately, but the bytes copied total may be inaccurate. Default `true`.

## Example

The following program sets up a cancelable context by SIGINT, and starts a copy operation. If a SIGINT occurs before the copy is finished, the copy operation will exit.

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
    defer stop()

    n, err := xio.Copy(ctx, dst, src)

    fmt.Println("copied %d bytes with error: %v", n, err)
}
```
