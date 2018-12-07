# proxy
--
    import "github.com/mwitkow/grpc-proxy/proxy"

Package proxy provides a reverse proxy handler for gRPC.

The implementation allows a `grpc.Server` to pass a received ServerStream to a
ClientStream without understanding the semantics of the messages exchanged. It
basically provides a transparent reverse-proxy.

This package is intentionally generic, exposing a `StreamDirector` function that
allows users of this package to implement whatever logic of backend-picking,
dialing and service verification to perform.

See examples on documented functions.

## Usage

#### func  Codec

```go
func Codec() grpc.Codec
```
Codec returns a proxying grpc.Codec with the default protobuf codec as parent.

See CodecWithParent.

#### func  CodecWithParent

```go
func CodecWithParent(fallback grpc.Codec) grpc.Codec
```
CodecWithParent returns a proxying grpc.Codec with a user provided codec as
parent.

This codec is *crucial* to the functioning of the proxy. It allows the proxy
server to be oblivious to the schema of the forwarded messages. It basically
treats a gRPC message frame as raw bytes. However, if the server handler, or the
client caller are not proxy-internal functions it will fall back to trying to
decode the message using a fallback codec.

#### func  RegisterService

```go
func RegisterService(server *grpc.Server, director StreamDirector, serviceName string, methodNames ...string)
```
RegisterService sets up a proxy handler for a particular gRPC service and
method. The behaviour is the same as if you were registering a handler method,
e.g. from a codegenerated pb.go file.

This can *only* be used if the `server` also uses grpcproxy.CodecForServer()
ServerOption.

#### func  TransparentHandler

```go
func TransparentHandler(director StreamDirector) grpc.StreamHandler
```
TransparentHandler returns a handler that attempts to proxy all requests that
are not registered in the server. The indented use here is as a transparent
proxy, where the server doesn't know about the services implemented by the
backends. It should be used as a `grpc.UnknownServiceHandler`.

This can *only* be used if the `server` also uses grpcproxy.CodecForServer()
ServerOption.

#### type StreamDirector

```go
type StreamDirector func(ctx context.Context, fullMethodName string) (*grpc.ClientConn, error)
```

StreamDirector returns a gRPC ClientConn to be used to forward the call to.

The presence of the `Context` allows for rich filtering, e.g. based on Metadata
(headers). If no handling is meant to be done, a `codes.NotImplemented` gRPC
error should be returned.

It is worth noting that the StreamDirector will be fired *after* all server-side
stream interceptors are invoked. So decisions around authorization, monitoring
etc. are better to be handled there.

See the rather rich example.
