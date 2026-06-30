output "bucket" {
  description = "The provisioned bucket name."
  value       = aws_s3_bucket.logs.id
}

output "seeded_keys" {
  description = "Object keys seeded into the bucket."
  value       = sort(keys(var.seed_objects))
}
