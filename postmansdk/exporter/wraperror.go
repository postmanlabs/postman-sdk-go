package exporter

// This file is copied from below package, as you can't import packages with
// internal in the path
// "go.opentelemetry.io/otel/exporters/otlp/internal/wrappederror.go"
type ErrorKind int

const (
	// TracesExport indicates the error comes from the OTLP trace exporter.
	TracesExport ErrorKind = iota
)

// prefix returns a prefix for the Error() string.
func (k ErrorKind) prefix() string {
	switch k {
	case TracesExport:
		return "traces export: "
	default:
		return "unknown: "
	}
}

// wrappedExportError wraps an OTLP exporter error with the kind of
// signal that produced it.
type wrappedExportError struct {
	wrap error
	kind ErrorKind
}

// WrapTracesError wraps an error from the OTLP exporter for traces.
func WrapTracesError(err error) error {
	return wrappedExportError{
		wrap: err,
		kind: TracesExport,
	}
}

var _ error = wrappedExportError{}

// Error attaches a prefix corresponding to the kind of exporter.
func (t wrappedExportError) Error() string {
	return t.kind.prefix() + t.wrap.Error()
}

// Unwrap returns the wrapped error.
func (t wrappedExportError) Unwrap() error {
	return t.wrap
}
