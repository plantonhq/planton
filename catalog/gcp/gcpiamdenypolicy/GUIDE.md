# GcpIamDenyPolicy Guide

The judgment this guide protects: a deny policy is not another access
grant — it is the layer that overrides all of them. Deny always outranks
allow, so every mistake here is either a lockout (denied the wrong
principal) or a silent hole (removed a guardrail nobody noticed).

## Guardrails, not allow-policies

Allow-policies answer "who may do this"; deny policies answer "who may
NEVER do this, no matter what roles say". A project owner with every
role in the project is still blocked by a deny rule. Use deny for
invariants — nobody reads break-glass secrets, nobody deletes projects —
and keep day-to-day access in allow-policies. A deny policy used as
routine access control turns every onboarding into a policy edit.

## Always model the break-glass path

A denial with no exception is a lockout waiting for its incident. Put
the recovery identity — an on-call service account, a dedicated
break-glass account — in `exceptionPrincipals` so the rule denies
everyone (`principalSet://goog/public:all`) EXCEPT the path you will
need at 3am. Verify the exception identity actually works before the
policy ships; deny is not a layer to debug during an outage. Create the break-glass identity first, too: Google refuses a rule
naming a service account that does not exist ("The principal specified by
member ... was not found"), so the account must exist before the policy
that excepts it.

## Conditions carve out spaces, not people

`denialCondition` is a CEL expression over resource tags — the idiom for
"deny everywhere except tagged sandboxes":

    !resource.matchTag('12345678/env', 'sandbox')

People go in the exception lists; environments go in the condition. Tag
the sandbox hierarchy once and the guardrail leaves it alone without
enumerating projects.

## Destroy must not be silent

`deletionPolicy: PREVENT` is the right setting once a deny policy guards
anything real. Deleting a guardrail re-opens the surface it guards with
no incident-side symptom — nothing breaks, nothing pages; the denied
permission simply works again for whoever holds a role. `PREVENT` makes
removal a deliberate two-step (lift the policy setting, then destroy)
instead of a side effect.

## Org permissions are the point, not an obstacle

Creating deny policies requires org-level `roles/iam.denyAdmin` even for
project-attached policies. That is Google saying guardrails belong to
the platform team: project teams cannot quietly remove the rules that
constrain them. Plan ownership accordingly — deny policies ship through
the platform team's pipeline, with its org-scoped credentials.
