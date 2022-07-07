# sync-ignore-file: true
{
  local_domain: '.pdev.resf.localhost',
  default_domain: '.build.resf.org',
  service_mappings: {
    'peridotserver-http': {
      id: 'peridot-api',
      external: true,
    },
    'peridot-frontend-http': {
      id: 'peridot',
      external: true,
    },
    'yumrepofs-http': {
      id: 'yumrepofs',
      external: true,
    },
    'keykeeper-http': {
      id: 'keykeeper',
      external: false,
    },
    'keykeeper-grpc': {
      id: 'keykeeper-grpc',
      external: false,
    },
    'httpbin-http': {
      id: 'httpbin',
      external: true
    },
    'hydra-public-http': {
      id: 'hdr',
      external: true
    },
    'obsidian-http': {
      id: 'id-api',
      external: true,
    },
    'obsidian-frontend-http': {
      id: 'id',
      external: true,
    },
  }
}
