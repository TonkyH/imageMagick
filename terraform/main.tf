provider "google" {
  project = var.project_id
  region  = "asia-northeast1"
}

resource "random_id" "bucket_suffix" {
  byte_length = 8
}

// gcs
resource "google_storage_bucket" "imageproc_input" {
  name          = "input-bucket-${random_id.bucket_suffix.hex}"
  location      = "asia-northeast1"
  force_destroy = true
}

output "input_bucket_name" {
  value = google_storage_bucket.imageproc_input.name
}

resource "google_storage_bucket" "imageproc_output" {
  name          = "output-bucket-${random_id.bucket_suffix.hex}"
  location      = "asia-northeast1"
  force_destroy = true
}

output "blurred_bucket_name" {
  value = google_storage_bucket.imageproc_output.name
}

data "google_storage_project_service_account" "gcs_account" {}

resource "google_pubsub_topic_iam_binding" "binding" {
  topic   = google_pubsub_topic.default.name
  role    = "roles/pubsub.publisher"
  members = ["serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"]
}

resource "google_storage_notification" "notification" {
  bucket         = google_storage_bucket.imageproc_input.name
  payload_format = "JSON_API_V1"
  topic          = google_pubsub_topic.default.id
  depends_on     = [google_pubsub_topic_iam_binding.binding]
}

// pubsub
resource "google_pubsub_topic" "default" {
  name = "pubsub_topic"
}

resource "google_pubsub_subscription" "subscription" {
  name  = "pubsub_subscription"
  topic = google_pubsub_topic.default.name
  push_config {
    push_endpoint = google_cloud_run_v2_service.default.uri
    oidc_token {
      service_account_email = google_service_account.sa.email
    }
    attributes = {
      x-goog-version = "v1"
    }
  }
  depends_on = [google_cloud_run_v2_service.default]
}

// service account
resource "google_service_account" "sa" {
  account_id   = "cloud-run-pubsub-invoker"
  display_name = "Cloud Run Pub/Sub Invoker"
}

resource "google_cloud_run_service_iam_binding" "binding" {
  location = google_cloud_run_v2_service.default.location
  service  = google_cloud_run_v2_service.default.name
  role     = "roles/run.invoker"
  members  = ["serviceAccount:${google_service_account.sa.email}"]
}

resource "google_project_service_identity" "pubsub_agent" {
  provider = google-beta
  project  = var.project_id
  service  = "pubsub.googleapis.com"
}

resource "google_project_iam_binding" "project_token_creator" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  members = ["serviceAccount:${google_project_service_identity.pubsub_agent.email}"]
}

# Enable Cloud Run API
resource "google_project_service" "cloudrun_api" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

// cloud run
resource "google_cloud_run_v2_service" "default" {
  name     = "pusub-tutorial"
  location = "asia-northeast1"
  template {
    containers {
      image = var.application_image
      env {
        name  = "BLURRED_BUCKET_NAME"
        value = google_storage_bucket.imageproc_output.name
      }
      ports {
        container_port = 8080
      }
    }
  }
  depends_on = [google_project_service.cloudrun_api, google_storage_bucket.imageproc_output]
}
