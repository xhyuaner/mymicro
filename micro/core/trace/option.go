package trace

/**
telemetry:
	Name: user-srv
	Endpoint: http://127.0.0.1:14268/api/traces
	Sampler: 1.0
	Batcher: jaeger（或者zipkin、otlp）
*/

const TraceName = "xhyuaner"

type Options struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	// 采样率
	Sampler float64 `json:"sampler"`
	// 导出工具
	Batcher string `json:"batcher"`
}
