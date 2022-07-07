# publisher
(Peridot uses yumrepofs, this is for legacy errata/koji errata)

### Legacy errata mode
This mode only populates pungi generated repositories with errata metadata.
It can be deployed like this:
```
STABLE_STAGE=-prod bazel run --platforms @io_bazel_rules_go//go/toolchain:linux_amd64 //publisher/cmd/publisher-legacy-errata:publisher-legacy-errata-tool
```

After an updates compose is finished and merged into the correct point release directory (for example: 8.4-RC2) run:
```
ansible-playbook -i inventories/hosts.ini playbooks/secparse001-publish.yml
```

The ansible playbook is present in https://github.com/rocky-linux/peridot-ansible
