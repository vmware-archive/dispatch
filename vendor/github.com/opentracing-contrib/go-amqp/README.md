# go-amqp

[![Build Status]](https://travis-ci.org/opentracing-contrib/go-amqp)
[![GoDoc]](http://godoc.org/github.com/opentracing-contrib/go-amqp/amqptracer)

[AMQP] instrumentation in Go

For documentation on the packages, [check godoc].

**The APIs in the various packages are experimental and may change in
the future. You should vendor them to avoid spurious breakage.**

## Packages

Instrumentation is provided for the following packages, with the
following caveats:

- **github.com/streadway/amqp**: Client and server instrumentation. *Only supported
  with Go 1.7 and later.*

## Required Reading

In order to understand the [AMQP] instrumentation in Go, one must first
be familiar with the [OpenTracing project] and [terminology] more
specifically.  And also, [OpenTracing API for Go] contains enough examples
to get started with OpenTracing in Go.

## API overview for the AMQP instrumentation

Here are the example serialization and deserialization of the `opentracing`
`SapnContext` over the AMQP broker so that we can visualize the tracing
between the producers and the consumers.

#### Serializing to the wire

```go
    func PublishMessage(
        ctx context.Context,
        ch *amqp.Channel,
        immediate bool,
        msg *amqp.Publishing,
    ) error {
        sp := opentracing.SpanFromContext(ctx)
        defer sp.Finish()

        // Inject the span context into the AMQP header.
        if err := amqptracer.Inject(sp, msg.Headers); err != nil {
            return err
        }

        // Publish the message with the span context.
        return ch.Publish(exchange, key, mandatory, immediate, msg)
    }
```

#### Deserializing from the wire

```go
    func ConsumeMessage(ctx context.Context, msg *amqp.Delivery) error {
        // Extract the span context out of the AMQP header.
        spCtx, _ := amqptracer.Extract(msg.Headers)
        sp := opentracing.StartSpan(
            "ConsumeMessage",
            opentracing.FollowsFrom(spCtx),
        )
        defer sp.Finish()

	// Update the context with the span for the subsequent reference.
        ctx = opentracing.ContextWithSpan(ctx, sp)

        // Actual message processing.
        return ProcessMessage(ctx, msg)
    }
```

[OpenTracing project]: http://opentracing.io
[terminology]: http://opentracing.io/documentation/pages/spec.html
[OpenTracing API for Go]: https://github.com/opentracing/opentracing-go
[AMQP]: https://github.com/streadway/amqp
[Build Status]: https://travis-ci.org/opentracing-contrib/go-amqp.svg
[GoDoc]: https://godoc.org/github.com/opentracing-contrib/go-amqp/amqptracer?status.svg
[check godoc]: https://godoc.org/github.com/opentracing-contrib/go-amqp/amqptracer
