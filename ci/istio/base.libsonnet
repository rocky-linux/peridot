{
  KIND_ENVOYFILTER: 'EnvoyFilter',

  // Types (unfinished):
  //   - EnvoyFilter
  networking_base(info)::
    local envoyFilterTest = std.assertEqual(info.kind, $.KIND_ENVOYFILTER);

    local _ = if !envoyFilterTest then non_existing_value;

    {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: info.kind,
      metadata: info.metadata,
    } + if std.objectHas(info, 'spec') then info.spec else {},
}
