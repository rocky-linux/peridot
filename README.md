# Peridot
Named after the Gemstone,  Peridot (pronounced  - PERR-ih-dot) is a cloud-native build and release tool used for building, releasing and maintaining Linux distributions and forks.

## Structure
__Other components pending__

* publisher - `Composer for Peridot (currently only includes legacy mode)`
* peridot - `Modern build system`
* apollo - `Errata mirroring and publishing platform`
* utils - `Common utilities`
* modulemd - `Modulemd parser in Go`


## Development
Before the setup install `jq`, `golang`, `make`, `bazelisk`, `docker`, `helm`, and `kubectl`:

On Linux, jq, golang, make and docker can be installed using the package manager.

Links for installing the other software:
* Bazelisk: https://github.com/bazelbuild/bazelisk/releases
* Helm: https://helm.sh/docs/intro/install/
* Kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/

A local Kubernetes cluster is also required. Docker Desktop is a good solution.

Configure kubectl to manage the local Kubernetes cluster by placing the
KubeConfig yaml into `$HOME/.kube/config` and do `chmod 600 $HOME/.kube/config`

Then download istio into a local directory:
https://istio.io/latest/docs/setup/getting-started/

#### Initial setup (will soon be replaced by one command dev cluster)
```bash
# In the directory where you downloaded istio
bin/istioctl install --set profile=default --set hub=docker.io/querycapistio --set tag=1.12.1 -y
# On aarch64 (ex. M1 Mac) only and add arm64 to list of preferred schedule archs
# Run this while install is running
kubectl -n istio-system edit deployment istio-ingressgateway
sudo hack/deploy_dev_registry
hack/setup_external_dev_services
# Run `kubectl get svc` and add the port of postgres-postgresql to your rc file
# Example:
# postgres-postgresql          NodePort    10.102.68.75     <none>        5432:32442/TCP                  3m32s
# export POSTGRES_PORT="32442"
hack/setup_k8s_dev_env
git clone https://github.com/temporalio/temporal /tmp/temporal && pushd /tmp/temporal && make temporal-sql-tool && popd && hack/setup_dev_temporal /tmp/temporal
# Sometimes the namespace registration may fail because
# Temporal tools CrashLooped before we could run the migrations.
# Run `kubectl delete pods -l "app.kubernetes.io/name=temporal"` and then re-run
# `kubectl exec -it services/temporal-admintools -- tctl --namespace default namespace re`
hack/setup_base_internal_services
# For the cert, mkcert is recommended (mkcert.dev)
# Add default cert using `kubectl -n istio-system create secret tls default-cert --cert=cert.pem --key=cert.key`
# Create the Istio gateway
bazel run //infrastructure/istio-dev
```
Running `./hack/govendor` should create the necessary structure for development

For best experience use IntelliJ+Bazel but `govendor` creates structure that is compatible with all other Go tools
#### Auto generate (only) BUILD files for Go
`bazel run //:gazelle`
#### Vendor Go dependencies
`./hack/govendor`
#### Run UI in development mode
`ibazel run //TARGET:TARGET.server` - example: `ibazel run //apollo/ui:apollo.server`
#### Find UI server targets
`bazel query 'attr(tags, "resf_frontend_server", //...)'`

## Reporting Issues / Bugs

Before opening any issues in this GitHub repository, please take a moment to read the wiki page [Reporting Bugs and RFE's](https://wiki.rockylinux.org/rocky/bugs/)
