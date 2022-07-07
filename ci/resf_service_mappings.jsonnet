# rename: ci/service_mappings.jsonnet
{
  local_domain: '.pdev.rocky.localhost',
  default_domain: '.build.rockylinux.org',
  service_mappings: {
    'peridotserver-http': {
      id: 'peridot',
      external: true,
    },
    'yumrepofs-http': {
      id: 'yumrepofs',
      external: true,
    },
    'httpbin-http': {
      id: 'httpbin',
      external: false
    }
  }
}