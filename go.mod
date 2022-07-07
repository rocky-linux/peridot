//gen:comment:this file is generated with nofussvendor. DO NOT EDIT
module peridot.resf.org

go 1.15

require (
	alexejk.io/go-xmlrpc v0.2.0
	bazel.build/protobuf v0.0.0-00010101000000-000000000000
	cirello.io/dynamolock v1.4.0
	github.com/ProtonMail/go-crypto v0.0.0-20220113124808-70ae35bab23f
	github.com/ProtonMail/gopenpgp/v2 v2.4.7
	github.com/PuerkitoBio/goquery v1.7.0
	github.com/antchfx/xmlquery v1.3.6 // indirect
	github.com/authzed/authzed-go v0.3.0
	github.com/authzed/grpcutil v0.0.0-20211115181027-063820eb2511 // indirect
	github.com/aws/aws-sdk-go v1.36.12
	github.com/cavaliergopher/rpm v1.2.0
	github.com/coreos/go-oidc/v3 v3.0.0
	github.com/creack/pty v1.1.18 // indirect
	github.com/docker/docker v20.10.14+incompatible // indirect
	github.com/fatih/color v1.12.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-git/go-billy/v5 v5.3.1
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-openapi/runtime v0.19.31
	github.com/go-openapi/strfmt v0.20.2
	github.com/gobwas/glob v0.2.3
	github.com/gocolly/colly/v2 v2.1.0
	github.com/gogo/status v1.1.0
	github.com/google/uuid v1.3.0
	github.com/gosimple/slug v1.12.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.6.0
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jedib0t/go-pretty/v6 v6.2.7 // indirect
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.2
	github.com/mattn/go-tty v0.0.4 // indirect
	github.com/ory/hydra-client-go v1.10.6
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/rocky-linux/srpmproc v0.3.16
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/temoto/robotstxt v1.1.2 // indirect
	github.com/uber/jaeger-client-go v2.29.1+incompatible // indirect
	github.com/vbauerster/mpb/v7 v7.0.2
	github.com/vishvananda/netlink v1.1.0 // indirect
	github.com/xanzy/go-gitlab v0.50.4
	go.temporal.io/api v1.6.1-0.20211110205628-60c98e9cbfe2
	go.temporal.io/sdk v1.13.1
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	golang.org/x/sys v0.0.0-20211110154304-99a53858aa08
	google.golang.org/genproto v0.0.0-20211104193956-4c6863e31247
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.64 // indirect
	peridot.resf.org/common v0.0.0-00010101000000-000000000000
	peridot.resf.org/obsidian/pb v0.0.0-00010101000000-000000000000
	peridot.resf.org/peridot/keykeeper/pb v0.0.0-00010101000000-000000000000
	peridot.resf.org/peridot/pb v0.0.0-00010101000000-000000000000
	peridot.resf.org/peridot/yumrepofs/pb v0.0.0-00010101000000-000000000000
	peridot.resf.org/secparse/admin/proto v0.0.0-00010101000000-000000000000
	peridot.resf.org/secparse/proto v0.0.0-00010101000000-000000000000
)

// sync-replace-start
replace (
	bazel.build/protobuf => ./bazel-bin/build/bazel/protobuf/bazelbuild_go_proto_/bazel.build/protobuf
	bazel.build/remote/execution/v2 => ./bazel-bin/build/bazel/remote/execution/v2/remoteexecution_go_proto_/bazel.build/remote/execution/v2
	bazel.build/semver => ./bazel-bin/build/bazel/semver/semver_go_proto_/bazel.build/semver
	peridot.resf.org/obsidian/pb => ./bazel-bin/obsidian/proto/v1/obsidianpb_go_proto_/peridot.resf.org/obsidian/pb
	peridot.resf.org/peridot/pb => ./bazel-bin/peridot/proto/v1/peridotpb_go_proto_/peridot.resf.org/peridot/pb
	peridot.resf.org/peridot/keykeeper/pb => ./bazel-bin/peridot/proto/v1/keykeeper/keykeeperpb_go_proto_/peridot.resf.org/peridot/keykeeper/pb
	peridot.resf.org/peridot/yumrepofs/pb => ./bazel-bin/peridot/proto/v1/yumrepofs/yumrepofspb_go_proto_/peridot.resf.org/peridot/yumrepofs/pb
	peridot.resf.org/common => ./bazel-bin/proto/commonpb_go_proto_/peridot.resf.org/common
	peridot.resf.org/secparse/admin/proto => ./bazel-bin/secparse/admin/proto/v1/secparseadminpb_go_proto_/peridot.resf.org/secparse/admin/proto
	peridot.resf.org/secparse/proto => ./bazel-bin/secparse/proto/v1/secparsepb_go_proto_/peridot.resf.org/secparse/proto
	github.com/envoyproxy/protoc-gen-validate/validate => ./bazel-bin/vendor/github.com/envoyproxy/protoc-gen-validate/validate/go_default_library_/github.com/envoyproxy/protoc-gen-validate/validate
)
// sync-replace-end
