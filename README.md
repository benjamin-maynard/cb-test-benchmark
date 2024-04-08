# cb-test-benchmark

Quick and dirty tests of concurrent GCS copies using `gcloud` and the Golang SDK.

## Files

1. `cloudbuild-benchmark-gcloud.yaml` - Run `sar` from `sysstat` and then executes 4 parallel `gcloud` copies.
2. `cloudbuild-benchmark-golang.yaml` - Run `sar` from `sysstat` and then executes 4 parallel executions of [custom-copy](custom-copy).
3. `cloudbuild-gcloud.yaml` - Runs 8 concurrent gcloud executions - designed to measure overall copy time.
4. `cloudbuild-golang.yaml` - Runs 8 concurrent [custom-copy](custom-copy) executions - designed to measure overall copy time.
