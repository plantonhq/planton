# Mail Server IPv4 with Reverse DNS

This preset allocates a persistent public IPv4 address with a reverse DNS (rDNS) record and delete protection enabled. It is designed for mail servers and any service where clients verify the server's identity through reverse DNS lookups. The IaC module creates both an `hcloud_primary_ip` and an `hcloud_rdns` resource, linking the rDNS pointer to the allocated IP address.

Mail deliverability depends heavily on matching forward and reverse DNS. Receiving mail servers routinely reject messages from IPs whose rDNS does not resolve to the sending domain. Losing a mail server's IP address also means losing the IP reputation built over months of legitimate sending, so delete protection is warranted.

## When to Use

- Mail servers (Postfix, Exim, Exchange) where SPF, DKIM, and DMARC verification require matching forward and reverse DNS
- SMTP relay or transactional email services that need a clean IP with verifiable rDNS
- VPN endpoints, monitoring systems, or any service where clients perform reverse DNS lookups for identity verification

## Key Configuration Choices

- **IPv4** (`type: ipv4`) -- mail delivery still overwhelmingly relies on IPv4; most spam blocklists and reputation systems track IPv4 addresses
- **Reverse DNS** (`dnsPtr: <mail-server-hostname>`) -- sets the PTR record for the allocated IP; must match the hostname in your MX record and the server's HELO/EHLO greeting for reliable delivery
- **Delete protection** (`deleteProtection: true`) -- prevents accidental destruction of an IP whose reputation has been built over time; must be explicitly disabled before the resource can be removed
- **Falkenstein location** (`location: fsn1`) -- Hetzner's largest datacenter; change to match your server's location since Primary IPs can only be assigned to servers in the same location

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<mail-server-hostname>` | Fully qualified hostname the IP should reverse-resolve to (e.g., `mail.example.com`) | Your DNS zone's MX record target; must also match the server's HELO/EHLO identity |

## Related Presets

- **01-standard-ipv4** -- simpler variant without rDNS or delete protection, for servers that do not require reverse DNS verification
