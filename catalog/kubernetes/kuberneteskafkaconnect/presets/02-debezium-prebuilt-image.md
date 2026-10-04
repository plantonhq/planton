# Debezium prebuilt image preset

The prebuilt-image arm in production shape: three workers running an
image that already carries the Debezium connector plugins (the
`image` reference here is an example — point it at the Debezium
Connect image your team consumes, or the output of a previous
operator `build`), connected to a TLS + SCRAM Kafka cluster with
internal topics at replication factor 3, sized JVM/resources, and JMX
Prometheus metrics.

This is the fastest path to CDC when someone else maintains the
plugin artifact set: no build machinery, no registry push from the
operator — versioning the plugins means retagging the image. When you
need an exact artifact set baked reproducibly instead, use the
operator-built-image preset (the `image` and `build` arms are
mutually exclusive, spec-enforced).

The TLS and authentication blocks are the composition seams: in an
Infra Chart the CA Secret comes from a KubernetesKafka reference and
the SCRAM credential Secret from a KubernetesKafkaUser reference.
Declare the actual pipes as KubernetesKafkaConnector resources in
this namespace — the debezium-postgres-cdc connector preset pairs
with this cluster.

See [02-debezium-prebuilt-image.yaml](./02-debezium-prebuilt-image.yaml)
for the manifest.
