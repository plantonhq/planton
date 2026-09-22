# GcpNetworkEndpointGroup Guide

The judgment this guide protects: a network endpoint group is the list of
places a load balancer may send traffic. Getting the scope and endpoint
type right up front matters because both are immutable; getting the list
right matters because the list IS the membership.

## Pick the scope from the load balancer, not the endpoints

`zone` set builds a ZONAL group: VMs by instance (`GCE_VM_IP_PORT` for
Application and proxy load balancers, `GCE_VM_IP` for passthrough Network
Load Balancers), hybrid addresses reached over VPN or Interconnect
(`NON_GCP_PRIVATE_IP_PORT`), or internet endpoints for a REGIONAL external
Application Load Balancer. `zone` empty builds a GLOBAL internet group --
the only shape a GLOBAL external Application Load Balancer accepts for an
origin outside Google Cloud. Serverless (Cloud Run, Cloud Functions, App
Engine), Private Service Connect, and regional internet groups are a
different resource family: `GcpRegionNetworkEndpointGroup`. The scope, the
type, the network, and the name all recreate the group when changed.

## The list is the membership

`endpoints` is written as one set (zonal groups use Google's bulk endpoint
operation; global groups get one endpoint resource each). An endpoint you
remove from the manifest is detached; one you add is attached; nothing
outside the list survives an apply. If an autoscaler or another controller
manages membership, leave `endpoints` empty and let it -- a group declared
with no endpoints is valid and carries no endpoint resource at all.

## Which fields an endpoint carries

The group's type decides: `GCE_VM_IP_PORT` endpoints name an `instance`
(a `GcpComputeInstance` reference) and optionally an `ipAddress` (a
primary or alias IP of that VM; unset means the primary) and `port`
(unset means `defaultPort`); `GCE_VM_IP` endpoints name an instance and
no port at all -- the passthrough load balancer forwards every port;
hybrid and `INTERNET_IP_PORT` endpoints are an `ipAddress` and a `port`;
`INTERNET_FQDN_PORT` endpoints are an `fqdn` and a `port`. Every mismatch
is rejected before deploy.

## Hybrid groups have a scheme rule

`NON_GCP_PRIVATE_IP_PORT` groups are accepted only by backend services
with the `EXTERNAL`, `EXTERNAL_MANAGED`, `INTERNAL_MANAGED`, or
`INTERNAL_SELF_MANAGED` scheme in RATE or CONNECTION balancing mode --
Google's rule, reported at the backend service, so choose the backend
service's shape with the group in mind.

## Teardown

Google refuses to delete a group a backend service still uses: detach it
from the backend service first. `deletionPolicy` unset means DELETE (the
group and its endpoints go); `PREVENT` makes destroy fail; `ABANDON` leaves
the group serving but unmanaged.
