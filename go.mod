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
	github.com/authzed/grpcutil v0.0.0-20211115181027-063820eb2511
	github.com/aws/aws-sdk-go v1.44.129
	github.com/cavaliergopher/rpm v1.2.0
	github.com/coreos/go-oidc/v3 v3.0.0
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
	github.com/gorilla/feeds v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.6.0
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.2
	github.com/ory/hydra-client-go v1.10.6
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.13.0
	github.com/rocky-linux/srpmproc v0.4.3
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/temoto/robotstxt v1.1.2 // indirect
	github.com/vbauerster/mpb/v7 v7.0.2
	github.com/xanzy/go-gitlab v0.50.4
	go.temporal.io/api v1.6.1-0.20211110205628-60c98e9cbfe2
	go.temporal.io/sdk v1.13.1
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b
	golang.org/x/tools v0.1.6-0.20210726203631-07bc1bf47fb2 // indirect
	google.golang.org/genproto v0.0.0-20211104193956-4c6863e31247
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	openapi.peridot.resf.org/peridotopenapi v0.0.0-00010101000000-000000000000
	peridot.resf.org/apollo/pb v0.0.0-00010101000000-000000000000
	peridot.resf.org/common v0.0.0-00010101000000-000000000000
	peridot.resf.org/obsidian/pb v0.0.0-00010101000000-000000000000
	peridot.resf.org/peridot/keykeeper/pb v0.0.0-00010101000000-000000000000
	peridot.resf.org/peridot/pb v0.0.0-00010101000000-000000000000
	peridot.resf.org/peridot/yumrepofs/pb v0.0.0-00010101000000-000000000000
)

// Manual replace
replace (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible => github.com/golang-jwt/jwt/v4 v4.4.2
	openapi.peridot.resf.org/peridotopenapi => ./bazel-bin/peridot/proto/v1/client_go
)

// sync-replace-start
replace (
	peridot.resf.org/apollo/pb => ./bazel-bin/apollo/proto/v1/apollopb_go_proto_/peridot.resf.org/apollo/pb
	bazel.build/protobuf => ./bazel-bin/build/bazel/protobuf/bazelbuild_go_proto_/bazel.build/protobuf
	bazel.build/remote/execution/v2 => ./bazel-bin/build/bazel/remote/execution/v2/remoteexecution_go_proto_/bazel.build/remote/execution/v2
	bazel.build/semver => ./bazel-bin/build/bazel/semver/semver_go_proto_/bazel.build/semver
	peridot.resf.org/obsidian/pb => ./bazel-bin/obsidian/proto/v1/obsidianpb_go_proto_/peridot.resf.org/obsidian/pb
	peridot.resf.org/peridot/pb => ./bazel-bin/peridot/proto/v1/peridotpb_go_proto_/peridot.resf.org/peridot/pb
	peridot.resf.org/peridot/keykeeper/pb => ./bazel-bin/peridot/proto/v1/keykeeper/keykeeperpb_go_proto_/peridot.resf.org/peridot/keykeeper/pb
	peridot.resf.org/peridot/yumrepofs/pb => ./bazel-bin/peridot/proto/v1/yumrepofs/yumrepofspb_go_proto_/peridot.resf.org/peridot/yumrepofs/pb
	peridot.resf.org/common => ./bazel-bin/proto/commonpb_go_proto_/peridot.resf.org/common
	github.com/envoyproxy/protoc-gen-validate/validate => ./bazel-bin/vendor/github.com/envoyproxy/protoc-gen-validate/validate/go_default_library_/github.com/envoyproxy/protoc-gen-validate/validate
)
// sync-replace-end
