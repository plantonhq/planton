# The script the agent-reader Job runs, delivered through the module-owned
# `<name>-agent-reader-script` ConfigMap (agent_reader.tf). PARITY: the
# Pulumi module carries this exact content as the Go constant
# agentReaderScript (scripts.go); the ONLY difference is HCL's escaping of
# `%{` as `%%{` -- the rendered ConfigMap must stay byte-identical across
# engines, and the `# parity:` marker above the heredoc lets the
# repository's cross-engine script parity guard prove it on every change.
# The Job's name hashes this text (local.agent_reader_job_name), so a
# changed script is a new run on the next apply -- safe, because the script
# is idempotent. A plain (non-indented) heredoc on purpose: the rendered
# text must match the Go constant byte for byte.
#
# Environment contract: see scripts.go.
locals {
  # parity: scripts.go agentReaderScript
  agent_reader_script = <<EOT
#!/bin/sh
# Keeps agent teammates' read-only way into Grafana: one Viewer service
# account, and one current token for it in a Secret owned by this Job's
# ServiceAccount. A run whose stored token still answers at the declared
# generation changes nothing. With DISABLED=true it refuses every token of
# the account instead. Every failure names what happened and the next step.
set -eu

# The release is Ready before this Job starts, but a restart or a rollout
# can still be in flight: wait for Grafana's own health call, bounded.
waited=0
until [ "$(curl -s -o /dev/null -w '%%{http_code}' "$GRAFANA_URL/api/health")" = "200" ]; do
  if [ "$waited" -ge 300 ]; then
    echo "Grafana at $GRAFANA_URL did not answer /api/health for 300 seconds." >&2
    echo "Check its pods: kubectl -n $NAMESPACE get pods -l app.kubernetes.io/instance=$RELEASE_NAME" >&2
    exit 1
  fi
  sleep 5
  waited=$((waited + 5))
done

api() {
  method=$1
  path=$2
  shift 2
  curl -sS -f -u "$GRAFANA_ADMIN_USER:$GRAFANA_ADMIN_PASSWORD" -X "$method" \
    -H 'Content-Type: application/json' "$GRAFANA_URL$path" "$@"
}

if ! api GET /api/org >/dev/null; then
  echo "Grafana refused the admin credentials read from Secret $ADMIN_SECRET." >&2
  echo "The Job signs in with them to manage the service account; make the Secret match Grafana's admin." >&2
  exit 1
fi

id=$(api GET "/api/serviceaccounts/search?perpage=1000&query=$SERVICE_ACCOUNT" |
  jq -r --arg n "$SERVICE_ACCOUNT" '[.serviceAccounts[] | select(.name == $n)][0].id // empty')

if [ "$DISABLED" = "true" ]; then
  if [ -n "$id" ]; then
    api PATCH "/api/serviceaccounts/$id" -d '{"isDisabled":true}' >/dev/null
    echo "disabled service account $SERVICE_ACCOUNT: Grafana refuses every token it holds"
  else
    echo "no service account $SERVICE_ACCOUNT exists; nothing to refuse"
  fi
  kubectl -n "$NAMESPACE" delete secret "$TOKEN_SECRET" --ignore-not-found >/dev/null
  echo "Secret $TOKEN_SECRET is gone"
  exit 0
fi

if [ -z "$id" ]; then
  id=$(api POST /api/serviceaccounts -d "{\"name\":\"$SERVICE_ACCOUNT\",\"role\":\"Viewer\"}" | jq -r .id)
  echo "created service account $SERVICE_ACCOUNT ($id) as a Viewer"
fi

account=$(api GET "/api/serviceaccounts/$id")
if [ "$(printf %s "$account" | jq -r .isDisabled)" = "true" ]; then
  # Enabling the account would make every token it still holds answer
  # again, so they are revoked first and a fresh one is minted below.
  for old in $(api GET "/api/serviceaccounts/$id/tokens" | jq -r '.[].id'); do
    api DELETE "/api/serviceaccounts/$id/tokens/$old" >/dev/null
    echo "revoked token $old before enabling the account again"
  done
  api PATCH "/api/serviceaccounts/$id" -d '{"isDisabled":false}' >/dev/null
  echo "enabled service account $SERVICE_ACCOUNT"
fi
role=$(printf %s "$account" | jq -r .role)
if [ "$role" != "Viewer" ]; then
  api PATCH "/api/serviceaccounts/$id" -d '{"role":"Viewer"}' >/dev/null
  echo "service account $SERVICE_ACCOUNT had the role $role; set back to Viewer"
fi

verb=create
if current=$(kubectl -n "$NAMESPACE" get secret "$TOKEN_SECRET" -o json 2>/dev/null); then
  verb=replace
  generation=$(printf %s "$current" | jq -r '.data.generation // "" | @base64d')
  token=$(printf %s "$current" | jq -r '.data.token // "" | @base64d')
  answer=$(curl -s -o /dev/null -w '%%{http_code}' -H "Authorization: Bearer $token" "$GRAFANA_URL/api/user")
  if [ "$generation" = "$TOKEN_GENERATION" ] && [ "$answer" = "200" ]; then
    echo "the generation $generation token in Secret $TOKEN_SECRET answers; nothing to do"
    exit 0
  fi
  echo "replacing the token in Secret $TOKEN_SECRET (it held generation $generation, and Grafana answered it $answer)"
fi

# Prove the token can be stored before minting it.
if ! kubectl -n "$NAMESPACE" auth can-i create secrets -q ||
  ! kubectl -n "$NAMESPACE" auth can-i update "secrets/$TOKEN_SECRET" -q; then
  echo "ServiceAccount $OWNER_SERVICE_ACCOUNT may not write Secret $TOKEN_SECRET in $NAMESPACE." >&2
  echo "Its Role $OWNER_SERVICE_ACCOUNT grants exactly that; re-apply the resource to restore it." >&2
  exit 1
fi
owner_uid=$(kubectl -n "$NAMESPACE" get serviceaccount "$OWNER_SERVICE_ACCOUNT" -o jsonpath='{.metadata.uid}')

minted=$(api POST "/api/serviceaccounts/$id/tokens" \
  -d "{\"name\":\"generation-$TOKEN_GENERATION-$(date -u +%Y%m%dT%H%M%SZ)\"}")
kept=$(printf %s "$minted" | jq -r .id)

# The Secret is owned by the module's ServiceAccount, so Kubernetes deletes
# it with the block or the resource; the token never enters deploy state.
# jq -j: the token's bytes exactly, with no trailing newline, because the
# Secret's value goes verbatim into an Authorization header.
printf %s "$minted" | jq -j .key |
  kubectl -n "$NAMESPACE" create secret generic "$TOKEN_SECRET" \
    --from-file=token=/dev/stdin --from-literal=generation="$TOKEN_GENERATION" \
    --dry-run=client -o json |
  jq --arg sa "$OWNER_SERVICE_ACCOUNT" --arg uid "$owner_uid" --argjson labels "$SECRET_LABELS" \
    '.metadata.labels = $labels | .metadata.ownerReferences = [{apiVersion: "v1", kind: "ServiceAccount", name: $sa, uid: $uid}]' |
  kubectl -n "$NAMESPACE" "$verb" -f - >/dev/null
echo "wrote the generation $TOKEN_GENERATION token to Secret $TOKEN_SECRET"

for old in $(api GET "/api/serviceaccounts/$id/tokens" | jq -r --argjson kept "$kept" '.[] | select(.id != $kept) | .id'); do
  api DELETE "/api/serviceaccounts/$id/tokens/$old" >/dev/null
  echo "revoked the account's older token $old"
done
EOT
}
