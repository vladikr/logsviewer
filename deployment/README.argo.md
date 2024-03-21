FIXME

# Overview

[app.yaml](app.yaml) and [kustomization.yaml](kustomization.yaml)
can be used togetehr to self-service logsviewer.

# Deployment
# Admin preparation

1. Deploy OCP GitOps Operator on the cluster (by admin)

# User preparation

2. Create `argocd` CR in a fresh namespace
   Wait for argo to be deployed there

# User workflow for instance creation
3. For every case:     and/or create a branch of this repository
4. Adjust `kustomization.yaml` to point to your target namespace
5. Adjust `app.yaml` to point to your fork and your target namespace
6. `oc apply -f app.yaml`
