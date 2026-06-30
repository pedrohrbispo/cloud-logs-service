# Provisions the S3 test bucket (and seeds it) against LocalStack. The provider
# is pointed at the local endpoint and all the AWS account/credential checks are
# skipped, so this applies with no real cloud account. Run with `tflocal` or with
# plain `terraform` (the endpoint override below makes plain terraform work too).

provider "aws" {
  region     = var.region
  access_key = "test"
  secret_key = "test"

  # Make the AWS provider talk to LocalStack instead of real AWS.
  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    s3 = var.s3_endpoint
  }
}

resource "aws_s3_bucket" "logs" {
  bucket = var.bucket
}

# Demonstrate a real lifecycle policy: expire log objects after 30 days.
resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    id     = "expire-old-logs"
    status = "Enabled"
    filter {}
    expiration {
      days = 30
    }
  }
}

# Seed the bucket with sample log objects (an IaC alternative to the Go seeder).
resource "aws_s3_object" "seed" {
  for_each = var.seed_objects

  bucket        = aws_s3_bucket.logs.id
  key           = each.key
  content       = each.value
  content_type  = "text/plain; charset=utf-8"
  cache_control = "private, no-store"
}
