output "name" {
  description = "Full resource name of the processor (projects/{project}/locations/{location}/processors/{processor_id})"
  value       = google_document_ai_processor.this.id
}

output "processor_id" {
  description = "The id Google assigned the processor"
  value       = google_document_ai_processor.this.name
}

output "location" {
  description = "The location the processor lives in"
  value       = google_document_ai_processor.this.location
}

output "process_endpoint" {
  description = "The REST endpoint documents are posted to"
  value       = "https://${google_document_ai_processor.this.location}-documentai.googleapis.com/v1/${google_document_ai_processor.this.id}:process"
}
