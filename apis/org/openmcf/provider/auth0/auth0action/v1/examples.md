# Auth0Action Examples

## Post-Login: Enrich Tokens with Custom Claims

Add organization roles and user email to ID and access tokens.

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Action
metadata:
  name: enrich-token-claims
  org: acme-corp
  env: production
spec:
  supported_trigger:
    id: post-login
    version: v3
  code: |
    exports.onExecutePostLogin = async (event, api) => {
      const namespace = 'https://myapp.example.com';
      api.idToken.setCustomClaim(`${namespace}/roles`, event.authorization?.roles || []);
      api.accessToken.setCustomClaim(`${namespace}/email`, event.user.email);
      api.accessToken.setCustomClaim(`${namespace}/org_id`, event.organization?.id || '');
    };
  deploy: true
  trigger_binding:
    display_name: Enrich Token Claims
```

## Pre-Registration: Email Domain Allowlist

Only allow users with specific email domains to register.

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Action
metadata:
  name: validate-email-domain
  org: acme-corp
  env: production
spec:
  supported_trigger:
    id: pre-user-registration
    version: v2
  code: |
    exports.onExecutePreUserRegistration = async (event, api) => {
      const allowedDomains = event.secrets.ALLOWED_DOMAINS.split(',');
      const domain = event.user.email.split('@')[1];
      if (!allowedDomains.includes(domain)) {
        api.access.deny('registration_denied', 'Your email domain is not allowed to register.');
      }
    };
  runtime: node22
  deploy: true
  secrets:
    - name: ALLOWED_DOMAINS
      value: "acme.com,acme-corp.com"
  trigger_binding: {}
```

## Post-Login: Slack Notification on Login

Send a Slack message when a user logs in from a new device.

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Action
metadata:
  name: slack-login-alert
  org: acme-corp
  env: production
spec:
  supported_trigger:
    id: post-login
    version: v3
  code: |
    const axios = require('axios');

    exports.onExecutePostLogin = async (event, api) => {
      if (event.stats.logins_count <= 1) {
        await axios.post(event.secrets.SLACK_WEBHOOK_URL, {
          text: `:wave: New user signed up: ${event.user.email} via ${event.connection.name}`
        });
      }
    };
  deploy: true
  dependencies:
    - name: axios
      version: "1.7.0"
  secrets:
    - name: SLACK_WEBHOOK_URL
      value: "https://hooks.slack.com/services/T00/B00/xxx"
  trigger_binding:
    display_name: Slack Login Alert
```

## Credentials Exchange: M2M Audit Logging

Log machine-to-machine token exchanges for compliance.

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Action
metadata:
  name: m2m-audit-log
  org: acme-corp
  env: production
spec:
  supported_trigger:
    id: credentials-exchange
    version: v2
  code: |
    const axios = require('axios');

    exports.onExecuteCredentialsExchange = async (event, api) => {
      await axios.post(event.secrets.AUDIT_ENDPOINT, {
        timestamp: new Date().toISOString(),
        client_id: event.client.client_id,
        client_name: event.client.name,
        audience: event.request.body.audience,
        scopes: event.request.body.scope,
      }, {
        headers: { 'Authorization': `Bearer ${event.secrets.AUDIT_TOKEN}` }
      });
    };
  deploy: true
  dependencies:
    - name: axios
      version: "1.7.0"
  secrets:
    - name: AUDIT_ENDPOINT
      value: "https://audit.example.com/events"
    - name: AUDIT_TOKEN
      value: "audit-bearer-token"
  trigger_binding:
    display_name: M2M Audit Log
```

## Send Phone Message: Custom Twilio SMS Provider

Use Twilio as a custom SMS provider for MFA.

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Action
metadata:
  name: twilio-sms-provider
  org: acme-corp
  env: production
spec:
  supported_trigger:
    id: send-phone-message
    version: v2
  code: |
    const twilio = require('twilio');

    exports.onExecuteSendPhoneMessage = async (event, api) => {
      const client = twilio(event.secrets.TWILIO_SID, event.secrets.TWILIO_TOKEN);
      await client.messages.create({
        body: event.message_options.text,
        to: event.message_options.recipient,
        from: event.secrets.TWILIO_FROM,
      });
    };
  deploy: true
  dependencies:
    - name: twilio
      version: "4.23.0"
  secrets:
    - name: TWILIO_SID
      value: "AC1234567890abcdef"
    - name: TWILIO_TOKEN
      value: "auth-token-value"
    - name: TWILIO_FROM
      value: "+15551234567"
  trigger_binding:
    display_name: Twilio SMS Provider
```

## Unbound Action (No Trigger Binding)

Create and deploy an action without binding it to a trigger. Useful when trigger ordering is managed externally.

```yaml
apiVersion: auth0.openmcf.org/v1
kind: Auth0Action
metadata:
  name: mfa-challenge
  org: acme-corp
  env: staging
spec:
  supported_trigger:
    id: post-login
    version: v3
  code: |
    exports.onExecutePostLogin = async (event, api) => {
      if (!event.authentication?.methods?.find(m => m.name === 'mfa')) {
        api.authentication.challengeWithAny([{ type: 'otp' }, { type: 'push-notification' }]);
      }
    };
  deploy: true
```
