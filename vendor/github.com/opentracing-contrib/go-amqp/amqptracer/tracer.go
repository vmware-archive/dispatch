package amqptracer

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/streadway/amqp"
)

// Inject injects the span context into the AMQP header.
//
// Example:
//
//	func PublishMessage(
//		ctx context.Context,
//		ch *amqp.Channel,
//		exchange, key string,
//		mandatory, immediate bool,
//		msg *amqp.Publishing,
//	) error {
//		sp := opentracing.SpanFromContext(ctx)
//		defer sp.Finish()
//
//		// Inject the span context into the AMQP header.
//		if err := amqptracer.Inject(sp, msg.Headers); err != nil {
//			return err
//		}
//
//		// Publish the message with the span context.
//		return ch.Publish(exchange, key, mandatory, immediate, msg)
//	}
func Inject(span opentracing.Span, hdrs amqp.Table) error {
	c := amqpHeadersCarrier(hdrs)
	return span.Tracer().Inject(span.Context(), opentracing.TextMap, c)
}

// Extract extracts the span context out of the AMQP header.
//
// Example:
//
//	func ConsumeMessage(ctx context.Context, msg *amqp.Delivery) error {
//		// Extract the span context out of the AMQP header.
//		spCtx, _ := amqptracer.Extract(msg.Headers)
//		sp := opentracing.StartSpan(
//			"ConsumeMessage",
//			opentracing.FollowsFrom(spCtx),
//		)
//		defer sp.Finish()
//
//		// Update the context with the span for the subsequent reference.
//		ctx = opentracing.ContextWithSpan(ctx, sp)
//
//		// Actual message processing.
//		return ProcessMessage(ctx, msg)
//	}
func Extract(hdrs amqp.Table) (opentracing.SpanContext, error) {
	c := amqpHeadersCarrier(hdrs)
	return opentracing.GlobalTracer().Extract(opentracing.TextMap, c)
}
