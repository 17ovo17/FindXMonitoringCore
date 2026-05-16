#!/usr/bin/env bash
set -euo pipefail

repo_root="${1:-}"

if [ -z "${repo_root}" ]; then
  echo "usage: $0 <package-repository-root>" >&2
  echo "refusing to write without an explicit repository root" >&2
  exit 2
fi

case "${repo_root}" in
  /|/opt|/opt/ai-workbench-runtime|/opt/ai-workbench-runtime/packages)
    echo "refusing unsafe repository root: ${repo_root}" >&2
    exit 2
    ;;
esac

require_artifact() {
  local ref="$1"
  case "${ref}" in
    /*|*../*|../*|..)
      echo "unsafe artifact ref: ${ref}" >&2
      exit 3
      ;;
  esac
  if [ ! -f "${repo_root}/${ref}" ]; then
    echo "required package artifact is missing: ${repo_root}/${ref}" >&2
    exit 4
  fi
}

write_ref() {
  local ref="$1"
  case "${ref}" in
    /*|*../*|../*|..)
      echo "unsafe ref: ${ref}" >&2
      exit 3
      ;;
  esac
  mkdir -p "${repo_root}/$(dirname "${ref}")"
  printf 'test-only evidence for %s\n' "${ref}" >"${repo_root}/${ref}"
}

write_tool_refs() {
  local os_name tool
  for os_name in linux windows; do
    for tool in signature_verifier checksum_verifier archive_extractor service_manager plugin_config_writer plugin_reloader; do
      write_ref "tools/${os_name}/${tool}.ref"
    done
  done
}

emit_tools() {
  local bundled="$1"
  local system="$2"
  local first=1
  local os_name tool
  for os_name in linux windows; do
    for tool in signature_verifier checksum_verifier archive_extractor service_manager plugin_config_writer plugin_reloader; do
      if [ "${first}" -eq 0 ]; then
        printf ',\n'
      fi
      first=0
      printf '    {"name":"%s","os":"%s","arch":"amd64","evidence_ref":"tools/%s/%s.ref","bundled":%s,"system":%s}' \
        "${tool}" "${os_name}" "${os_name}" "${tool}" "${bundled}" "${system}"
    done
  done
  printf '\n'
}

mkdir -p "${repo_root}/manifests"

for artifact_ref in \
  "bin/findx-categraf-linux-amd64" \
  "bin/findx-categraf-windows-amd64.exe" \
  "bin/findx-catpaw-linux-amd64" \
  "bin/findx-catpaw-windows-amd64.exe"; do
  require_artifact "${artifact_ref}"
done

for ref in \
  "checksums/SHA256SUMS" \
  "signatures/SHA256SUMS.asc" \
  "signatures/test-public-key.asc" \
  "signatures/test-key-fingerprint.txt" \
  "security/package-repository.ref" \
  "security/signature.ref" \
  "security/checksum.ref" \
  "manifests/script-manifest.json" \
  "executors/local-executor.ref" \
  "security/safety-policy.yaml" \
  "runners/local-runner.ref" \
  "security/ssh-host-key.ref" \
  "security/ssh-fingerprint.ref" \
  "services/findx-agent.service" \
  "installers/linux-curl.sh" \
  "config/findx-agent.env.tpl" \
  "installers/windows-installer.ps1" \
  "manifests/service.yaml" \
  "security/install-root-policy.yaml" \
  "manifests/rollback.yaml" \
  "manifests/uninstall.yaml" \
  "config/receiver-endpoint.ref" \
  "commands/install.sh" \
  "commands/uninstall.sh" \
  "commands/configure.sh" \
  "commands/plugins.sh" \
  "kubernetes/cluster.ref" \
  "kubernetes/namespace.ref" \
  "kubernetes/workload-selector.ref" \
  "kubernetes/helm/findx-agent-chart.ref" \
  "kubernetes/manifests/findx-agent-bundle.ref" \
  "kubernetes/helm/values.ref" \
  "kubernetes/rbac.ref" \
  "kubernetes/service-account.ref" \
  "kubernetes/images/findx-agent.ref" \
  "kubernetes/config-map.ref" \
  "kubernetes/secret.ref" \
  "kubernetes/rollout-strategy.ref" \
  "kubernetes/rollout-receipt.ref" \
  "validators/data-arrival.sh" \
  "audit/package-repository-audit.json" \
  "evidence/package-repository-evidence.json"; do
  write_ref "${ref}"
done

write_tool_refs

{
  cat <<'JSON'
{
  "repository": "findx-agent-runtime-local",
  "status": "checksum_ready_test_signature_ready",
  "signature_scope": "test-only-runtime-generated",
  "checksum_file": "checksums/SHA256SUMS",
  "checksum_signature": "signatures/SHA256SUMS.asc",
  "public_key": "signatures/test-public-key.asc",
  "public_key_fingerprint_file": "signatures/test-key-fingerprint.txt",
  "package_repository_ref": "security/package-repository.ref",
  "signature_ref": "security/signature.ref",
  "checksum": "security/checksum.ref",
  "script_manifest_ref": "manifests/script-manifest.json",
  "executor_ref": "executors/local-executor.ref",
  "safety_policy_path": "security/safety-policy.yaml",
  "runner": "runners/local-runner.ref",
  "ssh_host_key": "security/ssh-host-key.ref",
  "ssh_fingerprint": "security/ssh-fingerprint.ref",
  "systemd_unit_ref": "services/findx-agent.service",
  "curl_installer_ref": "installers/linux-curl.sh",
  "env_template_ref": "config/findx-agent.env.tpl",
  "windows_installer_ref": "installers/windows-installer.ps1",
  "service_manifest_ref": "manifests/service.yaml",
  "install_root_policy_ref": "security/install-root-policy.yaml",
  "rollback_manifest_ref": "manifests/rollback.yaml",
  "uninstall_manifest_ref": "manifests/uninstall.yaml",
  "receiver_endpoint_ref": "config/receiver-endpoint.ref",
  "install_command_ref": "commands/install.sh",
  "uninstall_command_ref": "commands/uninstall.sh",
  "config_command_ref": "commands/configure.sh",
  "plugin_command_ref": "commands/plugins.sh",
  "data_arrival_validator_ref": "validators/data-arrival.sh",
  "audit_ref": "audit/package-repository-audit.json",
  "evidence_chain_ref": "evidence/package-repository-evidence.json",
  "method": "linux-curl/windows-powershell/windows-cmd/kubernetes-helm/operator/daemonset/sidecar/initcontainer",
  "os": "linux/windows/kubernetes",
  "cluster_ref": "kubernetes/cluster.ref",
  "namespace_ref": "kubernetes/namespace.ref",
  "workload_selector_ref": "kubernetes/workload-selector.ref",
  "helm_chart_ref": "kubernetes/helm/findx-agent-chart.ref",
  "manifest_bundle_ref": "kubernetes/manifests/findx-agent-bundle.ref",
  "values_ref": "kubernetes/helm/values.ref",
  "rbac_ref": "kubernetes/rbac.ref",
  "service_account_ref": "kubernetes/service-account.ref",
  "image_ref": "kubernetes/images/findx-agent.ref",
  "config_map_ref": "kubernetes/config-map.ref",
  "secret_ref": "kubernetes/secret.ref",
  "rollout_strategy_ref": "kubernetes/rollout-strategy.ref",
  "rollout_receipt_ref": "kubernetes/rollout-receipt.ref",
  "artifacts": [
    {"id":"host-collector","artifact":"bin/findx-categraf-linux-amd64","os":"linux","arch":"amd64"},
    {"id":"host-collector","artifact":"bin/findx-categraf-windows-amd64.exe","os":"windows","arch":"amd64"},
    {"id":"inspection-runner","artifact":"bin/findx-catpaw-linux-amd64","os":"linux","arch":"amd64"},
    {"id":"inspection-runner","artifact":"bin/findx-catpaw-windows-amd64.exe","os":"windows","arch":"amd64"}
  ],
  "required_tools": [
JSON
  emit_tools false true
  cat <<'JSON'
  ],
  "bundled_tools": [
JSON
  emit_tools true false
  cat <<'JSON'
  ]
}
JSON
} >"${repo_root}/manifests/manifest.json"

python3 -m json.tool "${repo_root}/manifests/manifest.json" >/dev/null
echo "wrote test-only manifest to ${repo_root}/manifests/manifest.json"
