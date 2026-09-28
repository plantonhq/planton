# Short Idle Sessions

This preset ends a login session that sits unused for an hour, and every session after a day however active, while keeping the session across browser restarts inside those limits. It suits a console people keep open in a tab: a forgotten laptop signs itself out, and a person who closes and reopens the browser mid-task is not asked to sign in again.

## When to Use

- An admin console, internal tool, or any application where an unattended session is the bigger risk than signing in once more

## Key Configuration Choices

- **Idle timeout** (`idleSessionLifetime: 1`) -- hours a session survives unused ("Inactivity timeout" in the dashboard). Values of an hour or more are whole hours; below an hour, give a fraction (0.5 is 30 minutes). Auth0 caps it at 72 hours outside the Enterprise plan.
- **Absolute lifetime** (`sessionLifetime: 24`) -- hours a session lasts however active ("Require log in after"). Auth0 caps it at 720 hours outside the Enterprise plan.
- **Persistent cookie** (`sessionCookie.mode: persistent`) -- the session outlives the browser; choose `non-persistent` to end it when the browser closes, and set the ephemeral lifetimes instead
- **Unset is unmanaged** -- every other tenant setting keeps its current value

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.idleSessionLifetime` | Hours an unused session survives | Your security policy |
| `spec.sessionLifetime` | Hours before anyone signs in again | Your security policy |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-product-branded-login** -- the page people return to when a session ends
