# Enable Only

Firebase plus Cloud Messaging on the project, nothing else. The shape a
push-only backend needs: the control plane sends notifications, the mobile
apps register for them, and no Firebase storage or enforcement is involved.

## What it configures

- Nothing beyond the enablement: `firebase.googleapis.com` and
  `fcm.googleapis.com` are enabled and the project becomes a Firebase
  project. The `project_number` output is the FCM sender id.

## Adjust before deploying

- **The target project** — enabling Firebase is permanent and one per GCP
  project. Set `projectId` deliberately or rely on the provider default.
- Add the sender's identity alongside: a `GcpServiceAccount` and a
  `GcpProjectIamMember` granting `roles/firebasemessaging.admin`, bound to
  the control plane through Workload Identity.
- Declare the app registrations (`GcpFirebaseAndroidApp`,
  `GcpFirebaseAppleApp`, `GcpFirebaseWebApp`) referencing this resource.

## When to choose something else

Apps that upload content take the **Messaging with Default Bucket** preset;
backends that must reject unattested clients take the **App Check
Enforced** preset.
