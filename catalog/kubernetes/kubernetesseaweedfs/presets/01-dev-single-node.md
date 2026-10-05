# Dev single node preset

The smallest useful SeaweedFS: one master, one volume server and one
filer on small PersistentVolumeClaims, the S3 gateway embedded on the
filer with authentication on (the default — the chart materializes
admin and read-only credential pairs in the `dev-seaweedfs-s3-secret`
Secret), and one bucket created at install. For developers who need a
real in-cluster S3 endpoint to point an SDK at, without a cloud
bucket or a heavyweight storage system.

Know the shape's limits: a single volume server means no replication
(the `replication` code would have nowhere to place a copy), and the
single filer's embedded leveldb metadata store is per-pod — this
preset is one node of everything, with availability equal to the
StatefulSets rescheduling their pods and the PVCs surviving it.
Nothing is exposed outside the cluster; SDKs use the exported
`s3_endpoint` (path-style, port 8333) with the credentials from the
Secret in the outputs.

Change first: the bucket name (and add one entry per bucket your app
needs — the hook creates them at install), then `volume.data_volume.size`
if 10Gi will not hold your data; the volume tier is the only one that
grows with stored bytes, and growing a PVC later depends on the
StorageClass allowing expansion.

See [01-dev-single-node.yaml](./01-dev-single-node.yaml) for the
manifest.
