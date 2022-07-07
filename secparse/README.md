# secparse
Errata mirroring and publishing platform

### Testing
`bazel test --test_arg=-test.v --test_output=all $(bazel query 'tests(//secparse/...)')`

### Development
* Add `127.0.0.1 errata.pdot.localhost` to `/etc/hosts`
* Have a PostgreSQL database running with `postgres` user with `postgres` as password
* Create and migrate database `./hack/recreate_with_seed secparse`
* You can then run all components like this:
    ```
    bazel run //secparse/cmd/secparse
    bazel run //secparse/cmd/secparseadmin
    bazel run //secparse/cmd/secparsecron
    ibazel run //secparse/ui:secparse.server
    ```

You can then visit `http://errata.pdot.localhost:9007`

### Deployment (excluding `publisher`)
* Push all containers and tag with current git hash
```
STABLE_STAGE=-prod bazel run --platforms @io_bazel_rules_go//go/toolchain:linux_amd64 //secparse/cmd/secparse:secparse-server
STABLE_STAGE=-prod bazel run --platforms @io_bazel_rules_go//go/toolchain:linux_amd64 //secparse/cmd/secparseadmin:secparseadmin-server
STABLE_STAGE=-prod bazel run --platforms @io_bazel_rules_go//go/toolchain:linux_amd64 //secparse/cmd/secparsecron:secparsecron-server
STABLE_STAGE=-prod bazel run --platforms @build_bazel_rules_nodejs//toolchains/node:linux_amd64 //secparse/ui:secparse-frontend
```
* Clone `git@github.com:rocky-linux/peridot-ansible.git` and cd into `peridot-ansible`
* Change hashes in `roles/local/{name}/defaults/main.yml`
* First run migrate if the database schema has changed `ansible-playbook -i inventories/hosts.ini playbooks/secparse001-migrate.yml`
* Deploy containers `ansible-playbook -i inventories/hosts.ini playbooks/secparse001.yml`