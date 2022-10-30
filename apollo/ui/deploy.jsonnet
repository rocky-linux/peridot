local bycdeploy = import 'ci/bycdeploy.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local frontend = import 'ci/frontend.jsonnet';

local tag = std.extVar('tag');

bycdeploy.new({
  name: 'apollo-frontend',
  backend: false,
  migrate: false,
  image: kubernetes.tag($.name),
  tag: tag,
  env: frontend.server_env,
  ports: [
    {
      name: 'http',
      containerPort: 8086,
      protocol: 'TCP',
      expose: true,
    },
  ],
  health: {
    port: 8086,
  },
})
