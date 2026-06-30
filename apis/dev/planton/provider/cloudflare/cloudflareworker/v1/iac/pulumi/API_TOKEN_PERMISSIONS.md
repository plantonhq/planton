# Cloudflare API Token Permissions for Worker Deployment

## Required Permissions

The Cloudflare API token used for deploying Workers needs the following permissions:

### Account-Level Permissions

1. **Workers Scripts: Edit**
   - Scope: Account
   - Required for: Creating and updating Worker scripts
   - Permission: `com.cloudflare.api.account.worker.script`

2. **Workers Routes: Edit**
   - Scope: Zone
   - Required for: Creating Worker routes (attaching Workers to domains)
   - Permission: `com.cloudflare.api.account.worker.route`

3. **Workers KV Storage: Edit** (if using KV bindings)
   - Scope: Account
   - Required for: Binding KV namespaces to Workers
   - Permission: `com.cloudflare.api.account.storage.kv`

### Zone-Level Permissions

If you're creating Worker routes (specified via `route_pattern` and `zone_id`):

1. **Zone: Read**
   - Scope: Specific Zone or All Zones
   - Required for: Validating zone existence and reading zone details
   - Permission: `com.cloudflare.api.zone`

## Creating the API Token

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com/) → My Profile → API Tokens
2. Click "Create Token"
3. Select "Create Custom Token"
4. Configure permissions:

   ```
   Account Permissions:
   - Workers Scripts: Edit
   - Workers KV Storage: Edit (if using KV)
   
   Zone Permissions:
   - Workers Routes: Edit
   - Zone: Read
   ```

5. **Account Resources:**
   - Include: Specific account or All accounts
   
6. **Zone Resources:**
   - Include: Specific zone (e.g., planton.live) or All zones

7. Set token expiration (optional)
8. Click "Continue to Summary"
9. Click "Create Token"
10. **Save the token securely** - it will only be shown once

## Common Errors

### Authentication Error (10000)

```
error: 1 error occurred:
    * error creating worker route: Authentication error (10000)
```

**Cause:** The API token doesn't have "Workers Routes: Edit" permission for the zone.

**Solution:** 
1. Verify the token has "Workers Routes: Edit" permission
2. Ensure the token has access to the specific zone (check Zone Resources)
3. Recreate the token with correct permissions if needed

### Insufficient Permissions

```
error: error creating worker script: Permission denied (10000)
```

**Cause:** Missing "Workers Scripts: Edit" permission.

**Solution:** Add "Workers Scripts: Edit" permission to the API token.

## Environment Variables

Set the API token as an environment variable:

```bash
export CLOUDFLARE_API_TOKEN=your_api_token_here
```

Or provide it via the Cloudflare provider configuration in your deployment pipeline.

## Best Practices

1. **Use separate tokens** for different environments (dev, staging, prod)
2. **Limit token scope** to only the accounts and zones needed
3. **Set token expiration** and rotate regularly
4. **Store tokens securely** using secrets management (e.g., HashiCorp Vault, AWS Secrets Manager)
5. **Never commit tokens** to version control

## Verification

After creating the token, you can verify it has the correct permissions:

```bash
curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json"
```

The response will show the token's permissions and status.

