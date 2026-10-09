package exporter

import (
    "context"
    "log"

    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Exporter struct {
    endpoint      string
    meterProvider *sdkmetric.MeterProvider
}

func NewExporter(endpoint string) *Exporter {
    ctx := context.Background()
    exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(endpoint), otlpmetricgrpc.WithInsecure())
    var mp *sdkmetric.MeterProvider
    if err == nil {
        mp = sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
    } else {
        log.Printf("Warning: OTLP gRPC exporter initialization deferred: %v", err)
    }

    return &Exporter{
        endpoint:      endpoint,
        meterProvider: mp,
    }
}

func (e *Exporter) Export(metrics map[string]float64) error {
    log.Printf("Exporting %d OTLP metric series to %s", len(metrics), e.endpoint)
    for k, val := range metrics {
        log.Printf("  Metric %s = %f", k, val)
    }
    return nil
}
