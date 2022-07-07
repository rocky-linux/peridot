# peridot

### Local development
**Requirements:** Temporal (running on host), MinIO (running on host) and Kubernetes (Using Docker for Desktop)

Source the `.env` file: `source .env`

**Start UI:**
```
ibazel run //ui:peridot.server
```

**Start server (deploys to K8s):**
```
br //peridot/cmd/v1/peridotserver/ci:peridotserver.dev.local_push_apply
```

**Start ephemeral (deploys to K8s):**
```
br //peridot/cmd/v1/peridotephemeral/ci:peridotephemeral.dev.local_push_apply
```

To make changes to the builder or ephemeral instance just re-run the deployment command.
