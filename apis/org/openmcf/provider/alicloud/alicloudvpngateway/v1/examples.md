# AliCloudVpnGateway Examples

## Minimal: Single Site-to-Site Connection

The simplest VPN setup: one gateway with a single IPsec connection to a remote office.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudVpnGateway
metadata:
  name: office-vpn
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  vswitchId:
    value: vsw-abc123
  vpnGatewayName: office-vpn
  bandwidth: 10
  connections:
    - name: office-hq
      customerGatewayIp: "203.0.113.1"
      localSubnets:
        - "10.0.0.0/8"
      remoteSubnets:
        - "192.168.0.0/16"
```

## Production: Multi-Site with Custom IKE/IPsec

VPN Gateway connecting to two remote sites with strong encryption and health checks.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudVpnGateway
metadata:
  name: prod-vpn
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  vpcId:
    valueFrom:
      name: prod-vpc
  vswitchId:
    valueFrom:
      name: vpn-vswitch
  vpnGatewayName: prod-vpn-gateway
  description: Production VPN for datacenter connectivity
  bandwidth: 100
  tags:
    team: network
    costCenter: shared-infra
  connections:
    - name: datacenter-primary
      customerGatewayIp: "198.51.100.1"
      localSubnets:
        - "10.0.0.0/8"
        - "172.16.0.0/12"
      remoteSubnets:
        - "192.168.1.0/24"
        - "192.168.2.0/24"
      ikeConfig:
        psk: "strong-secret-key-dc1"
        ikeVersion: ikev2
        ikeEncAlg: aes256
        ikeAuthAlg: sha256
        ikePfs: group14
      ipsecConfig:
        ipsecEncAlg: aes256
        ipsecAuthAlg: sha256
        ipsecPfs: group14
      healthCheckConfig:
        enable: true
        sip: "10.0.0.1"
        dip: "192.168.1.1"
        interval: 5
        retry: 3
    - name: datacenter-dr
      customerGatewayIp: "198.51.100.2"
      localSubnets:
        - "10.0.0.0/8"
      remoteSubnets:
        - "192.168.10.0/24"
      ikeConfig:
        psk: "strong-secret-key-dc2"
        ikeVersion: ikev2
        ikeEncAlg: aes256
        ikeAuthAlg: sha256
        ikePfs: group14
      ipsecConfig:
        ipsecEncAlg: aes256
        ipsecAuthAlg: sha256
        ipsecPfs: group14
      healthCheckConfig:
        enable: true
        sip: "10.0.0.1"
        dip: "192.168.10.1"
```

## SSL VPN Enabled

VPN Gateway with SSL VPN for remote client access alongside a site-to-site connection.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudVpnGateway
metadata:
  name: hybrid-vpn
spec:
  region: ap-southeast-1
  vpcId:
    value: vpc-sea1
  vswitchId:
    value: vsw-sea1a
  vpnGatewayName: hybrid-vpn
  bandwidth: 50
  enableSsl: true
  sslConnections: 50
  connections:
    - name: singapore-office
      customerGatewayIp: "203.0.113.10"
      localSubnets:
        - "10.0.0.0/8"
      remoteSubnets:
        - "172.20.0.0/16"
```

## Cross-Reference with valueFrom

VPN Gateway referencing VPC and VSwitch from other OpenMCF resources.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudVpnGateway
metadata:
  name: ref-vpn
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: shared-vpc
  vswitchId:
    valueFrom:
      name: vpn-vswitch
  vpnGatewayName: ref-vpn
  bandwidth: 20
  connections:
    - name: branch-office
      customerGatewayIp: "198.51.100.50"
      customerGatewayAsn: "65001"
      localSubnets:
        - "10.0.0.0/8"
      remoteSubnets:
        - "192.168.100.0/24"
      enableDpd: true
      enableNatTraversal: true
```
