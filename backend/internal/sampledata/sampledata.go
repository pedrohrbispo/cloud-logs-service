// Package sampledata holds the seed log files uploaded to the emulators at
// startup, so both the seeder and tests share one source of truth.
package sampledata

// LogFile is one seed object.
type LogFile struct {
	Name    string
	Content string
}

// Logs returns the sample log files. Their names/levels reproduce the design's
// info/warn/error mix when the frontend infers level from the filename.
func Logs() []LogFile {
	return []LogFile{
		{
			Name: "payment.log",
			Content: `2026-06-29T14:32:01Z INFO  payment-service charge ok order=ord_8812 amount=42.00 currency=USD
2026-06-29T14:32:03Z INFO  payment-service charge ok order=ord_8813 amount=18.50 currency=USD
2026-06-29T14:32:09Z INFO  payment-service refund ok order=ord_8790 amount=12.00 currency=USD
`,
		},
		{
			Name: "auth.log",
			Content: `2026-06-29T09:14:00Z INFO  auth-service login ok user=admin@example.com ip=10.0.0.4
2026-06-29T09:14:07Z WARN  auth-service login failed user=ghost@example.com reason=invalid_credentials ip=10.0.0.9
2026-06-29T09:15:21Z WARN  auth-service token refresh throttled user=viewer@example.com
`,
		},
		{
			Name: "gateway.log",
			Content: `2026-06-29T11:05:00Z INFO  api-gateway GET /v1/orders 200 12ms
2026-06-29T11:05:01Z INFO  api-gateway GET /v1/logs 200 8ms
2026-06-29T11:05:03Z INFO  api-gateway POST /v1/auth/login 200 251ms
`,
		},
		{
			Name: "orders.log",
			Content: `2026-06-29T08:47:00Z INFO  orders-service created order=ord_8801 items=3
2026-06-29T08:47:11Z INFO  orders-service shipped order=ord_8794 carrier=dhl
2026-06-29T08:47:40Z INFO  orders-service created order=ord_8802 items=1
`,
		},
		{
			Name: "worker.log",
			Content: `2026-06-27T22:01:00Z INFO  worker email job started id=job_551
2026-06-27T22:01:02Z ERROR worker email job failed id=job_551 error="smtp timeout" retry=1
2026-06-27T22:01:30Z ERROR worker email job failed id=job_551 error="smtp timeout" retry=2 giving_up=true
`,
		},
	}
}
