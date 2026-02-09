# OpenStackProject Examples

## Minimal Project

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: dev-team-alpha
spec: {}
```

## Project with Description and Tags

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: platform-engineering
spec:
  description: Platform engineering team workloads
  tags:
    - team:platform
    - env:production
    - cost-center:engineering
```

## Project in a Specific Domain

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: arm-dev-sandbox
spec:
  description: ARM developer sandbox environment
  domain_id: default
  tags:
    - team:arm
    - purpose:sandbox
```

## Nested Project Hierarchy

```yaml
# Parent project
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: engineering-org
spec:
  description: Top-level engineering organization project

---

# Child project referencing parent
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: eng-team-backend
spec:
  description: Backend team under engineering
  parent_id: "<engineering-org-project-uuid>"
```

## Disabled Project

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: decommissioned-project
spec:
  description: Project disabled for decommissioning
  enabled: false
```

## Project with Region Override

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: region-two-project
spec:
  description: Project managed via RegionTwo Keystone
  region: RegionTwo
```

## Full Landing Zone Pattern (with InfraChart context)

```yaml
# In an InfraChart, this project becomes the foundation
# for downstream resources (networks, security groups, etc.)
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: "{{ .Values.projectName }}"
spec:
  description: "{{ .Values.projectDescription }}"
  tags:
    - "team:{{ .Values.teamName }}"
    - "environment:{{ .Values.environment }}"
```
