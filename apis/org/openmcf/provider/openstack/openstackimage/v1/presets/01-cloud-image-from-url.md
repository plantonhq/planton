# Cloud Image from URL

This preset imports a cloud image into Glance from a URL. The image is downloaded by the Glance service and stored in its backend (Ceph, Swift, filesystem). Most Linux cloud images are distributed as qcow2 files in the bare container format.

## When to Use

- Importing official cloud images (Ubuntu, CentOS, Debian, Fedora) into a new OpenStack deployment
- Registering custom-built images from a CI/CD pipeline
- Any image import where the source is an HTTP/HTTPS URL

## Key Configuration Choices

- **bare container format** -- standard for single-file images (no OVF metadata envelope)
- **qcow2 disk format** -- the most common format for cloud images; supports copy-on-write and sparse allocation
- **URL import** (`imageSourceUrl`) -- Glance downloads the image from the specified URL
- **Private visibility** -- default; only visible to the owning project (change to `shared` or `public` if needed)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<image-download-url>` | HTTP/HTTPS URL of the image file (e.g., `https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img`) | Distro cloud image mirrors |
