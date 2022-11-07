local kubernetes = import 'ci/kubernetes.jsonnet';
local utils = import 'ci/utils.jsonnet';

{
  kube_env(prefix): [
    {
      name: '%s_S3_ENDPOINT' % prefix,
      value: if utils.helm_mode then '{{ .Values.s3Endpoint | default !"!" }}' else if utils.local_image then 'minio.default.svc.cluster.local:9000' else '',
    },
    {
      name: '%s_S3_DISABLE_SSL' % prefix,
      value: if utils.helm_mode then '{{ .Values.s3DisableSsl | default !"false!" | quote }}' else if utils.local_image then 'true' else 'false',
    },
    {
      name: '%s_S3_FORCE_PATH_STYLE' % prefix,
      value: if utils.helm_mode then '{{ .Values.s3ForcePathStyle | default !"false!" | quote }}' else if utils.local_image then 'true' else 'false',
    },
    {
      name: '%s_S3_REGION' % prefix,
      value: if utils.helm_mode then '{{ .Values.awsRegion | default !"us-east-2!" }}' else 'us-east-2',
    },
    {
      name: '%s_S3_BUCKET' % prefix,
      value: if utils.helm_mode then '{{ .Values.s3Bucket | default !"!" }}' else if kubernetes.prod() then 'resf-peridot-prod' else '',
    },
    {
      name: '%s_S3_ASSUME_ROLE' % prefix,
      value: if utils.helm_mode then '{{ .Values.s3AssumeRole | default !"!" }}' else '',
    },
  ],
}
