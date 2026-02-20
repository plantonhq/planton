# OciDnsRecord produces no outputs.
#
# DNS record sets do not generate an OCID or any composable identifier
# that downstream components would reference. The resource is identified
# by the (zone, domain, rtype) tuple, all of which are inputs.
