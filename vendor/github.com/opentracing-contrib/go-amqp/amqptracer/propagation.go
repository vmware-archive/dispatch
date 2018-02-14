package amqptracer

// amqpHeadersCarrier satisfies both TextMapWriter and TextMapReader.
//
// Example usage for server side:
//
//     carrier := amqpHeadersCarrier(amqp.Table)
//     clientContext, err := tracer.Extract(
//         opentracing.TextMap,
//         carrier)
//
// Example usage for client side:
//
//     carrier := amqpHeadersCarrier(amqp.Table)
//     err := tracer.Inject(
//         span.Context(),
//         opentracing.TextMap,
//         carrier)
//
type amqpHeadersCarrier map[string]interface{}

// ForeachKey conforms to the TextMapReader interface.
func (c amqpHeadersCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, val := range c {
		v, ok := val.(string)
		if !ok {
			continue
		}
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Set implements Set() of opentracing.TextMapWriter.
func (c amqpHeadersCarrier) Set(key, val string) {
	c[key] = val
}
