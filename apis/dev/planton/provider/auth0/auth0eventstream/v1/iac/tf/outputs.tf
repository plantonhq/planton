# Auth0EventStream Outputs
# These outputs match the Auth0EventStreamStackOutputs protobuf message

output "id" {
  description = "The unique identifier of the Auth0 event stream"
  value       = auth0_event_stream.this.id
}

output "name" {
  description = "The name of the event stream"
  value       = auth0_event_stream.this.name
}

output "status" {
  description = "The current status of the event stream"
  value       = auth0_event_stream.this.status
}

output "destination_type" {
  description = "The type of event stream destination (eventbridge or webhook)"
  value       = auth0_event_stream.this.destination_type
}

output "subscriptions" {
  description = "List of event types this stream is subscribed to"
  value       = auth0_event_stream.this.subscriptions
}

output "created_at" {
  description = "The ISO 8601 timestamp when the stream was created"
  value       = auth0_event_stream.this.created_at
}

output "updated_at" {
  description = "The ISO 8601 timestamp when the stream was last updated"
  value       = auth0_event_stream.this.updated_at
}

output "aws_partner_event_source" {
  description = "The AWS partner event source name (only for EventBridge destinations)"
  value       = try(auth0_event_stream.this.eventbridge_configuration[0].aws_partner_event_source, null)
}

