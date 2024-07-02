package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type OtelHandler struct {
	// Next represents the next handler in the chain.
	Next slog.Handler
	// NoBaggage determines whether to add context baggage members to the log record.
	NoBaggage bool
	// NoTraceEvents determines whether to record an event for every log on the active trace.
	NoTraceEvents bool
}

type OtelHandlerOpt func(handler *OtelHandler)

// HandlerFn defines the handler used by slog.Handler as return value.
type HandlerFn func(slog.Handler) slog.Handler

// WithNoBaggage returns an OtelHandlerOpt, which sets the NoBaggage flag
func WithNoBaggage(noBaggage bool) OtelHandlerOpt {
	return func(handler *OtelHandler) {
		handler.NoBaggage = noBaggage
	}
}

// WithNoTraceEvents returns an OtelHandlerOpt, which sets the NoTraceEvents flag
func WithNoTraceEvents(noTraceEvents bool) OtelHandlerOpt {
	return func(handler *OtelHandler) {
		handler.NoTraceEvents = noTraceEvents
	}
}

// New creates a new OtelHandler to use with log/slog
func New(next slog.Handler, opts ...OtelHandlerOpt) *OtelHandler {
	ret := &OtelHandler{
		Next: next,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

// NewOtelHandler creates and returns a new HandlerFn, which wraps a handler with OtelHandler to use with log/slog.
func NewOtelHandler(opts ...OtelHandlerOpt) HandlerFn {
	return func(next slog.Handler) slog.Handler {
		return New(next, opts...)
	}
}

// Handle handles the provided log record and adds correlation between a slog record and an Open-Telemetry span.
func (h OtelHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx == nil {
		return h.Next.Handle(ctx, record)
	}

	if !h.NoBaggage {
		// Adding context baggage members to log record.
		b := baggage.FromContext(ctx)
		for _, m := range b.Members() {
			record.AddAttrs(slog.String(m.Key(), m.Value()))
		}
	}

	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return h.Next.Handle(ctx, record)
	}

	if !h.NoTraceEvents {
		// Adding log info to span event.
		eventAttrs := make([]attribute.KeyValue, 0, record.NumAttrs())
		eventAttrs = append(eventAttrs, attribute.String(slog.MessageKey, record.Message))
		eventAttrs = append(eventAttrs, attribute.String(slog.LevelKey, record.Level.String()))
		eventAttrs = append(eventAttrs, attribute.String(slog.TimeKey, record.Time.Format(time.RFC3339Nano)))
		record.Attrs(func(attr slog.Attr) bool {
			otelAttr := h.slogAttrToOtelAttr(attr)
			if otelAttr.Valid() {
				eventAttrs = append(eventAttrs, otelAttr)
			}

			return true
		})

		span.AddEvent("LogRecord", trace.WithAttributes(eventAttrs...))
	}

	// Adding span info to log record.
	spanContext := span.SpanContext()
	if spanContext.HasTraceID() {
		traceID := spanContext.TraceID().String()
		record.AddAttrs(slog.String("TraceId", traceID))
	}

	if spanContext.HasSpanID() {
		spanID := spanContext.SpanID().String()
		record.AddAttrs(slog.String("SpanId", spanID))
	}

	// Setting span status if the log is an error.
	// Purposely leaving as codes.Unset (default) otherwise.
	if record.Level >= slog.LevelError {
		span.SetStatus(codes.Error, record.Message)
	}

	return h.Next.Handle(ctx, record)
}

// WithAttrs returns a new Otel whose attributes consists of handler's attributes followed by attrs.
func (h OtelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return OtelHandler{
		Next:          h.Next.WithAttrs(attrs),
		NoBaggage:     h.NoBaggage,
		NoTraceEvents: h.NoTraceEvents,
	}
}

// WithGroup returns a new Otel with a group, provided the group's name.
func (h OtelHandler) WithGroup(name string) slog.Handler {
	return OtelHandler{
		Next:          h.Next.WithGroup(name),
		NoBaggage:     h.NoBaggage,
		NoTraceEvents: h.NoTraceEvents,
	}
}

// Enabled reports whether the logger emits log records at the given context and level.
// Note: We handover the decision down to the next handler.
func (h OtelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Next.Enabled(ctx, level)
}

// slogAttrToOtelAttr converts a slog attribute to an OTel one.
// Note: returns an empty attribute if the provided slog attribute is empty.
func (h OtelHandler) slogAttrToOtelAttr(attr slog.Attr, groupKeys ...string) attribute.KeyValue {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return attribute.KeyValue{}
	}

	key := func(k string, prefixes ...string) string {
		for _, prefix := range prefixes {
			k = fmt.Sprintf("%s.%s", prefix, k)
		}

		return k
	}(attr.Key, groupKeys...)

	value := attr.Value.Resolve()

	switch attr.Value.Kind() {
	case slog.KindBool:
		return attribute.Bool(key, value.Bool())
	case slog.KindFloat64:
		return attribute.Float64(key, value.Float64())
	case slog.KindInt64:
		return attribute.Int64(key, value.Int64())
	case slog.KindString:
		return attribute.String(key, value.String())
	case slog.KindTime:
		return attribute.String(key, value.Time().Format(time.RFC3339Nano))
	case slog.KindGroup:
		groupAttrs := value.Group()
		if len(groupAttrs) == 0 {
			return attribute.KeyValue{}
		}

		for _, groupAttr := range groupAttrs {
			return h.slogAttrToOtelAttr(groupAttr, append(groupKeys, key)...)
		}
	case slog.KindAny:
		switch v := attr.Value.Any().(type) {
		case []string:
			return attribute.StringSlice(key, v)
		case []int:
			return attribute.IntSlice(key, v)
		case []int64:
			return attribute.Int64Slice(key, v)
		case []float64:
			return attribute.Float64Slice(key, v)
		case []bool:
			return attribute.BoolSlice(key, v)
		default:
			return attribute.KeyValue{}
		}
	default:
		return attribute.KeyValue{}
	}

	return attribute.KeyValue{}
}
