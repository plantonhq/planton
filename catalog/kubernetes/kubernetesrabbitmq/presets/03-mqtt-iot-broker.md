# MQTT IoT broker preset

A RabbitMQ cluster serving an IoT / device fleet over MQTT 5.0 on the
shared broker core: 3 nodes (device fleets reconnect in herds, so
availability matters from day one), the rabbitmq_mqtt and
rabbitmq_web_mqtt plugins on top of the always-on essentials, a
LoadBalancer client Service so devices outside the cluster get a
reachable address, and 2Gi of memory per node with requests equal to
limits (the memory-high-watermark rule).

Know what the plugins buy: MQTT is not a separate broker — it is a
protocol listener on the same RabbitMQ core, so MQTT traffic from
devices and AMQP traffic from backend consumers meet in one cluster,
one set of credentials, one management UI. rabbitmq_web_mqtt adds the
WebSocket form for browser dashboards and WebSocket-only networks.
Changing the plugin list later rolls the cluster — plan plugin
changes like the restarts they are. The LoadBalancer Service is the
deliberate exposure surface (this component creates no ingress); the
sample annotation is the AWS NLB recipe, and your cloud's LB
controller has its own vocabulary.

Change first: the Service annotation, for your cloud. Then declare
`tls` before real devices connect — a fleet over the public internet
should not speak plaintext, and `disable_non_tls_listeners: true`
closes every plain port, including the plain ports of the enabled
plugins (their WebSocket forms too), once every client speaks TLS.
Backend consumers read credentials from the operator-generated
`iot-broker-default-user` Secret exported in the outputs.

See [03-mqtt-iot-broker.yaml](./03-mqtt-iot-broker.yaml) for the
manifest.
