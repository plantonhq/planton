#!/usr/bin/env bash
set -euo pipefail

# Guard: a catalog kind that installs a Planton-owned chart installs the
# newest release of it.
#
# WHY THIS EXISTS
# KubernetesPlantonRunner and KubernetesPlantonOperator install charts this
# repository publishes, and an empty chart_version installs the version the
# modules pin. A chart release that fixes what a customer runs -- a build
# Role missing the verb the runner needs, a probe that restarts a healthy
# runner -- reaches the catalog only when someone moves that pin, and a pin
# nobody is reminded to move falls releases behind: every runner deployed
# from the catalog then carries defects its chart has already fixed. The two
# engines pin the version in two files, and the
# operator's schema default in a third; one moved without the others deploys
# different software from the same manifest depending on the engine.
#
# THE RULE
# For every kind in the table below, the Pulumi default (vars.go
# DefaultChartVersion), the Terraform default (locals.tf
# default_chart_version) and, where the kind's schema declares one, the
# proto default of spec.chart_version are equal, and equal the version of
# the chart's newest release tag. A chart release is followed by the pin
# moving in the same change that re-validates the values the modules render.
#
# WHERE IT RUNS
# lint.helm.yaml, on pull requests and pushes that touch either kind or this
# script, and on pushes to main that touch a chart. The chart's tag is cut
# after its merge, so red on main right after a chart release means the pin
# is owed a move; the sentence names the files that carry it.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

# kind | chart | release-tag prefix | the schema declares a chart_version default
kinds=(
  "kubernetesplantonrunner|planton-runner|helm/planton-runner/|no"
  "kubernetesplantonoperator|planton-operator|operator/|yes"
)

failures=()

for row in "${kinds[@]}"; do
  IFS='|' read -r kind chart tag_prefix has_proto_default <<<"${row}"
  kind_dir="catalog/kubernetes/${kind}"
  vars_go="${kind_dir}/iac/pulumi/module/vars.go"
  locals_tf="${kind_dir}/iac/tf/locals.tf"
  spec_proto="${kind_dir}/v1alpha1/spec.proto"

  tag="$(git tag --list "${tag_prefix}v*" --sort=-v:refname | head -n 1)"
  if [ -z "${tag}" ]; then
    failures+=("${kind} installs the ${chart} chart, but no ${tag_prefix}v* tag is visible -- fetch the tags (git fetch --tags) and re-run this check")
    continue
  fi
  newest="${tag#"${tag_prefix}"v}"

  pulumi_pin="$(sed -nE 's/^[[:space:]]*DefaultChartVersion:[[:space:]]*"([^"]*)".*/\1/p' "${vars_go}" | head -n 1)"
  tf_pin="$(sed -nE 's/^[[:space:]]*default_chart_version[[:space:]]*=[[:space:]]*"([^"]*)".*/\1/p' "${locals_tf}" | head -n 1)"

  pins="Pulumi pins ${pulumi_pin:-nothing} (${vars_go}), Terraform pins ${tf_pin:-nothing} (${locals_tf})"
  files="${vars_go} DefaultChartVersion, ${locals_tf} default_chart_version"
  stale="no"
  [ "${pulumi_pin}" != "${newest}" ] && stale="yes"
  [ "${tf_pin}" != "${newest}" ] && stale="yes"

  if [ "${has_proto_default}" = "yes" ]; then
    # The default option sits inside the chart_version field's option list,
    # which closes at the first "];" after the field line.
    proto_pin="$(awk '/string chart_version = /{f=1} f && /options\.default\)/{match($0, /"[^"]*"/); print substr($0, RSTART+1, RLENGTH-2); exit} f && /\];/{exit}' "${spec_proto}")"
    pins="${pins}, the schema defaults to ${proto_pin:-nothing} (${spec_proto})"
    files="${files}, ${spec_proto} chart_version (dev.planton.shared.options.default), then make protos"
    [ "${proto_pin}" != "${newest}" ] && stale="yes"
  fi

  if [ "${stale}" = "yes" ]; then
    failures+=("${kind}: ${pins}; the newest ${chart} release is ${newest} (${tag}) -- re-validate the values the modules render against ${newest}, then set it in ${files}")
  fi
done

if [ ${#failures[@]} -gt 0 ]; then
  echo "A catalog kind installs an older release of a Planton chart than the newest one:"
  for f in "${failures[@]}"; do
    echo "  - $f"
  done
  exit 1
fi

echo "OK: every catalog kind that installs a Planton chart pins its newest release in both engines."
