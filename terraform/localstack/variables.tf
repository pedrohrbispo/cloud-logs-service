variable "region" {
  description = "AWS region (LocalStack ignores it but the provider requires one)."
  type        = string
  default     = "us-east-1"
}

variable "s3_endpoint" {
  description = "S3 endpoint. localhost for host-run terraform; localstack:4566 inside the compose network."
  type        = string
  default     = "http://localhost:4566"
}

variable "bucket" {
  description = "Bucket to provision."
  type        = string
  default     = "production-logs"
}

variable "seed_objects" {
  description = "Map of object key -> content to seed into the bucket."
  type        = map(string)
  default = {
    "payment.log" = "2026-06-29T14:32:01Z INFO  payment-service charge ok order=ord_8812 amount=42.00 currency=USD\n"
    "auth.log"    = "2026-06-29T09:14:07Z WARN  auth-service login failed user=ghost@example.com reason=invalid_credentials\n"
    "worker.log"  = "2026-06-27T22:01:30Z ERROR worker email job failed id=job_551 error=\"smtp timeout\" giving_up=true\n"
  }
}
