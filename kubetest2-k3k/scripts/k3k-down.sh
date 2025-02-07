#!/usr/bin/env bash
set -x 
## might be used as a fallback if the cluster endpoint isn't reachable or the cluster is not responding
## Check for required tools
command -v k3kcli > /dev/null || { echo "k3kcli is required but it's not installed.  Aborting." >&2; exit 1; }
command -v kubectl > /dev/null || { echo "kubectl is required but it's not installed.  Aborting." >&2; exit 1; }
## Default variables
KUBECONFIG="/etc/rancher/k3s/k3s.yaml:/etc/rancher/rke2/rke2.yaml"

## Check for required environment variables
[ -z "$K3K_CLUSTER_NAMESPACE" ] && { echo "K3K_CLUSTER_NAMESPACE is required but not set. Aborting." >&2; exit 1; }
[ -z "$K3K_CLUSTER_NAME" ] && { echo "K3K_CLUSTER_NAME is required but not set. Aborting." >&2; exit 1; }
