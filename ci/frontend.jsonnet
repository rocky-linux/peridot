{
  base: [
    {
      name: 'PORT',
      value: '8086',
    },
    {
      name: 'NODE_ENV',
      value: 'production',
    },
    {
      name: 'RESF_SECRET',
      valueFrom: true,
      secret: {
        name: 'server',
        key: 'byc-secret'
      },
    },
  ],
  server_env: $.base + [
    {
      name: 'HYDRA_SECRET',
      valueFrom: true,
      secret: {
        name: 'server',
        key: 'hydra-secret'
      },
    },
  ]
}
