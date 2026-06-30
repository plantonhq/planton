# Standard Pool Member

This preset adds a backend server to an Octavia pool. Each member has an IP address and port that the pool forwards traffic to based on its load-balancing algorithm. Create one member resource per backend server.

## When to Use

- Adding a backend instance to a load balancer pool
- Any server that should receive traffic from the load balancer

## Key Configuration Choices

- **Explicit subnet** (`subnetId`) -- tells Octavia which subnet the member lives on for L3 routing; important when the member is on a different subnet than the VIP
- **Default weight** -- weight is unset (Octavia defaults to 1, equal distribution); set explicitly for weighted balancing
- **Admin state up** -- default (true), member receives traffic immediately

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<pool-id>` | ID of the pool to add this member to | OpenStack console or `OpenStackLoadBalancerPool` status outputs |
| `<backend-ip-address>` | IP address of the backend server (e.g., `192.168.1.10`) | Instance details or `OpenStackInstance` status outputs |
| `<backend-port>` | Port the backend listens on (e.g., `8080`) | Your application configuration |
| `<member-subnet-id>` | ID of the subnet where the backend resides | OpenStack console or `OpenStackSubnet` status outputs |
