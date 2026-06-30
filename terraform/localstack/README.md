# Terraform — LocalStack S3 (Infrastructure as Code)

Provisions the S3 test bucket, a lifecycle policy, and a few seed objects against
**LocalStack** — no real AWS account required. This is the IaC alternative to the
Go seeder (`cmd/seed`).

## Run

Start LocalStack first (`make up`, or `docker compose up -d localstack`), then:

**Option A — `tflocal`** (the LocalStack wrapper; auto-points the provider at LocalStack):
```bash
pip install terraform-local
cd terraform/localstack
tflocal init
tflocal apply -auto-approve
```

**Option B — plain Terraform** (the provider is already configured for the local endpoint):
```bash
cd terraform/localstack
terraform init
terraform apply -auto-approve
```

## Verify

```bash
aws --endpoint-url http://localhost:4566 s3 ls s3://production-logs
```

## Notes

- Credentials are the LocalStack dummies (`test`/`test`); all account/credential
  checks are skipped in the provider config.
- `s3_endpoint` defaults to `http://localhost:4566` (host). Set it to
  `http://localstack:4566` to apply from inside the compose network.
- **Scope:** only S3/LocalStack is covered. The `google`/`azurerm` providers do not
  cleanly target fake-gcs-server / Azurite, so GCS and Azure are seeded by the Go
  seeder instead — a deliberate decision, not a gap.
- `terraform destroy` tears the bucket down.
