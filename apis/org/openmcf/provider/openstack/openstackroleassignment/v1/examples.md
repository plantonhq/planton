# OpenStackRoleAssignment Examples

## User + Project Assignment (Literal)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: admin-on-dev-project
spec:
  role_id: "admin-role-uuid"
  project_id: "dev-project-uuid"
  user_id: "phani-user-uuid"
```

## User + Project Assignment (FK Reference)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: member-on-new-project
spec:
  role_id: "member-role-uuid"
  project_id:
    value_from:
      name: arm-dev-sandbox
  user_id: "phani-user-uuid"
```

## Group + Project Assignment

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: team-access
spec:
  role_id: "member-role-uuid"
  project_id:
    value_from:
      name: platform-engineering
  group_id: "platform-team-group-uuid"
```

## User + Domain Assignment

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: domain-admin
spec:
  role_id: "admin-role-uuid"
  domain_id: "default"
  user_id: "admin-user-uuid"
```

## Landing Zone Pattern

```yaml
# In project-landing-zone InfraChart
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: "{{ .Values.projectName }}-admin"
spec:
  role_id: "{{ .Values.adminRoleId }}"
  project_id:
    value_from:
      name: "{{ .Values.projectName }}"
  user_id: "{{ .Values.adminUserId }}"
```
