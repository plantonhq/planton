# OpenStackImage Examples

## Ubuntu 22.04 Cloud Image (Most Common)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackImage
metadata:
  name: ubuntu-22-04
spec:
  container_format: bare
  disk_format: qcow2
  image_source_url: https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img
  min_disk_gb: 10
  min_ram_mb: 512
  tags:
    - ubuntu
    - "22.04"
    - cloud-init
```

## CentOS 9 Stream

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackImage
metadata:
  name: centos-9-stream
spec:
  container_format: bare
  disk_format: qcow2
  image_source_url: https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-GenericCloud-9-latest.x86_64.qcow2
  min_disk_gb: 10
  min_ram_mb: 1024
  tags:
    - centos
    - "9-stream"
```

## Protected Production Image

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackImage
metadata:
  name: golden-app-image-v3
spec:
  container_format: bare
  disk_format: qcow2
  image_source_url: https://artifacts.example.com/images/golden-app-v3.qcow2
  min_disk_gb: 20
  min_ram_mb: 2048
  protected: true
  visibility: shared
  tags:
    - production
    - golden
    - v3
```

## Minimal Metadata-Only Image

```yaml
# Image data uploaded separately via glance CLI
apiVersion: openstack.openmcf.org/v1
kind: OpenStackImage
metadata:
  name: custom-appliance
spec:
  container_format: bare
  disk_format: raw
```

## ISO Image

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackImage
metadata:
  name: ubuntu-22-04-installer
spec:
  container_format: bare
  disk_format: iso
  image_source_url: https://releases.ubuntu.com/22.04/ubuntu-22.04.3-live-server-amd64.iso
  min_disk_gb: 25
  min_ram_mb: 2048
```

## Hidden Development Image

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackImage
metadata:
  name: dev-debug-image
spec:
  container_format: bare
  disk_format: qcow2
  image_source_url: https://artifacts.example.com/images/debug-tools.qcow2
  hidden: true
  tags:
    - development
    - debug
```

## Public Community Image (Admin Required)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackImage
metadata:
  name: company-base-image
spec:
  container_format: bare
  disk_format: qcow2
  image_source_url: https://artifacts.example.com/images/base-v1.qcow2
  visibility: public
  min_disk_gb: 10
  min_ram_mb: 512
  tags:
    - base
    - company-standard
```

## InfraChart Template Usage

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackImage
metadata:
  name: "{{ .Values.imageName }}"
spec:
  container_format: bare
  disk_format: qcow2
  image_source_url: "{{ .Values.imageSourceUrl }}"
  min_disk_gb: {{ .Values.minDiskGb | default 10 }}
  min_ram_mb: {{ .Values.minRamMb | default 512 }}
  tags:
    - "managed-by:planton"
```
