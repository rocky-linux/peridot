FROM quay.io/authzed/spicedb:v1.28.0 as spicedb
FROM quay.io/rockylinux/rockylinux:9

COPY --from=spicedb /usr/local/bin/spicedb /usr/local/bin/spicedb
COPY --from=spicedb /usr/local/bin/grpc_health_probe /usr/local/bin/grpc_health_probe
