local base(info) = {
  apiVersion: 'pubsub.cnrm.cloud.google.com/v1beta1',
  kind: info.kind,
  metadata: info.metadata
} + if std.objectHas(info, 'spec') then info.spec else {};

{
  pubsub_topic(info)::
    local pubSubTopicInfo = {
      kind: 'PubSubTopic',
      metadata: info.metadata,
    };

    base(pubSubTopicInfo)
}
