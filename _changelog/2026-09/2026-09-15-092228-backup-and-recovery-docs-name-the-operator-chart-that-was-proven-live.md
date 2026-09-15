# The backup and recovery docs name the operator chart that was proven live

**Date**: September 15, 2026
**Type**: Documentation
**Components**: KubernetesPlantonPlatform (GUIDE), the self-hosting docs page "Backup and Recovery"

## Summary

Both places an adopter reads before declaring a platform's backup now say the same floor: operator chart `0.16.3` or newer. Earlier charts declared the same fields, but the first live backup and restore of a self-hosted platform found that their backup engine could not finish installing, that a restore could come back as an empty database in the archive's place when it did not, and that a restored platform's sign-in server could stall; `0.16.3` is the first chart on which the whole path has been proven end to end, and the pages say so plainly instead of naming the two earlier charts that carried the fields.
