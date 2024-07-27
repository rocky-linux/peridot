load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_repositories():
    go_repository(
        name = "co_honnef_go_tools",
        importpath = "honnef.co/go/tools",
        sum = "h1:UoveltGrhghAA7ePc+e+QYDHXrBps2PqFZiHkGR/xK8=",
        version = "v0.0.1-2020.1.4",
    )
    go_repository(
        name = "com_github_acarl005_stripansi",
        importpath = "github.com/acarl005/stripansi",
        sum = "h1:licZJFw2RwpHMqeKTCYkitsPqHNxTmd4SNR5r94FGM8=",
        version = "v0.0.0-20180116102854-5a71ef0e047d",
    )
    go_repository(
        name = "com_github_alcortesm_tgz",
        importpath = "github.com/alcortesm/tgz",
        sum = "h1:uSoVVbwJiQipAclBbw+8quDsfcvFjOpI5iCf4p/cqCs=",
        version = "v0.0.0-20161220082320-9c5fe88206d7",
    )
    go_repository(
        name = "com_github_alecthomas_template",
        importpath = "github.com/alecthomas/template",
        sum = "h1:JYp7IbQjafoB+tBA3gMyHYHrpOtNuDiK/uB5uXxq5wM=",
        version = "v0.0.0-20190718012654-fb15b899a751",
    )
    go_repository(
        name = "com_github_alecthomas_units",
        importpath = "github.com/alecthomas/units",
        sum = "h1:UQZhZ2O0vMHr2cI+DC1Mbh0TJxzA3RcLoMsFw+aXw7E=",
        version = "v0.0.0-20190924025748-f65c72e2690d",
    )
    go_repository(
        name = "com_github_anmitsu_go_shlex",
        importpath = "github.com/anmitsu/go-shlex",
        sum = "h1:kFOfPq6dUM1hTo4JG6LR5AXSUEsOjtdm0kw0FtQtMJA=",
        version = "v0.0.0-20161002113705-648efa622239",
    )
    go_repository(
        name = "com_github_antihax_optional",
        importpath = "github.com/antihax/optional",
        sum = "h1:xK2lYat7ZLaVVcIuj82J8kIro4V6kDe0AUDFboUCwcg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_armon_circbuf",
        importpath = "github.com/armon/circbuf",
        sum = "h1:QEF07wC0T1rKkctt1RINW/+RMTVmiwxETico2l3gxJA=",
        version = "v0.0.0-20150827004946-bbbad097214e",
    )
    go_repository(
        name = "com_github_armon_go_metrics",
        importpath = "github.com/armon/go-metrics",
        sum = "h1:8GUt8eRujhVEGZFFEjBj46YV4rDjvGrNxb0KMWYkL2I=",
        version = "v0.0.0-20180917152333-f0300d1749da",
    )
    go_repository(
        name = "com_github_armon_go_radix",
        importpath = "github.com/armon/go-radix",
        sum = "h1:BUAU3CGlLvorLI26FmByPp2eC2qla6E1Tw+scpcg/to=",
        version = "v0.0.0-20180808171621-7fddfc383310",
    )
    go_repository(
        name = "com_github_armon_go_socks5",
        importpath = "github.com/armon/go-socks5",
        sum = "h1:0CwZNZbxp69SHPdPJAN/hZIm0C4OItdklCFmMRWYpio=",
        version = "v0.0.0-20160902184237-e75332964ef5",
    )
    go_repository(
        name = "com_github_asaskevich_govalidator",
        importpath = "github.com/asaskevich/govalidator",
        sum = "h1:idn718Q4B6AGu/h5Sxe66HYVdqdGu2l9Iebqhi/AEoA=",
        version = "v0.0.0-20190424111038-f61b66f89f4a",
    )
    go_repository(
        name = "com_github_authzed_authzed_go",
        importpath = "github.com/authzed/authzed-go",
        sum = "h1:m5eqPX9p1mhdbd8jrFhNORx5PvnHQ2e1bISweEwja+E=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_authzed_grpcutil",
        importpath = "github.com/authzed/grpcutil",
        sum = "h1:wbh8IK+aMLTCey9sZasO7b6BWLAJnHHvb79fvWCXwxw=",
        version = "v0.0.0-20240123194739-2ea1e3d2d98b",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go",
        importpath = "github.com/aws/aws-sdk-go",
        sum = "h1:yld8Rc8OCahLtenY1mnve4w1jVeBu/rSiscGzodaDOs=",
        version = "v1.44.129",
    )
    go_repository(
        name = "com_github_benbjohnson_clock",
        importpath = "github.com/benbjohnson/clock",
        sum = "h1:Q92kusRqC1XV2MjkWETPvjJVqKetz1OzxZB7mHJLju8=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_beorn7_perks",
        importpath = "github.com/beorn7/perks",
        sum = "h1:VlbKKnNfV8bJzeqoa4cOKqO6bYr3WgKZxO8Z16+hsOM=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_bgentry_speakeasy",
        importpath = "github.com/bgentry/speakeasy",
        sum = "h1:ByYyxL9InA1OWqxJqqp2A5pYHUrCiAL6K3J+LKSsQkY=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_bketelsen_crypt",
        importpath = "github.com/bketelsen/crypt",
        sum = "h1:+0HFd5KSZ/mm3JmhmrDukiId5iR6w4+BdFtfSy4yWIc=",
        version = "v0.0.3-0.20200106085610-5cbc8cc4026c",
    )
    go_repository(
        name = "com_github_bluekeyes_go_gitdiff",
        importpath = "github.com/bluekeyes/go-gitdiff",
        sum = "h1:AXrIoy/VEA9Baz2lhwMlpdzDJ/sKof6C9yTt1oqw4hQ=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_burntsushi_toml",
        importpath = "github.com/BurntSushi/toml",
        sum = "h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_burntsushi_xgb",
        importpath = "github.com/BurntSushi/xgb",
        sum = "h1:1BDTz0u9nC3//pOCMdNH+CiXJVYJh5UQNCOBG7jbELc=",
        version = "v0.0.0-20160522181843-27f122750802",
    )
    go_repository(
        name = "com_github_cavaliergopher_rpm",
        importpath = "github.com/cavaliergopher/rpm",
        sum = "h1:s0h+QeVK252QFTolkhGiMeQ1f+tMeIMhGl8B1HUmGUc=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_census_instrumentation_opencensus_proto",
        importpath = "github.com/census-instrumentation/opencensus-proto",
        sum = "h1:iKLQ0xPNFxR/2hzXZMrBo8f1j86j5WHzznCCQxV/b8g=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_certifi_gocertifi",
        importpath = "github.com/certifi/gocertifi",
        sum = "h1:S2NE3iHSwP0XV47EEXL8mWmRdEfGscSJ+7EgePNgt0s=",
        version = "v0.0.0-20210507211836-431795d63e8d",
    )
    go_repository(
        name = "com_github_cespare_xxhash",
        importpath = "github.com/cespare/xxhash",
        sum = "h1:a6HrQnmkObjyL+Gs60czilIUGqrzKutQD6XZog3p+ko=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_cespare_xxhash_v2",
        importpath = "github.com/cespare/xxhash/v2",
        sum = "h1:DC2CZ1Ep5Y4k3ZQ899DldepgrayRUGE6BBZ/cd9Cj44=",
        version = "v2.2.0",
    )
    go_repository(
        name = "com_github_chzyer_logex",
        importpath = "github.com/chzyer/logex",
        sum = "h1:Swpa1K6QvQznwJRcfTfQJmTE72DqScAa40E+fbHEXEE=",
        version = "v1.1.10",
    )
    go_repository(
        name = "com_github_chzyer_readline",
        importpath = "github.com/chzyer/readline",
        sum = "h1:lSwwFrbNviGePhkewF1az4oLmcwqCZijQ2/Wi3BGHAI=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_chzyer_test",
        importpath = "github.com/chzyer/test",
        sum = "h1:q763qf9huN11kDQavWsoZXJNW3xEE4JJyHa5Q25/sd8=",
        version = "v0.0.0-20180213035817-a1ea475d72b1",
    )
    go_repository(
        name = "com_github_client9_misspell",
        importpath = "github.com/client9/misspell",
        sum = "h1:ta993UF76GwbvJcIo3Y68y/M3WxlpEHPWIGDkJYwzJI=",
        version = "v0.3.4",
    )
    go_repository(
        name = "com_github_cncf_udpa_go",
        importpath = "github.com/cncf/udpa/go",
        sum = "h1:QQ3GSy+MqSHxm/d8nCtnAiZdYFd45cYZPs8vOOIYKfk=",
        version = "v0.0.0-20220112060539-c52dc94e7fbe",
    )
    go_repository(
        name = "com_github_cncf_xds_go",
        importpath = "github.com/cncf/xds/go",
        sum = "h1:jQCWAUqqlij9Pgj2i/PB79y4KOPYVyFYdROxgaCwdTQ=",
        version = "v0.0.0-20231128003011-0fa0005c9caa",
    )
    go_repository(
        name = "com_github_coreos_bbolt",
        importpath = "github.com/coreos/bbolt",
        sum = "h1:wZwiHHUieZCquLkDL0B8UhzreNWsPHooDAG3q34zk0s=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_coreos_etcd",
        importpath = "github.com/coreos/etcd",
        sum = "h1:8F3hqu9fGYLBifCmRCJsicFqDx/D68Rt3q1JMazcgBQ=",
        version = "v3.3.13+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_oidc_v3",
        importpath = "github.com/coreos/go-oidc/v3",
        sum = "h1:/mAA0XMgYJw2Uqm7WKGCsKnjitE/+A0FFbOmiRJm7LQ=",
        version = "v3.0.0",
    )
    go_repository(
        name = "com_github_coreos_go_semver",
        importpath = "github.com/coreos/go-semver",
        sum = "h1:wkHLiw0WNATZnSG7epLsujiMCgPAc9xhjJ4tgnAxmfM=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_coreos_go_systemd",
        importpath = "github.com/coreos/go-systemd",
        sum = "h1:Wf6HqHfScWJN9/ZjdUKyjop4mf3Qdd+1TvvltAvM3m8=",
        version = "v0.0.0-20190321100706-95778dfbb74e",
    )
    go_repository(
        name = "com_github_coreos_pkg",
        importpath = "github.com/coreos/pkg",
        sum = "h1:lBNOc5arjvs8E5mO2tbpBpLoyyu8B6e44T7hJy6potg=",
        version = "v0.0.0-20180928190104-399ea9e2e55f",
    )
    go_repository(
        name = "com_github_cpuguy83_go_md2man_v2",
        importpath = "github.com/cpuguy83/go-md2man/v2",
        sum = "h1:EoUDS0afbrsXAZ9YQ9jdu/mZ2sXgT1/2yyNng4PGlyM=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_creack_pty",
        importpath = "github.com/creack/pty",
        sum = "h1:uDmaGzcdjhF4i/plgjmEsriH11Y0o7RKapEf/LDaM3w=",
        version = "v1.1.9",
    )
    go_repository(
        name = "com_github_davecgh_go_spew",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_dgrijalva_jwt_go",
        importpath = "github.com/dgrijalva/jwt-go",
        replace = "github.com/golang-jwt/jwt/v4",
        sum = "h1:rcc4lwaZgFMCZ5jxF9ABolDcIHdBytAFgqFPbSJQAYs=",
        version = "v4.4.2",
    )
    go_repository(
        name = "com_github_dgryski_go_sip13",
        importpath = "github.com/dgryski/go-sip13",
        sum = "h1:RMLoZVzv4GliuWafOuPuQDKSm1SJph7uCRnnS61JAn4=",
        version = "v0.0.0-20181026042036-e10d5fee7954",
    )
    go_repository(
        name = "com_github_emicklei_go_restful_v3",
        importpath = "github.com/emicklei/go-restful/v3",
        sum = "h1:rAQeMHw1c7zTmncogyy8VvRZwtkmkZ4FxERmMY4rD+g=",
        version = "v3.11.0",
    )
    go_repository(
        name = "com_github_emirpasic_gods",
        importpath = "github.com/emirpasic/gods",
        sum = "h1:QAUIPSaCu4G+POclxeqb3F+WPpdKqFGlw36+yOzGlrg=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_github_envoyproxy_go_control_plane",
        importpath = "github.com/envoyproxy/go-control-plane",
        sum = "h1:4X+VP1GHd1Mhj6IB5mMeGbLCleqxjletLK6K0rbxyZI=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_github_envoyproxy_protoc_gen_validate",
        importpath = "github.com/envoyproxy/protoc-gen-validate",
        sum = "h1:gVPz/FMfvh57HdSJQyvBtF00j8JU4zdyUgIUNhlgg0A=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_evanphx_json_patch",
        importpath = "github.com/evanphx/json-patch",
        sum = "h1:4onqiflcdA9EOZ4RxV643DvftH5pOlLGNtQ5lPWQu84=",
        version = "v4.12.0+incompatible",
    )
    go_repository(
        name = "com_github_facebookgo_clock",
        importpath = "github.com/facebookgo/clock",
        sum = "h1:yDWHCSQ40h88yih2JAcL6Ls/kVkSE8GFACTGVnMPruw=",
        version = "v0.0.0-20150410010913-600d898af40a",
    )
    go_repository(
        name = "com_github_fatih_color",
        importpath = "github.com/fatih/color",
        sum = "h1:kOqh6YHBtK8aywxGerMG2Eq3H6Qgoqeo13Bk2Mv/nBs=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_github_felixge_httpsnoop",
        importpath = "github.com/felixge/httpsnoop",
        sum = "h1:NFTV2Zj1bL4mc9sqWACXbQFVBBg2W3GPvqp8/ESS2Wg=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_flynn_go_shlex",
        importpath = "github.com/flynn/go-shlex",
        sum = "h1:BHsljHzVlRcyQhjrss6TZTdY2VfCqZPbv5k3iBFa2ZQ=",
        version = "v0.0.0-20150515145356-3f9db97f8568",
    )
    go_repository(
        name = "com_github_fsnotify_fsnotify",
        importpath = "github.com/fsnotify/fsnotify",
        sum = "h1:hsms1Qyu0jgnwNXIxa+/V/PDsU6CfLf6CNO8H7IWoS4=",
        version = "v1.4.9",
    )
    go_repository(
        name = "com_github_ghodss_yaml",
        importpath = "github.com/ghodss/yaml",
        sum = "h1:wQHKEahhL6wmXdzwWG11gIVCkOv05bNOh+Rxn0yngAk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_gliderlabs_ssh",
        importpath = "github.com/gliderlabs/ssh",
        sum = "h1:6zsha5zo/TWhRhwqCD3+EarCAgZ2yN28ipRnGPnwkI0=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_go_chi_chi",
        importpath = "github.com/go-chi/chi",
        sum = "h1:fGFk2Gmi/YKXk0OmGfBh0WgmN3XB8lVnEyNz34tQRec=",
        version = "v4.1.2+incompatible",
    )
    go_repository(
        name = "com_github_go_git_gcfg",
        importpath = "github.com/go-git/gcfg",
        sum = "h1:Q5ViNfGF8zFgyJWPqYwA7qGFoMTEiBmdlkcfRmpIMa4=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_go_git_go_billy_v5",
        importpath = "github.com/go-git/go-billy/v5",
        sum = "h1:CPiOUAzKtMRvolEKw+bG1PLRpT7D3LIs3/3ey4Aiu34=",
        version = "v5.3.1",
    )
    go_repository(
        name = "com_github_go_git_go_git_fixtures_v4",
        importpath = "github.com/go-git/go-git-fixtures/v4",
        sum = "h1:PbKy9zOy4aAKrJ5pibIRpVO2BXnK1Tlcg+caKI7Ox5M=",
        version = "v4.0.2-0.20200613231340-f56387b50c12",
    )
    go_repository(
        name = "com_github_go_git_go_git_v5",
        importpath = "github.com/go-git/go-git/v5",
        sum = "h1:YPBLG/3UK1we1ohRkncLjaXWLW+HKp5QNM/jTli2JgI=",
        version = "v5.2.0",
    )
    go_repository(
        name = "com_github_go_gl_glfw",
        importpath = "github.com/go-gl/glfw",
        sum = "h1:QbL/5oDUmRBzO9/Z7Seo6zf912W/a6Sr4Eu0G/3Jho0=",
        version = "v0.0.0-20190409004039-e6da0acd62b1",
    )
    go_repository(
        name = "com_github_go_gl_glfw_v3_3_glfw",
        importpath = "github.com/go-gl/glfw/v3.3/glfw",
        sum = "h1:WtGNWLvXpe6ZudgnXrq0barxBImvnnJoMEhXAzcbM0I=",
        version = "v0.0.0-20200222043503-6f7a984d4dc4",
    )
    go_repository(
        name = "com_github_go_kit_kit",
        importpath = "github.com/go-kit/kit",
        sum = "h1:wDJmvq38kDhkVxi50ni9ykkdUr1PKgqKOoi01fa0Mdk=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_go_kit_log",
        importpath = "github.com/go-kit/log",
        sum = "h1:7i2K3eKTos3Vc0enKCfnVcgHh2olr/MyfboYq7cAcFw=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_go_logfmt_logfmt",
        importpath = "github.com/go-logfmt/logfmt",
        sum = "h1:otpy5pqBCBZ1ng9RQ0dPu4PN7ba75Y/aA+UpowDyNVA=",
        version = "v0.5.1",
    )
    go_repository(
        name = "com_github_go_logr_logr",
        importpath = "github.com/go-logr/logr",
        sum = "h1:pKouT5E8xu9zeFC39JXRDukb6JFQPXM5p5I91188VAQ=",
        version = "v1.4.1",
    )
    go_repository(
        name = "com_github_go_logr_stdr",
        importpath = "github.com/go-logr/stdr",
        sum = "h1:hSWxHoqTgW2S2qGc0LTAI563KZ5YKYRhT3MFKZMbjag=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_go_openapi_jsonpointer",
        importpath = "github.com/go-openapi/jsonpointer",
        sum = "h1:eCs3fxoIi3Wh6vtgmLTOjdhSpiqphQ+DaPn38N2ZdrE=",
        version = "v0.19.6",
    )
    go_repository(
        name = "com_github_go_openapi_jsonreference",
        importpath = "github.com/go-openapi/jsonreference",
        sum = "h1:3sVjiK66+uXK/6oQ8xgcRKcFgQ5KXa2KvnJRumpMGbE=",
        version = "v0.20.2",
    )
    go_repository(
        name = "com_github_go_openapi_swag",
        importpath = "github.com/go-openapi/swag",
        sum = "h1:yMBqmnQ0gyZvEb/+KzuWZOXgllrXT4SADYbvDaXHv/g=",
        version = "v0.22.3",
    )
    go_repository(
        name = "com_github_go_sql_driver_mysql",
        importpath = "github.com/go-sql-driver/mysql",
        sum = "h1:ozyZYNQW3x3HtqT1jira07DN2PArx2v7/mN66gGcHOs=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_go_stack_stack",
        importpath = "github.com/go-stack/stack",
        sum = "h1:5SgMzNM5HxrEjV0ww2lTmX6E2Izsfxas4+YHWRs3Lsk=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_github_go_task_slim_sprig",
        importpath = "github.com/go-task/slim-sprig",
        sum = "h1:tfuBGBXKqDEevZMzYi5KSi8KkcZtzBcTgAUUtapy0OI=",
        version = "v0.0.0-20230315185526-52ccab3ef572",
    )
    go_repository(
        name = "com_github_gobwas_glob",
        importpath = "github.com/gobwas/glob",
        sum = "h1:A4xDbljILXROh+kObIiy5kIaPYD8e96x1tgBhUI5J+Y=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_gogo_googleapis",
        importpath = "github.com/gogo/googleapis",
        sum = "h1:1Yx4Myt7BxzvUr5ldGSbwYiZG6t9wGBZ+8/fX3Wvtq0=",
        version = "v1.4.1",
    )
    go_repository(
        name = "com_github_gogo_protobuf",
        importpath = "github.com/gogo/protobuf",
        sum = "h1:Ov1cvc58UF3b5XjBnZv7+opcTcQFZebYjWzi34vdm4Q=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_gogo_status",
        importpath = "github.com/gogo/status",
        sum = "h1:+eIkrewn5q6b30y+g/BJINVVdi2xH7je5MPJ3ZPK3JA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_golang_glog",
        importpath = "github.com/golang/glog",
        sum = "h1:uCdmnmatrKCgMBlM4rMuJZWOkPDqdbZPnrMXDY4gI68=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_golang_groupcache",
        importpath = "github.com/golang/groupcache",
        sum = "h1:oI5xCqsCo564l8iNU+DwB5epxmsaqB+rhGL0m5jtYqE=",
        version = "v0.0.0-20210331224755-41bb18bfe9da",
    )
    go_repository(
        name = "com_github_golang_mock",
        importpath = "github.com/golang/mock",
        sum = "h1:ErTB+efbowRARo13NNdxyJji2egdxLGQhRaY+DUumQc=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_golang_protobuf",
        importpath = "github.com/golang/protobuf",
        sum = "h1:KhyjKVUg7Usr/dYsdSqoFveMYd5ko72D+zANwlG1mmg=",
        version = "v1.5.3",
    )
    go_repository(
        name = "com_github_golang_snappy",
        importpath = "github.com/golang/snappy",
        sum = "h1:yAGX7huGHXlcLOEtBnF4w7FQwA26wojNCwOYAEhLjQM=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_google_btree",
        importpath = "github.com/google/btree",
        sum = "h1:gK4Kx5IaGY9CD5sPJ36FHiBJ6ZXl0kilRiiCj+jdYp4=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_google_gnostic_models",
        importpath = "github.com/google/gnostic-models",
        sum = "h1:yo/ABAfM5IMRsS1VnXjTBvUb61tFIHozhlYvRgGre9I=",
        version = "v0.6.8",
    )
    go_repository(
        name = "com_github_google_go_cmp",
        importpath = "github.com/google/go-cmp",
        sum = "h1:ofyhxvXcZhMsU5ulbFiLKl/XBFqE1GSq7atu8tAmTRI=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_google_go_pkcs11",
        importpath = "github.com/google/go-pkcs11",
        sum = "h1:OF1IPgv+F4NmqmJ98KTjdN97Vs1JxDPB3vbmYzV2dpk=",
        version = "v0.2.1-0.20230907215043-c6f79328ddf9",
    )
    go_repository(
        name = "com_github_google_go_querystring",
        importpath = "github.com/google/go-querystring",
        sum = "h1:Xkwi/a1rcvNg1PPYe5vI8GbeBY/jrVuDX5ASuANWTrk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_google_gofuzz",
        importpath = "github.com/google/gofuzz",
        sum = "h1:xRy4A+RhZaiKjJ1bPfwQ8sedCA+YS2YcCHW6ec7JMi0=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_google_martian",
        importpath = "github.com/google/martian",
        sum = "h1:/CP5g8u/VJHijgedC/Legn3BAbAaWPgecwXBIDzw5no=",
        version = "v2.1.0+incompatible",
    )
    go_repository(
        name = "com_github_google_martian_v3",
        importpath = "github.com/google/martian/v3",
        sum = "h1:IqNFLAmvJOgVlpdEBiQbDc2EwKW77amAycfTuWKdfvw=",
        version = "v3.3.2",
    )
    go_repository(
        name = "com_github_google_pprof",
        importpath = "github.com/google/pprof",
        sum = "h1:Xim43kblpZXfIBQsbuBVKCudVG457BR2GZFIz3uw3hQ=",
        version = "v0.0.0-20221118152302-e6195bd50e26",
    )
    go_repository(
        name = "com_github_google_renameio",
        importpath = "github.com/google/renameio",
        sum = "h1:GOZbcHa3HfsPKPlmyPyN2KEohoMXOhdMbHrvbpl2QaA=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_s2a_go",
        importpath = "github.com/google/s2a-go",
        sum = "h1:60BLSyTrOV4/haCDW4zb1guZItoSq8foHCXrAnjBo/o=",
        version = "v0.1.7",
    )
    go_repository(
        name = "com_github_google_uuid",
        importpath = "github.com/google/uuid",
        sum = "h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_googleapis_enterprise_certificate_proxy",
        importpath = "github.com/googleapis/enterprise-certificate-proxy",
        sum = "h1:Vie5ybvEvT75RniqhfFxPRy3Bf7vr3h0cechB90XaQs=",
        version = "v0.3.2",
    )
    go_repository(
        name = "com_github_googleapis_gax_go_v2",
        build_directives = [
            "gazelle:resolve go google.golang.org/genproto/googleapis/rpc/code @org_golang_google_genproto_googleapis_rpc//code",
        ],
        build_file_proto_mode = "disable",
        importpath = "github.com/googleapis/gax-go/v2",
        sum = "h1:9F8GV9r9ztXyAi00gsMQHNoF51xPZm8uj1dpYt2ZETM=",
        version = "v2.12.1",
    )
    go_repository(
        name = "com_github_googleapis_google_cloud_go_testing",
        importpath = "github.com/googleapis/google-cloud-go-testing",
        sum = "h1:tlyzajkF3030q6M8SvmJSemC9DTHL/xaMa18b65+JM4=",
        version = "v0.0.0-20200911160855-bcd43fbb19e8",
    )
    go_repository(
        name = "com_github_gopherjs_gopherjs",
        importpath = "github.com/gopherjs/gopherjs",
        sum = "h1:EGx4pi6eqNxGaHF6qqu48+N2wcFQ5qg5FXgOdqsJ5d8=",
        version = "v0.0.0-20181017120253-0766667cb4d1",
    )
    go_repository(
        name = "com_github_gorilla_websocket",
        importpath = "github.com/gorilla/websocket",
        sum = "h1:PPwGk2jz7EePpoHN/+ClbZu8SPxiqlu12wZP/3sWmnc=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_gregjones_httpcache",
        importpath = "github.com/gregjones/httpcache",
        sum = "h1:pdN6V1QBWetyv/0+wjACpqVH+eVULgEjkurDLq3goeM=",
        version = "v0.0.0-20180305231024-9cad4c3443a7",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_middleware",
        importpath = "github.com/grpc-ecosystem/go-grpc-middleware",
        sum = "h1:UH//fgunKIs4JdUbpDl1VZCDaL56wXCB/5+wF6uHfaI=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_prometheus",
        importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
        sum = "h1:Ovs26xHkKqVztRpIrF/92BcuyuQ/YW4NSIpoGtfXNho=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_grpc_gateway",
        importpath = "github.com/grpc-ecosystem/grpc-gateway",
        sum = "h1:gmcG1KaJ57LophUzW0Hy8NmPhnMZb4M0+kPpLofRdBo=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_grpc_gateway_v2",
        importpath = "github.com/grpc-ecosystem/grpc-gateway/v2",
        sum = "h1:/c3QmbOGMGTOumP2iT/rCwB7b0QDGLKzqOmktBjT+Is=",
        version = "v2.19.1",
    )
    go_repository(
        name = "com_github_hashicorp_consul_api",
        importpath = "github.com/hashicorp/consul/api",
        sum = "h1:BNQPM9ytxj6jbjjdRPioQ94T6YXriSopn0i8COv6SRA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_hashicorp_consul_sdk",
        importpath = "github.com/hashicorp/consul/sdk",
        sum = "h1:LnuDWGNsoajlhGyHJvuWW6FVqRl8JOTPqS6CPTsYjhY=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_hashicorp_errwrap",
        importpath = "github.com/hashicorp/errwrap",
        sum = "h1:hLrqtEDnRye3+sgx6z4qVLNuviH3MR5aQ0ykNJa/UYA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_cleanhttp",
        importpath = "github.com/hashicorp/go-cleanhttp",
        sum = "h1:dH3aiDG9Jvb5r5+bYHsikaOUIpcM0xvgMXVoDkXMzJM=",
        version = "v0.5.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_hclog",
        importpath = "github.com/hashicorp/go-hclog",
        sum = "h1:CG6TE5H9/JXsFWJCfoIVpKFIkFe6ysEuHirp4DxCsHI=",
        version = "v0.9.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_immutable_radix",
        importpath = "github.com/hashicorp/go-immutable-radix",
        sum = "h1:AKDB1HM5PWEA7i4nhcpwOrO2byshxBjXVn/J/3+z5/0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_msgpack",
        importpath = "github.com/hashicorp/go-msgpack",
        sum = "h1:zKjpN5BK/P5lMYrLmBHdBULWbJ0XpYR+7NGzqkZzoD4=",
        version = "v0.5.3",
    )
    go_repository(
        name = "com_github_hashicorp_go_multierror",
        importpath = "github.com/hashicorp/go-multierror",
        sum = "h1:iVjPR7a6H0tWELX5NxNe7bYopibicUzc7uPribsnS6o=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_net",
        importpath = "github.com/hashicorp/go.net",
        sum = "h1:sNCoNyDEvN1xa+X0baata4RdcpKwcMS6DH+xwfqPgjw=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_retryablehttp",
        importpath = "github.com/hashicorp/go-retryablehttp",
        sum = "h1:92lWxgpa+fF3FozM4B3UZtHZMJX8T5XT+TFdCxsPyWs=",
        version = "v0.6.8",
    )
    go_repository(
        name = "com_github_hashicorp_go_rootcerts",
        importpath = "github.com/hashicorp/go-rootcerts",
        sum = "h1:Rqb66Oo1X/eSV1x66xbDccZjhJigjg0+e82kpwzSwCI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_sockaddr",
        importpath = "github.com/hashicorp/go-sockaddr",
        sum = "h1:GeH6tui99pF4NJgfnhp+L6+FfobzVW3Ah46sLo0ICXs=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_syslog",
        importpath = "github.com/hashicorp/go-syslog",
        sum = "h1:KaodqZuhUoZereWVIYmpUgZysurB1kBLX2j0MwMrUAE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_uuid",
        importpath = "github.com/hashicorp/go-uuid",
        sum = "h1:fv1ep09latC32wFoVwnqcnKJGnMSdBanPczbHAYm1BE=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_hashicorp_golang_lru",
        importpath = "github.com/hashicorp/golang-lru",
        sum = "h1:0hERBMJE1eitiLkihrMvRVBYAkpHzc/J3QdDN+dAcgU=",
        version = "v0.5.1",
    )
    go_repository(
        name = "com_github_hashicorp_hcl",
        importpath = "github.com/hashicorp/hcl",
        sum = "h1:0Anlzjpi4vEasTeNFn2mLJgTSwt0+6sfsiTG8qcWGx4=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_logutils",
        importpath = "github.com/hashicorp/logutils",
        sum = "h1:dLEQVugN8vlakKOUE3ihGLTZJRB4j+M2cdTm/ORI65Y=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_mdns",
        importpath = "github.com/hashicorp/mdns",
        sum = "h1:WhIgCr5a7AaVH6jPUwjtRuuE7/RDufnUvzIr48smyxs=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_memberlist",
        importpath = "github.com/hashicorp/memberlist",
        sum = "h1:EmmoJme1matNzb+hMpDuR/0sbJSUisxyqBGG676r31M=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_hashicorp_serf",
        importpath = "github.com/hashicorp/serf",
        sum = "h1:YZ7UKsJv+hKjqGVUUbtE3HNj79Eln2oQ75tniF6iPt0=",
        version = "v0.8.2",
    )
    go_repository(
        name = "com_github_iancoleman_strcase",
        importpath = "github.com/iancoleman/strcase",
        sum = "h1:nTXanmYxhfFAMjZL34Ov6gkzEsSJZ5DbhxWjvSASxEI=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_ianlancetaylor_demangle",
        importpath = "github.com/ianlancetaylor/demangle",
        sum = "h1:rcanfLhLDA8nozr/K289V1zcntHr3V+SHlXwzz1ZI2g=",
        version = "v0.0.0-20220319035150-800ac71e25c2",
    )
    go_repository(
        name = "com_github_imdario_mergo",
        importpath = "github.com/imdario/mergo",
        sum = "h1:3tnifQM4i+fbajXKBHXWEH+KvNHqojZ778UH75j3bGA=",
        version = "v0.3.11",
    )
    go_repository(
        name = "com_github_inconshreveable_mousetrap",
        importpath = "github.com/inconshreveable/mousetrap",
        sum = "h1:Z8tu5sraLXCXIcARxBp/8cbvlwVa7Z1NHg9XEKhtSvM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jbenet_go_context",
        importpath = "github.com/jbenet/go-context",
        sum = "h1:BQSFePA1RWJOlocH6Fxy8MmwDt+yVQYULKfN0RoTN8A=",
        version = "v0.0.0-20150711004518-d14ea06fba99",
    )
    go_repository(
        name = "com_github_jessevdk_go_flags",
        importpath = "github.com/jessevdk/go-flags",
        sum = "h1:4IU2WS7AumrZ/40jfhf4QVDMsQwqA7VEHozFRrGARJA=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_jmespath_go_jmespath",
        importpath = "github.com/jmespath/go-jmespath",
        sum = "h1:BEgLn5cpjn8UN1mAw4NjwDrS35OdebyEtFe+9YPoQUg=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_jmespath_go_jmespath_internal_testify",
        importpath = "github.com/jmespath/go-jmespath/internal/testify",
        sum = "h1:shLQSRRSCCPj3f2gpwzGwWFoC7ycTf1rcQZHOlsJ6N8=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_jmoiron_sqlx",
        importpath = "github.com/jmoiron/sqlx",
        sum = "h1:wv+0IJZfL5z0uZoUjlpKgHkgaFSYD+r9CfrXjEXsO7w=",
        version = "v1.3.4",
    )
    go_repository(
        name = "com_github_jonboulle_clockwork",
        importpath = "github.com/jonboulle/clockwork",
        sum = "h1:VKV+ZcuP6l3yW9doeqz6ziZGgcynBVQO+obU0+0hcPo=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_josharian_intern",
        importpath = "github.com/josharian/intern",
        sum = "h1:vlS4z54oSdjm0bgjRigI+G1HpF+tI+9rE5LLzOg8HmY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jpillora_backoff",
        importpath = "github.com/jpillora/backoff",
        sum = "h1:uvFg412JmmHBHw7iwprIxkPMI+sGQ4kzOWsMeHnm2EA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_json_iterator_go",
        importpath = "github.com/json-iterator/go",
        sum = "h1:PV8peI4a0ysnczrg+LtxykD8LfKY9ML6u2jnxaEnrnM=",
        version = "v1.1.12",
    )
    go_repository(
        name = "com_github_jstemmer_go_junit_report",
        importpath = "github.com/jstemmer/go-junit-report",
        sum = "h1:6QPYqodiu3GuPL+7mfx+NwDdp2eTkp9IfEUpgAwUN0o=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_jtolds_gls",
        importpath = "github.com/jtolds/gls",
        sum = "h1:xdiiI2gbIgH/gLH7ADydsJ1uDOEzR8yvV7C0MuV77Wo=",
        version = "v4.20.0+incompatible",
    )
    go_repository(
        name = "com_github_julienschmidt_httprouter",
        importpath = "github.com/julienschmidt/httprouter",
        sum = "h1:U0609e9tgbseu3rBINet9P48AI/D3oJs4dN7jwJOQ1U=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_jzelinskie_stringz",
        importpath = "github.com/jzelinskie/stringz",
        sum = "h1:thTca5Eyouk5CEcJ75Cbw9CSAGE7TAc6rIi+WgHWpOE=",
        version = "v0.0.0-20210414224931-d6a8ce844a70",
    )
    go_repository(
        name = "com_github_kevinburke_ssh_config",
        importpath = "github.com/kevinburke/ssh_config",
        sum = "h1:Coekwdh0v2wtGp9Gmz1Ze3eVRAWJMLokvN3QjdzCHLY=",
        version = "v0.0.0-20190725054713-01f96b0aa0cd",
    )
    go_repository(
        name = "com_github_kisielk_errcheck",
        importpath = "github.com/kisielk/errcheck",
        sum = "h1:e8esj/e4R+SAOwFwN+n3zr0nYeCyeweozKfO23MvHzY=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_kisielk_gotool",
        importpath = "github.com/kisielk/gotool",
        sum = "h1:AV2c/EiW3KqPNT9ZKl07ehoAGi4C5/01Cfbblndcapg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_konsorten_go_windows_terminal_sequences",
        importpath = "github.com/konsorten/go-windows-terminal-sequences",
        sum = "h1:CE8S1cTafDpPvMhIxNJKvHsGVBgn1xWYf1NbHQhywc8=",
        version = "v1.0.3",
    )
    go_repository(
        name = "com_github_kr_fs",
        importpath = "github.com/kr/fs",
        sum = "h1:Jskdu9ieNAYnjxsi0LbQp1ulIKZV1LAFgK1tWhpZgl8=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_kr_logfmt",
        importpath = "github.com/kr/logfmt",
        sum = "h1:T+h1c/A9Gawja4Y9mFVWj2vyii2bbUNDw3kt9VxK2EY=",
        version = "v0.0.0-20140226030751-b84e30acd515",
    )
    go_repository(
        name = "com_github_kr_pretty",
        importpath = "github.com/kr/pretty",
        sum = "h1:flRD4NNwYAUpkphVc1HcthR4KEIFJ65n8Mw5qdRn3LE=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_kr_pty",
        importpath = "github.com/kr/pty",
        sum = "h1:VkoXIwSboBpnk99O/KFauAEILuNHv5DVFKZMBN/gUgw=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_kr_text",
        importpath = "github.com/kr/text",
        sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_lib_pq",
        importpath = "github.com/lib/pq",
        sum = "h1:AqzbZs4ZoCBp+GtejcpCpcxM3zlSMx29dXbUSeVtJb8=",
        version = "v1.10.2",
    )
    go_repository(
        name = "com_github_lyft_protoc_gen_star",
        importpath = "github.com/lyft/protoc-gen-star",
        sum = "h1:sImehRT+p7lW9n6R7MQc5hVgzWGEkDVZU4AsBQ4Isu8=",
        version = "v0.5.1",
    )
    go_repository(
        name = "com_github_lyft_protoc_gen_star_v2",
        importpath = "github.com/lyft/protoc-gen-star/v2",
        sum = "h1:/3+/2sWyXeMLzKd1bX+ixWKgEMsULrIivpDsuaF441o=",
        version = "v2.0.3",
    )
    go_repository(
        name = "com_github_magiconair_properties",
        importpath = "github.com/magiconair/properties",
        sum = "h1:ZC2Vc7/ZFkGmsVC9KvOjumD+G5lXy2RtTKyzRKO2BQ4=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_mailru_easyjson",
        importpath = "github.com/mailru/easyjson",
        sum = "h1:UGYAvKxe3sBsEDzO8ZeWOSlIQfWFlxbzLZe7hwFURr0=",
        version = "v0.7.7",
    )
    go_repository(
        name = "com_github_mattn_go_colorable",
        importpath = "github.com/mattn/go-colorable",
        sum = "h1:fFA4WZxdEF4tXPZVKMLwD8oUnCTTo08duU7wxecdEvA=",
        version = "v0.1.13",
    )
    go_repository(
        name = "com_github_mattn_go_isatty",
        importpath = "github.com/mattn/go-isatty",
        sum = "h1:JITubQf0MOLdlGRuRq+jtsDlekdYPia9ZFsB8h/APPA=",
        version = "v0.0.19",
    )
    go_repository(
        name = "com_github_mattn_go_runewidth",
        importpath = "github.com/mattn/go-runewidth",
        sum = "h1:lTGmDsbAYt5DmK6OnoV7EuIF1wEIFAcxld6ypU4OSgU=",
        version = "v0.0.13",
    )
    go_repository(
        name = "com_github_mattn_go_sqlite3",
        importpath = "github.com/mattn/go-sqlite3",
        sum = "h1:yOQRA0RpS5PFz/oikGwBEqvAWhWg5ufRz4ETLjwpU1Y=",
        version = "v1.14.16",
    )
    go_repository(
        name = "com_github_matttproud_golang_protobuf_extensions",
        importpath = "github.com/matttproud/golang_protobuf_extensions",
        sum = "h1:4hp9jkHxhMHkqkrB3Ix0jegS5sx/RkqARlsWZ6pIwiU=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_miekg_dns",
        importpath = "github.com/miekg/dns",
        sum = "h1:9jZdLNd/P4+SfEJ0TNyxYpsK8N4GtfylBLqtbYN1sbA=",
        version = "v1.0.14",
    )
    go_repository(
        name = "com_github_mitchellh_cli",
        importpath = "github.com/mitchellh/cli",
        sum = "h1:iGBIsUe3+HZ/AD/Vd7DErOt5sU9fa8Uj7A2s1aggv1Y=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_go_homedir",
        importpath = "github.com/mitchellh/go-homedir",
        sum = "h1:lukF9ziXFxDFPkA1vsr5zpc1XuPDn/wFntq5mG+4E0Y=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_mitchellh_go_testing_interface",
        importpath = "github.com/mitchellh/go-testing-interface",
        sum = "h1:fzU/JVNcaqHQEcVFAKeR41fkiLdIPrefOvVG1VZ96U0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_gox",
        importpath = "github.com/mitchellh/gox",
        sum = "h1:lfGJxY7ToLJQjHHwi0EX6uYBdK78egf954SQl13PQJc=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_mitchellh_iochan",
        importpath = "github.com/mitchellh/iochan",
        sum = "h1:C+X3KsSTLFVBr/tK1eYN/vs4rJcvsiLU338UhYPJWeY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_mapstructure",
        importpath = "github.com/mitchellh/mapstructure",
        sum = "h1:CpVNEelQCZBooIPDn+AR3NpivK/TIKU8bDxdASFVQag=",
        version = "v1.4.1",
    )
    go_repository(
        name = "com_github_moby_spdystream",
        importpath = "github.com/moby/spdystream",
        sum = "h1:cjW1zVyyoiM0T7b6UoySUFqzXMoqRckQtXwGPiBhOM8=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_modern_go_concurrent",
        importpath = "github.com/modern-go/concurrent",
        sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
        version = "v0.0.0-20180306012644-bacd9c7ef1dd",
    )
    go_repository(
        name = "com_github_modern_go_reflect2",
        importpath = "github.com/modern-go/reflect2",
        sum = "h1:xBagoLtFs94CBntxluKeaWgTMpvLxC4ur3nMaC9Gz0M=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_munnerz_goautoneg",
        importpath = "github.com/munnerz/goautoneg",
        sum = "h1:C3w9PqII01/Oq1c1nUAm88MOHcQC9l5mIlSMApZMrHA=",
        version = "v0.0.0-20191010083416-a7dc8b61c822",
    )
    go_repository(
        name = "com_github_mwitkow_go_conntrack",
        importpath = "github.com/mwitkow/go-conntrack",
        sum = "h1:KUppIJq7/+SVif2QVs3tOP0zanoHgBEVAwHxUSIzRqU=",
        version = "v0.0.0-20190716064945-2f068394615f",
    )
    go_repository(
        name = "com_github_mxk_go_flowrate",
        importpath = "github.com/mxk/go-flowrate",
        sum = "h1:y5//uYreIhSUg3J1GEMiLbxo1LJaP8RfCpH6pymGZus=",
        version = "v0.0.0-20140419014527-cca7078d478f",
    )
    go_repository(
        name = "com_github_niemeyer_pretty",
        importpath = "github.com/niemeyer/pretty",
        sum = "h1:fD57ERR4JtEqsWbfPhv4DMiApHyliiK5xCTNVSPiaAs=",
        version = "v0.0.0-20200227124842-a10e7caefd8e",
    )
    go_repository(
        name = "com_github_nytimes_gziphandler",
        importpath = "github.com/NYTimes/gziphandler",
        sum = "h1:lsxEuwrXEAokXB9qhlbKWPpo3KMLZQ5WB5WLQRW1uq0=",
        version = "v0.0.0-20170623195520-56545f4a5d46",
    )
    go_repository(
        name = "com_github_oklog_ulid",
        importpath = "github.com/oklog/ulid",
        sum = "h1:EGfNDEx6MqHz8B3uNV6QAib1UR2Lm97sHi3ocA6ESJ4=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_oneofone_xxhash",
        importpath = "github.com/OneOfOne/xxhash",
        sum = "h1:KMrpdQIwFcEqXDklaen+P1axHaj9BSKzvpUUfnHldSE=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_onsi_ginkgo_v2",
        importpath = "github.com/onsi/ginkgo/v2",
        sum = "h1:0jY9lJquiL8fcf3M4LAXN5aMlS/b2BV86HFFPCPMgE4=",
        version = "v2.13.0",
    )
    go_repository(
        name = "com_github_onsi_gomega",
        importpath = "github.com/onsi/gomega",
        sum = "h1:KIA/t2t5UBzoirT4H9tsML45GEbo3ouUnBHsCfD2tVg=",
        version = "v1.29.0",
    )
    go_repository(
        name = "com_github_opentracing_opentracing_go",
        importpath = "github.com/opentracing/opentracing-go",
        sum = "h1:pWlfV3Bxv7k65HYwkikxat0+s3pV4bsqf19k25Ur8rU=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_ory_hydra_client_go",
        importpath = "github.com/ory/hydra-client-go",
        sum = "h1:GwJjvH/DBcfYzoST4vUpi4pIRzDGH5oODKpIVuhwVyc=",
        version = "v1.11.8",
    )
    go_repository(
        name = "com_github_ory_hydra_client_go_v2",
        importpath = "github.com/ory/hydra-client-go/v2",
        sum = "h1:g8hw0YQD5Us1aAgZj7OyBmBGSDwlnY9/2Pb/pQQq8YE=",
        version = "v2.2.0",
    )
    go_repository(
        name = "com_github_pascaldekloe_goe",
        importpath = "github.com/pascaldekloe/goe",
        sum = "h1:Lgl0gzECD8GnQ5QCWA8o6BtfL6mDH5rQgM4/fX3avOs=",
        version = "v0.0.0-20180627143212-57f6aae5913c",
    )
    go_repository(
        name = "com_github_pborman_uuid",
        importpath = "github.com/pborman/uuid",
        sum = "h1:+ZZIw58t/ozdjRaXh/3awHfmWRbzYxJoAdNJxe/3pvw=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_pelletier_go_toml",
        importpath = "github.com/pelletier/go-toml",
        sum = "h1:1Nf83orprkJyknT6h7zbuEGUEjcyVlCxSUGTENmNCRM=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_peterbourgon_diskv",
        importpath = "github.com/peterbourgon/diskv",
        sum = "h1:UBdAOUP5p4RWqPBg048CAvpKN+vxiaj6gdUUzhl4XmI=",
        version = "v2.0.1+incompatible",
    )
    go_repository(
        name = "com_github_pkg_errors",
        importpath = "github.com/pkg/errors",
        sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_pkg_sftp",
        importpath = "github.com/pkg/sftp",
        sum = "h1:I2qBYMChEhIjOgazfJmV3/mZM256btk6wkCDRmW7JYs=",
        version = "v1.13.1",
    )
    go_repository(
        name = "com_github_pmezard_go_difflib",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_posener_complete",
        importpath = "github.com/posener/complete",
        sum = "h1:ccV59UEOTzVDnDUEFdT95ZzHVZ+5+158q8+SJb2QV5w=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_prometheus_client_golang",
        importpath = "github.com/prometheus/client_golang",
        sum = "h1:b71QUfeo5M8gq2+evJdTPfZhYMAU0uKPkyPJ7TPsloU=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_github_prometheus_client_model",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:VQw1hfvPvk3Uv6Qf29VrPF32JB6rtbgI6cYPYQjL0Qw=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_prometheus_common",
        importpath = "github.com/prometheus/common",
        sum = "h1:ccBbHCgIiT9uSoFY0vX8H3zsNR5eLt17/RQLUvn8pXE=",
        version = "v0.37.0",
    )
    go_repository(
        name = "com_github_prometheus_procfs",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:ODq8ZFEaYeCaZOJlZZdJA2AbQR98dSHSM1KW/You5mo=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_github_prometheus_tsdb",
        importpath = "github.com/prometheus/tsdb",
        sum = "h1:YZcsG11NqnK4czYLrWd9mpEuAJIHVQLwdrleYfszMAA=",
        version = "v0.7.1",
    )
    go_repository(
        name = "com_github_protonmail_go_crypto",
        importpath = "github.com/ProtonMail/go-crypto",
        sum = "h1:J2FzIrXN82q5uyUraeJpLIm7U6PffRwje2ORho5yIik=",
        version = "v0.0.0-20220113124808-70ae35bab23f",
    )
    go_repository(
        name = "com_github_protonmail_go_mime",
        importpath = "github.com/ProtonMail/go-mime",
        sum = "h1:CGq7OieOz3wyQJ1fO8S0eO9TCW1JyvLrf8fhzz1i8ko=",
        version = "v0.0.0-20220302105931-303f85f7fe0f",
    )
    go_repository(
        name = "com_github_protonmail_gopenpgp_v2",
        importpath = "github.com/ProtonMail/gopenpgp/v2",
        sum = "h1:V3xeelvXgJiZXZuPtSSE+uYbtPw4RmbmyPqXDAESPhg=",
        version = "v2.4.7",
    )
    go_repository(
        name = "com_github_rivo_uniseg",
        importpath = "github.com/rivo/uniseg",
        sum = "h1:S1pD9weZBuJdFmowNwbpi7BJ8TNftyUImj/0WQi72jY=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_robfig_cron",
        importpath = "github.com/robfig/cron",
        sum = "h1:ZjScXvvxeQ63Dbyxy76Fj3AT3Ut0aKsyd2/tl3DTMuQ=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_rocky_linux_srpmproc",
        importpath = "github.com/rocky-linux/srpmproc",
        sum = "h1:PnfuOQbGZEwOEr1u7Notz9Pm+Ol38Tu7mo8/63rgeaE=",
        version = "v0.4.3",
    )
    go_repository(
        name = "com_github_rogpeppe_fastuuid",
        importpath = "github.com/rogpeppe/fastuuid",
        sum = "h1:Ppwyp6VYCF1nvBTXL3trRso7mXMlRrw9ooo375wvi2s=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_rogpeppe_go_internal",
        importpath = "github.com/rogpeppe/go-internal",
        sum = "h1:TMyTOH3F/DB16zRVcYyreMH6GnZZrwQVAoYjRBZyWFQ=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_russross_blackfriday_v2",
        importpath = "github.com/russross/blackfriday/v2",
        sum = "h1:lPqVAte+HuHNfhJ/0LC98ESWRz8afy9tM/0RK8m9o+Q=",
        version = "v2.0.1",
    )
    go_repository(
        name = "com_github_ryanuber_columnize",
        importpath = "github.com/ryanuber/columnize",
        sum = "h1:UFr9zpz4xgTnIE5yIMtWAMngCdZ9p/+q6lTbgelo80M=",
        version = "v0.0.0-20160712163229-9b3edd62028f",
    )
    go_repository(
        name = "com_github_sean_seed",
        importpath = "github.com/sean-/seed",
        sum = "h1:nn5Wsu0esKSJiIVhscUtVbo7ada43DJhG55ua/hjS5I=",
        version = "v0.0.0-20170313163322-e2103e2c3529",
    )
    go_repository(
        name = "com_github_sergi_go_diff",
        importpath = "github.com/sergi/go-diff",
        sum = "h1:we8PVUC3FE2uYfodKH/nBHMSetSfHDR6scGdBi+erh0=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_shurcool_sanitized_anchor_name",
        importpath = "github.com/shurcooL/sanitized_anchor_name",
        sum = "h1:PdmoCO6wvbs+7yrJyMORt4/BmY5IYyJwS/kOiWx8mHo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_sirupsen_logrus",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:dJKuHgqk1NNQlqoA6BTlM1Wf9DOH3NBjQyu0h9+AZZE=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_smartystreets_assertions",
        importpath = "github.com/smartystreets/assertions",
        sum = "h1:zE9ykElWQ6/NYmHa3jpm/yHnI4xSofP+UP6SpjHcSeM=",
        version = "v0.0.0-20180927180507-b2de0cb4f26d",
    )
    go_repository(
        name = "com_github_smartystreets_goconvey",
        importpath = "github.com/smartystreets/goconvey",
        sum = "h1:fv0U8FUIMPNf1L9lnHLvLhgicrIVChEkdzIKYqbNC9s=",
        version = "v1.6.4",
    )
    go_repository(
        name = "com_github_soheilhy_cmux",
        importpath = "github.com/soheilhy/cmux",
        sum = "h1:0HKaf1o97UwFjHH9o5XsHUOF+tqmdA7KEzXLpiyaw0E=",
        version = "v0.1.4",
    )
    go_repository(
        name = "com_github_spaolacci_murmur3",
        importpath = "github.com/spaolacci/murmur3",
        sum = "h1:qLC7fQah7D6K1B0ujays3HV9gkFtllcxhzImRR7ArPQ=",
        version = "v0.0.0-20180118202830-f09979ecbc72",
    )
    go_repository(
        name = "com_github_spf13_afero",
        importpath = "github.com/spf13/afero",
        sum = "h1:EaGW2JJh15aKOejeuJ+wpFSHnbd7GE6Wvp3TsNhb6LY=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_spf13_cast",
        importpath = "github.com/spf13/cast",
        sum = "h1:oget//CVOEoFewqQxwr0Ej5yjygnqGkvggSE/gB35Q8=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_spf13_cobra",
        importpath = "github.com/spf13/cobra",
        sum = "h1:e5/vxKd/rZsfSJMUX1agtjeTDf+qv1/JdBF8gg5k9ZM=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_spf13_jwalterweatherman",
        importpath = "github.com/spf13/jwalterweatherman",
        sum = "h1:XHEdyB+EcvlqZamSM4ZOMGlc93t6AcsBEu9Gc1vn7yk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_spf13_pflag",
        importpath = "github.com/spf13/pflag",
        sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
        version = "v1.0.5",
    )
    go_repository(
        name = "com_github_spf13_viper",
        importpath = "github.com/spf13/viper",
        sum = "h1:pM5oEahlgWv/WnHXpgbKz7iLIxRf65tye2Ci+XFK5sk=",
        version = "v1.7.1",
    )
    go_repository(
        name = "com_github_stretchr_objx",
        importpath = "github.com/stretchr/objx",
        sum = "h1:1zr/of2m5FGMsad5YfcqgdqdWrIhu+EBEJRhR1U7z/c=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_stretchr_testify",
        importpath = "github.com/stretchr/testify",
        sum = "h1:CcVxjf3Q8PM0mHUKJCdn+eZZtm5yQwehR5yeSVQQcUk=",
        version = "v1.8.4",
    )
    go_repository(
        name = "com_github_subosito_gotenv",
        importpath = "github.com/subosito/gotenv",
        sum = "h1:Slr1R9HxAlEKefgq5jn9U+DnETlIUa6HfgEzj0g5d7s=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_tmc_grpc_websocket_proxy",
        importpath = "github.com/tmc/grpc-websocket-proxy",
        sum = "h1:LnC5Kc/wtumK+WB441p7ynQJzVuNRJiqddSIE3IlSEQ=",
        version = "v0.0.0-20190109142713-0ad062ec5ee5",
    )
    go_repository(
        name = "com_github_vbauerster_mpb_v7",
        importpath = "github.com/vbauerster/mpb/v7",
        sum = "h1:eN6AD/ytv1nqCO7Dm8MO0/pGMKmMyH/WMnTJhAUuc/w=",
        version = "v7.0.2",
    )
    go_repository(
        name = "com_github_vividcortex_ewma",
        importpath = "github.com/VividCortex/ewma",
        sum = "h1:f58SaIzcDXrSy3kWaHNvuJgJ3Nmz59Zji6XoJR/q1ow=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_xanzy_go_gitlab",
        importpath = "github.com/xanzy/go-gitlab",
        sum = "h1:DA0+D9eTBnZvrGBOQ66XV4ZV9gOFXKbbgm9L99EKCUs=",
        version = "v0.50.4",
    )
    go_repository(
        name = "com_github_xanzy_ssh_agent",
        importpath = "github.com/xanzy/ssh-agent",
        sum = "h1:TCbipTQL2JiiCprBWx9frJ2eJlCYT00NmctrHxVAr70=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_xiang90_probing",
        importpath = "github.com/xiang90/probing",
        sum = "h1:eY9dn8+vbi4tKz5Qo6v2eYzo7kUS51QINcR5jNpbZS8=",
        version = "v0.0.0-20190116061207-43a291ad63a2",
    )
    go_repository(
        name = "com_github_yuin_goldmark",
        importpath = "github.com/yuin/goldmark",
        sum = "h1:fVcFKWvrslecOb/tg+Cc05dkeYx540o0FuFt3nUVDoE=",
        version = "v1.4.13",
    )
    go_repository(
        name = "com_google_cloud_go",
        importpath = "cloud.google.com/go",
        sum = "h1:tpFCD7hpHFlQ8yPwT3x+QeXqc2T6+n6T+hmABHfDUSM=",
        version = "v0.112.0",
    )
    go_repository(
        name = "com_google_cloud_go_accessapproval",
        importpath = "cloud.google.com/go/accessapproval",
        sum = "h1:uzmAMSgYcnlHa9X9YSQZ4Q1wlfl4NNkZyQgho1Z6p04=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_accesscontextmanager",
        importpath = "cloud.google.com/go/accesscontextmanager",
        sum = "h1:2GLNaNu9KRJhJBFTIVRoPwk6xE5mUDgD47abBq4Zp/I=",
        version = "v1.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_aiplatform",
        importpath = "cloud.google.com/go/aiplatform",
        sum = "h1:0cSrii1ZeLr16MbBoocyy5KVnrSdiQ3KN/vtrTe7RqE=",
        version = "v1.60.0",
    )
    go_repository(
        name = "com_google_cloud_go_analytics",
        importpath = "cloud.google.com/go/analytics",
        sum = "h1:Q+y94XH84jM8SK8O7qiY/PJRexb6n7dRbQ6PiUa4YGM=",
        version = "v0.23.0",
    )
    go_repository(
        name = "com_google_cloud_go_apigateway",
        importpath = "cloud.google.com/go/apigateway",
        sum = "h1:sPXnpk+6TneKIrjCjcpX5YGsAKy3PTdpIchoj8/74OE=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeconnect",
        importpath = "cloud.google.com/go/apigeeconnect",
        sum = "h1:CrfIKv9Go3fh/QfQgisU3MeP90Ww7l/sVGmr3TpECo8=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_apigeeregistry",
        importpath = "cloud.google.com/go/apigeeregistry",
        sum = "h1:C+QU2K+DzDjk4g074ouwHQGkoff1h5OMQp6sblCVreQ=",
        version = "v0.8.3",
    )
    go_repository(
        name = "com_google_cloud_go_appengine",
        importpath = "cloud.google.com/go/appengine",
        sum = "h1:l2SviT44zWQiOv8bPoMBzW0vOcMO22iO0s+nVtVhdts=",
        version = "v1.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_area120",
        importpath = "cloud.google.com/go/area120",
        sum = "h1:vTs08KPLN/iMzTbxpu5ciL06KcsrVPMjz4IwcQyZ4uY=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_artifactregistry",
        importpath = "cloud.google.com/go/artifactregistry",
        sum = "h1:W9sVlyb1VRcUf83w7aM3yMsnp4HS4PoyGqYQNG0O5lI=",
        version = "v1.14.7",
    )
    go_repository(
        name = "com_google_cloud_go_asset",
        importpath = "cloud.google.com/go/asset",
        sum = "h1:xgFnBP3luSbUcC9RWJvb3Zkt+y/wW6PKwPHr3ssnIP8=",
        version = "v1.17.2",
    )
    go_repository(
        name = "com_google_cloud_go_assuredworkloads",
        importpath = "cloud.google.com/go/assuredworkloads",
        sum = "h1:gCrN3IyvqY3cP0wh2h43d99CgH3G+WYs9CeuFVKChR8=",
        version = "v1.11.5",
    )
    go_repository(
        name = "com_google_cloud_go_automl",
        importpath = "cloud.google.com/go/automl",
        sum = "h1:ijiJy9sYWh75WrqImXsfWc1e3HR3iO+ef9fvW03Ig/4=",
        version = "v1.13.5",
    )
    go_repository(
        name = "com_google_cloud_go_baremetalsolution",
        importpath = "cloud.google.com/go/baremetalsolution",
        sum = "h1:LFydisRmS7hQk9P/YhekwuZGqb45TW4QavcrMToWo5A=",
        version = "v1.2.4",
    )
    go_repository(
        name = "com_google_cloud_go_batch",
        importpath = "cloud.google.com/go/batch",
        sum = "h1:2HK4JerwVaIcCh/lJiHwh6+uswPthiMMWhiSWLELayk=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_beyondcorp",
        importpath = "cloud.google.com/go/beyondcorp",
        sum = "h1:qs0J0O9Ol2h1yA0AU+r7l3hOCPzs2MjE1d6d/kaHIKo=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_google_cloud_go_bigquery",
        importpath = "cloud.google.com/go/bigquery",
        sum = "h1:CpT+/njKuKT3CEmswm6IbhNu9u35zt5dO4yPDLW+nG4=",
        version = "v1.59.1",
    )
    go_repository(
        name = "com_google_cloud_go_billing",
        importpath = "cloud.google.com/go/billing",
        sum = "h1:oWUEQvuC4JvtnqLZ35zgzdbuHt4Itbftvzbe6aEyFdE=",
        version = "v1.18.2",
    )
    go_repository(
        name = "com_google_cloud_go_binaryauthorization",
        importpath = "cloud.google.com/go/binaryauthorization",
        sum = "h1:1jcyh2uIUwSZkJ/JmL8kd5SUkL/Krbv8zmYLEbAz6kY=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_certificatemanager",
        importpath = "cloud.google.com/go/certificatemanager",
        sum = "h1:UMBr/twXvH3jcT5J5/YjRxf2tvwTYIfrpemTebe0txc=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_channel",
        importpath = "cloud.google.com/go/channel",
        sum = "h1:/omiBnyFjm4S1ETHoOmJbL7LH7Ljcei4rYG6Sj3hc80=",
        version = "v1.17.5",
    )
    go_repository(
        name = "com_google_cloud_go_cloudbuild",
        importpath = "cloud.google.com/go/cloudbuild",
        sum = "h1:ZB6oOmJo+MTov9n629fiCrO9YZPOg25FZvQ7gIHu5ng=",
        version = "v1.15.1",
    )
    go_repository(
        name = "com_google_cloud_go_clouddms",
        importpath = "cloud.google.com/go/clouddms",
        sum = "h1:Sr0Zo5EAcPQiCBgHWICg3VGkcdS/LLP1d9SR7qQBM/s=",
        version = "v1.7.4",
    )
    go_repository(
        name = "com_google_cloud_go_cloudtasks",
        importpath = "cloud.google.com/go/cloudtasks",
        sum = "h1:EUt1hIZ9bLv8Iz9yWaCrqgMnIU+Tdh0yXM1MMVGhjfE=",
        version = "v1.12.6",
    )
    go_repository(
        name = "com_google_cloud_go_compute",
        importpath = "cloud.google.com/go/compute",
        sum = "h1:phWcR2eWzRJaL/kOiJwfFsPs4BaKq1j6vnpZrc1YlVg=",
        version = "v1.24.0",
    )
    go_repository(
        name = "com_google_cloud_go_compute_metadata",
        importpath = "cloud.google.com/go/compute/metadata",
        sum = "h1:mg4jlk7mCAj6xXp9UJ4fjI9VUI5rubuGBW5aJ7UnBMY=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_google_cloud_go_contactcenterinsights",
        importpath = "cloud.google.com/go/contactcenterinsights",
        sum = "h1:6Vs/YnDG5STGjlWMEjN/xtmft7MrOTOnOZYUZtGTx0w=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_google_cloud_go_container",
        importpath = "cloud.google.com/go/container",
        sum = "h1:MAaNH7VRNPWEhvqOypq2j+7ONJKrKzon4v9nS3nLZe0=",
        version = "v1.31.0",
    )
    go_repository(
        name = "com_google_cloud_go_containeranalysis",
        importpath = "cloud.google.com/go/containeranalysis",
        sum = "h1:doJ0M1ljS4hS0D2UbHywlHGwB7sQLNrt9vFk9Zyi7vY=",
        version = "v0.11.4",
    )
    go_repository(
        name = "com_google_cloud_go_datacatalog",
        importpath = "cloud.google.com/go/datacatalog",
        sum = "h1:A0vKYCQdxQuV4Pi0LL9p39Vwvg4jH5yYveMv50gU5Tw=",
        version = "v1.19.3",
    )
    go_repository(
        name = "com_google_cloud_go_dataflow",
        importpath = "cloud.google.com/go/dataflow",
        sum = "h1:RYHtcPhmE664+F0Je46p+NvFbG8z//KCXp+uEqB4jZU=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_dataform",
        importpath = "cloud.google.com/go/dataform",
        sum = "h1:5e4eqGrd0iDTCg4Q+VlAao5j2naKAA7xRurNtwmUknU=",
        version = "v0.9.2",
    )
    go_repository(
        name = "com_google_cloud_go_datafusion",
        importpath = "cloud.google.com/go/datafusion",
        sum = "h1:HQ/BUOP8OIGJxuztpYvNvlb+/U+/Bfs9SO8tQbh61fk=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_datalabeling",
        importpath = "cloud.google.com/go/datalabeling",
        sum = "h1:GpIFRdm0qIZNsxqURFJwHt0ZBJZ0nF/mUVEigR7PH/8=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_dataplex",
        importpath = "cloud.google.com/go/dataplex",
        sum = "h1:fxIfdU8fxzR3clhOoNI7XFppvAmndxDu1AMH+qX9WKQ=",
        version = "v1.14.2",
    )
    go_repository(
        name = "com_google_cloud_go_dataproc_v2",
        importpath = "cloud.google.com/go/dataproc/v2",
        sum = "h1:/u81Fd+BvCLp+xjctI1DiWVJn6cn9/s3Akc8xPH02yk=",
        version = "v2.4.0",
    )
    go_repository(
        name = "com_google_cloud_go_dataqna",
        importpath = "cloud.google.com/go/dataqna",
        sum = "h1:9ybXs3nr9BzxSGC04SsvtuXaHY0qmJSLIpIAbZo9GqQ=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_datastore",
        importpath = "cloud.google.com/go/datastore",
        sum = "h1:0P9WcsQeTWjuD1H14JIY7XQscIPQ4Laje8ti96IC5vg=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_google_cloud_go_datastream",
        importpath = "cloud.google.com/go/datastream",
        sum = "h1:o1QDKMo/hk0FN7vhoUQURREuA0rgKmnYapB+1M+7Qz4=",
        version = "v1.10.4",
    )
    go_repository(
        name = "com_google_cloud_go_deploy",
        importpath = "cloud.google.com/go/deploy",
        sum = "h1:m27Ojwj03gvpJqCbodLYiVmE9x4/LrHGGMjzc0LBfM4=",
        version = "v1.17.1",
    )
    go_repository(
        name = "com_google_cloud_go_dialogflow",
        importpath = "cloud.google.com/go/dialogflow",
        sum = "h1:KqG0oxGE71qo0lRVyAoeBozefCvsMfcDzDjoLYSY0F4=",
        version = "v1.49.0",
    )
    go_repository(
        name = "com_google_cloud_go_dlp",
        importpath = "cloud.google.com/go/dlp",
        sum = "h1:lTipOuJaSjlYnnotPMbEhKURLC6GzCMDDzVbJAEbmYM=",
        version = "v1.11.2",
    )
    go_repository(
        name = "com_google_cloud_go_documentai",
        importpath = "cloud.google.com/go/documentai",
        sum = "h1:lI62GMEEPO6vXJI9hj+G9WjOvnR0hEjvjokrnex4cxA=",
        version = "v1.25.0",
    )
    go_repository(
        name = "com_google_cloud_go_domains",
        importpath = "cloud.google.com/go/domains",
        sum = "h1:Mml/R6s3vQQvFPpi/9oX3O5dRirgjyJ8cksK8N19Y7g=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_edgecontainer",
        importpath = "cloud.google.com/go/edgecontainer",
        sum = "h1:tBY32km78ScpK2aOP84JoW/+wtpx5WluyPUSEE3270U=",
        version = "v1.1.5",
    )
    go_repository(
        name = "com_google_cloud_go_errorreporting",
        importpath = "cloud.google.com/go/errorreporting",
        sum = "h1:kj1XEWMu8P0qlLhm3FwcaFsUvXChV/OraZwA70trRR0=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_google_cloud_go_essentialcontacts",
        importpath = "cloud.google.com/go/essentialcontacts",
        sum = "h1:13eHn5qBnsawxI7mIrv4jRIEmQ1xg0Ztqw5ZGqtUNfA=",
        version = "v1.6.6",
    )
    go_repository(
        name = "com_google_cloud_go_eventarc",
        importpath = "cloud.google.com/go/eventarc",
        sum = "h1:ORkd6/UV5FIdA8KZQDLNZYKS7BBOrj0p01DXPmT4tE4=",
        version = "v1.13.4",
    )
    go_repository(
        name = "com_google_cloud_go_filestore",
        importpath = "cloud.google.com/go/filestore",
        sum = "h1:X5G4y/vrUo1B8Nsz93qSWTMAcM8LXbGUldq33OdcdCw=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_firestore",
        importpath = "cloud.google.com/go/firestore",
        sum = "h1:8aLcKnMPoldYU3YHgu4t2exrKhLQkqaXAGqT0ljrFVw=",
        version = "v1.14.0",
    )
    go_repository(
        name = "com_google_cloud_go_functions",
        importpath = "cloud.google.com/go/functions",
        sum = "h1:IWVylmK5F6hJ3R5zaRW7jI5PrWhCvtBVU4axQLmXSo4=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_google_cloud_go_gkebackup",
        importpath = "cloud.google.com/go/gkebackup",
        sum = "h1:iuE8KNtTsPOc79qeWoNS8zOWoXPD9SAdOmwgxtlCmh8=",
        version = "v1.3.5",
    )
    go_repository(
        name = "com_google_cloud_go_gkeconnect",
        importpath = "cloud.google.com/go/gkeconnect",
        sum = "h1:17d+ZSSXKqG/RwZCq3oFMIWLPI8Zw3b8+a9/BEVlwH0=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_gkehub",
        importpath = "cloud.google.com/go/gkehub",
        sum = "h1:RboLNFzf9wEMSo7DrKVBlf+YhK/A/jrLN454L5Tz99Q=",
        version = "v0.14.5",
    )
    go_repository(
        name = "com_google_cloud_go_gkemulticloud",
        importpath = "cloud.google.com/go/gkemulticloud",
        sum = "h1:rsSZAGLhyjyE/bE2ToT5fqo1qSW7S+Ubsc9jFOcbhSI=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_google_cloud_go_gsuiteaddons",
        importpath = "cloud.google.com/go/gsuiteaddons",
        sum = "h1:CZEbaBwmbYdhFw21Fwbo+C35HMe36fTE0FBSR4KSfWg=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_iam",
        importpath = "cloud.google.com/go/iam",
        sum = "h1:bEa06k05IO4f4uJonbB5iAgKTPpABy1ayxaIZV/GHVc=",
        version = "v1.1.6",
    )
    go_repository(
        name = "com_google_cloud_go_iap",
        importpath = "cloud.google.com/go/iap",
        sum = "h1:94zirc2r4t6KzhAMW0R6Dme005eTP6yf7g6vN4IhRrA=",
        version = "v1.9.4",
    )
    go_repository(
        name = "com_google_cloud_go_ids",
        importpath = "cloud.google.com/go/ids",
        sum = "h1:xd4U7pgl3GHV+MABnv1BF4/Vy/zBF7CYC8XngkOLzag=",
        version = "v1.4.5",
    )
    go_repository(
        name = "com_google_cloud_go_iot",
        importpath = "cloud.google.com/go/iot",
        sum = "h1:munTeBlbqI33iuTYgXy7S8lW2TCgi5l1hA4roSIY+EE=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_kms",
        importpath = "cloud.google.com/go/kms",
        sum = "h1:7caV9K3yIxvlQPAcaFffhlT7d1qpxjB1wHBtjWa13SM=",
        version = "v1.15.7",
    )
    go_repository(
        name = "com_google_cloud_go_language",
        importpath = "cloud.google.com/go/language",
        sum = "h1:iaJZg6K4j/2PvZZVcjeO/btcWWIllVRBhuTFjGO4LXs=",
        version = "v1.12.3",
    )
    go_repository(
        name = "com_google_cloud_go_lifesciences",
        importpath = "cloud.google.com/go/lifesciences",
        sum = "h1:gXvN70m2p+4zgJFzaz6gMKaxTuF9WJ0USYoMLWAOm8g=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_logging",
        importpath = "cloud.google.com/go/logging",
        sum = "h1:iEIOXFO9EmSiTjDmfpbRjOxECO7R8C7b8IXUGOj7xZw=",
        version = "v1.9.0",
    )
    go_repository(
        name = "com_google_cloud_go_longrunning",
        importpath = "cloud.google.com/go/longrunning",
        sum = "h1:GOE6pZFdSrTb4KAiKnXsJBtlE6mEyaW44oKyMILWnOg=",
        version = "v0.5.5",
    )
    go_repository(
        name = "com_google_cloud_go_managedidentities",
        importpath = "cloud.google.com/go/managedidentities",
        sum = "h1:+bpih1piZVLxla/XBqeSUzJBp8gv9plGHIMAI7DLpDM=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_maps",
        importpath = "cloud.google.com/go/maps",
        sum = "h1:EVCZAiDvog9So46460BGbCasPhi613exoaQbpilMVlk=",
        version = "v1.6.4",
    )
    go_repository(
        name = "com_google_cloud_go_mediatranslation",
        importpath = "cloud.google.com/go/mediatranslation",
        sum = "h1:c76KdIXljQHSCb/Cy47S8H4s05A4zbK3pAFGzwcczZo=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_memcache",
        importpath = "cloud.google.com/go/memcache",
        sum = "h1:yeDv5qxRedFosvpMSEswrqUsJM5OdWvssPHFliNFTc4=",
        version = "v1.10.5",
    )
    go_repository(
        name = "com_google_cloud_go_metastore",
        importpath = "cloud.google.com/go/metastore",
        sum = "h1:dR7vqWXlK6IYR8Wbu9mdFfwlVjodIBhd1JRrpZftTEg=",
        version = "v1.13.4",
    )
    go_repository(
        name = "com_google_cloud_go_monitoring",
        importpath = "cloud.google.com/go/monitoring",
        sum = "h1:NfkDLQDG2UR3WYZVQE8kwSbUIEyIqJUPl+aOQdFH1T4=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_google_cloud_go_networkconnectivity",
        importpath = "cloud.google.com/go/networkconnectivity",
        sum = "h1:GBfXFhLyPspnaBE3nI/BRjdhW8vcbpT9QjE/4kDCDdc=",
        version = "v1.14.4",
    )
    go_repository(
        name = "com_google_cloud_go_networkmanagement",
        importpath = "cloud.google.com/go/networkmanagement",
        sum = "h1:aLV5GcosBNmd6M8+a0ekB0XlLRexv4fvnJJrYnqeBcg=",
        version = "v1.9.4",
    )
    go_repository(
        name = "com_google_cloud_go_networksecurity",
        importpath = "cloud.google.com/go/networksecurity",
        sum = "h1:+caSxBTj0E8OYVh/5wElFdjEMO1S/rZtE1152Cepchc=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_notebooks",
        importpath = "cloud.google.com/go/notebooks",
        sum = "h1:FH48boYmrWVQ6k0Mx/WrnNafXncT5iSYxA8CNyWTgy0=",
        version = "v1.11.3",
    )
    go_repository(
        name = "com_google_cloud_go_optimization",
        importpath = "cloud.google.com/go/optimization",
        sum = "h1:63NZaWyN+5rZEKHPX4ACpw3BjgyeuY8+rCehiCMaGPY=",
        version = "v1.6.3",
    )
    go_repository(
        name = "com_google_cloud_go_orchestration",
        importpath = "cloud.google.com/go/orchestration",
        sum = "h1:YHgWMlrPttIVGItgGfuvO2KM7x+y9ivN/Yk92pMm1a4=",
        version = "v1.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_orgpolicy",
        importpath = "cloud.google.com/go/orgpolicy",
        sum = "h1:2JbXigqBJVp8Dx5dONUttFqewu4fP0p3pgOdIZAhpYU=",
        version = "v1.12.1",
    )
    go_repository(
        name = "com_google_cloud_go_osconfig",
        importpath = "cloud.google.com/go/osconfig",
        sum = "h1:Mo5jGAxOMKH/PmDY7fgY19yFcVbvwREb5D5zMPQjFfo=",
        version = "v1.12.5",
    )
    go_repository(
        name = "com_google_cloud_go_oslogin",
        importpath = "cloud.google.com/go/oslogin",
        sum = "h1:1K4nOT5VEZNt7XkhaTXupBYos5HjzvJMfhvyD2wWdFs=",
        version = "v1.13.1",
    )
    go_repository(
        name = "com_google_cloud_go_phishingprotection",
        importpath = "cloud.google.com/go/phishingprotection",
        sum = "h1:DH3WFLzEoJdW/6xgsmoDqOwT1xddFi7gKu0QGZQhpGU=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_policytroubleshooter",
        importpath = "cloud.google.com/go/policytroubleshooter",
        sum = "h1:c0WOzC6hz964QWNBkyKfna8A2jOIx1zzZa43Gx/P09o=",
        version = "v1.10.3",
    )
    go_repository(
        name = "com_google_cloud_go_privatecatalog",
        importpath = "cloud.google.com/go/privatecatalog",
        sum = "h1:UZ0assTnATXSggoxUIh61RjTQ4P9zCMk/kEMbn0nMYA=",
        version = "v0.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_pubsub",
        importpath = "cloud.google.com/go/pubsub",
        sum = "h1:dfEPuGCHGbWUhaMCTHUFjfroILEkx55iUmKBZTP5f+Y=",
        version = "v1.36.1",
    )
    go_repository(
        name = "com_google_cloud_go_pubsublite",
        importpath = "cloud.google.com/go/pubsublite",
        sum = "h1:pX+idpWMIH30/K7c0epN6V703xpIcMXWRjKJsz0tYGY=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_google_cloud_go_recaptchaenterprise_v2",
        importpath = "cloud.google.com/go/recaptchaenterprise/v2",
        sum = "h1:U3Wfq12X9cVMuTpsWDSURnXF0Z9hSPTHj+xsnXDRLsw=",
        version = "v2.9.2",
    )
    go_repository(
        name = "com_google_cloud_go_recommendationengine",
        importpath = "cloud.google.com/go/recommendationengine",
        sum = "h1:ineqLswaCSBY0csYv5/wuXJMBlxATK6Xc5jJkpiTEdM=",
        version = "v0.8.5",
    )
    go_repository(
        name = "com_google_cloud_go_recommender",
        importpath = "cloud.google.com/go/recommender",
        sum = "h1:LVLYS3r3u0MSCxQSDUtLSkporEGi9OAE6hGvayrZNPs=",
        version = "v1.12.1",
    )
    go_repository(
        name = "com_google_cloud_go_redis",
        importpath = "cloud.google.com/go/redis",
        sum = "h1:QF0maEdVv0Fj/2roU8sX3NpiDBzP9ICYTO+5F32gQNo=",
        version = "v1.14.2",
    )
    go_repository(
        name = "com_google_cloud_go_resourcemanager",
        importpath = "cloud.google.com/go/resourcemanager",
        sum = "h1:AZWr1vWVDKGwfLsVhcN+vcwOz3xqqYxtmMa0aABCMms=",
        version = "v1.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_resourcesettings",
        importpath = "cloud.google.com/go/resourcesettings",
        sum = "h1:BTr5MVykJwClASci/7Og4Qfx70aQ4n3epsNLj94ZYgw=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_retail",
        importpath = "cloud.google.com/go/retail",
        sum = "h1:Fn1GuAua1c6crCGqfJ1qMxG1Xh10Tg/x5EUODEHMqkw=",
        version = "v1.16.0",
    )
    go_repository(
        name = "com_google_cloud_go_run",
        importpath = "cloud.google.com/go/run",
        sum = "h1:m9WDA7DzTpczhZggwYlZcBWgCRb+kgSIisWn1sbw2rQ=",
        version = "v1.3.4",
    )
    go_repository(
        name = "com_google_cloud_go_scheduler",
        importpath = "cloud.google.com/go/scheduler",
        sum = "h1:5U8iXLoQ03qOB+ZXlAecU7fiE33+u3QiM9nh4cd0eTE=",
        version = "v1.10.6",
    )
    go_repository(
        name = "com_google_cloud_go_secretmanager",
        importpath = "cloud.google.com/go/secretmanager",
        sum = "h1:82fpF5vBBvu9XW4qj0FU2C6qVMtj1RM/XHwKXUEAfYY=",
        version = "v1.11.5",
    )
    go_repository(
        name = "com_google_cloud_go_security",
        importpath = "cloud.google.com/go/security",
        sum = "h1:wTKJQ10j8EYgvE8Y+KhovxDRVDk2iv/OsxZ6GrLP3kE=",
        version = "v1.15.5",
    )
    go_repository(
        name = "com_google_cloud_go_securitycenter",
        importpath = "cloud.google.com/go/securitycenter",
        sum = "h1:/5jjkZ+uGe8hZ7pvd7pO30VW/a+pT2MrrdgOqjyucKQ=",
        version = "v1.24.4",
    )
    go_repository(
        name = "com_google_cloud_go_servicedirectory",
        importpath = "cloud.google.com/go/servicedirectory",
        sum = "h1:da7HFI1229kyzIyuVEzHXip0cw0d+E0s8mjQby0WN+k=",
        version = "v1.11.4",
    )
    go_repository(
        name = "com_google_cloud_go_shell",
        importpath = "cloud.google.com/go/shell",
        sum = "h1:3Fq2hzO0ZSyaqBboJrFkwwf/qMufDtqwwA6ep8EZxEI=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_spanner",
        importpath = "cloud.google.com/go/spanner",
        sum = "h1:fJq+ZfQUDHE+cy1li0bJA8+sy2oiSGhuGqN5nqVaZdU=",
        version = "v1.57.0",
    )
    go_repository(
        name = "com_google_cloud_go_speech",
        importpath = "cloud.google.com/go/speech",
        sum = "h1:nuFc+Kj5B8de75nN4FdPyUbI2SiBoHZG6BLurXL56Q0=",
        version = "v1.21.1",
    )
    go_repository(
        name = "com_google_cloud_go_storage",
        importpath = "cloud.google.com/go/storage",
        sum = "h1:P0mOkAcaJxhCTvAkMhxMfrTKiNcub4YmmPBtlhAyTr8=",
        version = "v1.36.0",
    )
    go_repository(
        name = "com_google_cloud_go_storagetransfer",
        importpath = "cloud.google.com/go/storagetransfer",
        sum = "h1:dy4fL3wO0VABvzM05ycMUPFHxTPbJz9Em8ikAJVqSbI=",
        version = "v1.10.4",
    )
    go_repository(
        name = "com_google_cloud_go_talent",
        importpath = "cloud.google.com/go/talent",
        sum = "h1:JssV0CE3FNujuSWn7SkosOzg7qrMxVnt6txOfGcMSa4=",
        version = "v1.6.6",
    )
    go_repository(
        name = "com_google_cloud_go_texttospeech",
        importpath = "cloud.google.com/go/texttospeech",
        sum = "h1:dxY2Q5mHCbrGa3oPR2O3PCicdnvKa1JmwGQK36EFLOw=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_tpu",
        importpath = "cloud.google.com/go/tpu",
        sum = "h1:C8YyYda8WtNdBoCgFwwBzZd+S6+EScHOxM/z1h0NNp8=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_trace",
        importpath = "cloud.google.com/go/trace",
        sum = "h1:0pr4lIKJ5XZFYD9GtxXEWr0KkVeigc3wlGpZco0X1oA=",
        version = "v1.10.5",
    )
    go_repository(
        name = "com_google_cloud_go_translate",
        importpath = "cloud.google.com/go/translate",
        sum = "h1:upovZ0wRMdzZvXnu+RPam41B0mRJ+coRXFP2cYFJ7ew=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_google_cloud_go_video",
        importpath = "cloud.google.com/go/video",
        sum = "h1:TXwotxkShP1OqgKsbd+b8N5hrIHavSyLGvYnLGCZ7xc=",
        version = "v1.20.4",
    )
    go_repository(
        name = "com_google_cloud_go_videointelligence",
        importpath = "cloud.google.com/go/videointelligence",
        sum = "h1:mYaWH8uhUCXLJCN3gdXswKzRa2+lK0zN6/KsIubm6pE=",
        version = "v1.11.5",
    )
    go_repository(
        name = "com_google_cloud_go_vision_v2",
        importpath = "cloud.google.com/go/vision/v2",
        sum = "h1:W52z1b6LdGI66MVhE70g/NFty9zCYYcjdKuycqmlhtg=",
        version = "v2.8.0",
    )
    go_repository(
        name = "com_google_cloud_go_vmmigration",
        importpath = "cloud.google.com/go/vmmigration",
        sum = "h1:5v9RT2vWyuw3pK2ox0HQpkoftO7Q7/8591dTxxQc79g=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_vmwareengine",
        importpath = "cloud.google.com/go/vmwareengine",
        sum = "h1:EGdDi9QbqThfZq3ILcDK5g+m9jTevc34AY5tACx5v7k=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_google_cloud_go_vpcaccess",
        importpath = "cloud.google.com/go/vpcaccess",
        sum = "h1:XyL6hTLtEM/eE4F1GEge8xUN9ZCkiVWn44K/YA7z1rQ=",
        version = "v1.7.5",
    )
    go_repository(
        name = "com_google_cloud_go_webrisk",
        importpath = "cloud.google.com/go/webrisk",
        sum = "h1:251MvGuC8wisNN7+jqu9DDDZAi38KiMXxOpA/EWy4dE=",
        version = "v1.9.5",
    )
    go_repository(
        name = "com_google_cloud_go_websecurityscanner",
        importpath = "cloud.google.com/go/websecurityscanner",
        sum = "h1:YqWZrZYabG88TZt7364XWRJGhxmxhony2ZUyZEYMF2k=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_google_cloud_go_workflows",
        importpath = "cloud.google.com/go/workflows",
        sum = "h1:uHNmUiatTbPQ4H1pabwfzpfEYD4BBnqDHqMm2IesOh4=",
        version = "v1.12.4",
    )
    go_repository(
        name = "com_shuralyov_dmitri_gpu_mtl",
        importpath = "dmitri.shuralyov.com/gpu/mtl",
        sum = "h1:VpgP7xuJadIUuKccphEpTJnWhS2jkQyMt6Y7pJCD7fY=",
        version = "v0.0.0-20190408044501-666a987793e9",
    )
    go_repository(
        name = "in_gopkg_alecthomas_kingpin_v2",
        importpath = "gopkg.in/alecthomas/kingpin.v2",
        sum = "h1:jMFz6MfLP0/4fUyZle81rXUoxOBFi19VUFKVDOQfozc=",
        version = "v2.2.6",
    )
    go_repository(
        name = "in_gopkg_check_v1",
        importpath = "gopkg.in/check.v1",
        sum = "h1:Hei/4ADfdWqJk1ZMxUNpqntNwaWcugrBjAiHlqqRiVk=",
        version = "v1.0.0-20201130134442-10cb98267c6c",
    )
    go_repository(
        name = "in_gopkg_errgo_v2",
        importpath = "gopkg.in/errgo.v2",
        sum = "h1:0vLT13EuvQ0hNvakwLuFZ/jYrLp5F3kcWHXdRggjCE8=",
        version = "v2.1.0",
    )
    go_repository(
        name = "in_gopkg_inf_v0",
        importpath = "gopkg.in/inf.v0",
        sum = "h1:73M5CoZyi3ZLMOyDlQh031Cx6N9NDJ2Vvfl76EDAgDc=",
        version = "v0.9.1",
    )
    go_repository(
        name = "in_gopkg_ini_v1",
        importpath = "gopkg.in/ini.v1",
        sum = "h1:9unxIsFcTt4I55uWluz+UmL95q4kdJ0buvQ1ZIqVQww=",
        version = "v1.57.0",
    )
    go_repository(
        name = "in_gopkg_resty_v1",
        importpath = "gopkg.in/resty.v1",
        sum = "h1:CuXP0Pjfw9rOuY6EP+UvtNvt5DSqHpIxILZKT/quCZI=",
        version = "v1.12.0",
    )
    go_repository(
        name = "in_gopkg_square_go_jose_v2",
        importpath = "gopkg.in/square/go-jose.v2",
        sum = "h1:7odma5RETjNHWJnR32wx8t+Io4djHE1PqxCFx3iiZ2w=",
        version = "v2.5.1",
    )
    go_repository(
        name = "in_gopkg_warnings_v0",
        importpath = "gopkg.in/warnings.v0",
        sum = "h1:wFXVbFY8DY5/xOe1ECiWdKCzZlxgshcYVNkBHstARME=",
        version = "v0.1.2",
    )
    go_repository(
        name = "in_gopkg_yaml_v2",
        importpath = "gopkg.in/yaml.v2",
        sum = "h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=",
        version = "v2.4.0",
    )
    go_repository(
        name = "in_gopkg_yaml_v3",
        importpath = "gopkg.in/yaml.v3",
        sum = "h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=",
        version = "v3.0.1",
    )
    go_repository(
        name = "io_alexejk_go_xmlrpc",
        importpath = "alexejk.io/go-xmlrpc",
        sum = "h1:3EdJ+qtYQjyXiE4E0V/GFAnUvvOjqh1+OGIgy5OSdZk=",
        version = "v0.2.0",
    )
    go_repository(
        name = "io_cirello_dynamolock",
        importpath = "cirello.io/dynamolock",
        sum = "h1:dBOcspk0HEJpy4bvHUPrk134PkiWCzllzchOU38i+r4=",
        version = "v1.4.0",
    )
    go_repository(
        name = "io_etcd_go_bbolt",
        importpath = "go.etcd.io/bbolt",
        sum = "h1:Z/90sZLPOeCy2PwprqkFa25PdkusRzaj9P8zm/KNyvk=",
        version = "v1.3.2",
    )
    go_repository(
        name = "io_k8s_api",
        importpath = "k8s.io/api",
        sum = "h1:hBC7B9+MU+ptchxEqTNW2DkUosJpp1P+Wn6YncZ474A=",
        version = "v0.29.2",
    )
    go_repository(
        name = "io_k8s_apimachinery",
        importpath = "k8s.io/apimachinery",
        sum = "h1:EWGpfJ856oj11C52NRCHuU7rFDwxev48z+6DSlGNsV8=",
        version = "v0.29.2",
    )
    go_repository(
        name = "io_k8s_client_go",
        importpath = "k8s.io/client-go",
        sum = "h1:FEg85el1TeZp+/vYJM7hkDlSTFZ+c5nnK44DJ4FyoRg=",
        version = "v0.29.2",
    )
    go_repository(
        name = "io_k8s_gengo",
        importpath = "k8s.io/gengo",
        sum = "h1:pWEwq4Asjm4vjW7vcsmijwBhOr1/shsbSYiWXmNGlks=",
        version = "v0.0.0-20230829151522-9cce18d56c01",
    )
    go_repository(
        name = "io_k8s_klog_v2",
        importpath = "k8s.io/klog/v2",
        sum = "h1:U/Af64HJf7FcwMcXyKm2RPM22WZzyR7OSpYj5tg3cL0=",
        version = "v2.110.1",
    )
    go_repository(
        name = "io_k8s_kube_openapi",
        importpath = "k8s.io/kube-openapi",
        sum = "h1:aVUu9fTY98ivBPKR9Y5w/AuzbMm96cd3YHRTU83I780=",
        version = "v0.0.0-20231010175941-2dd684a91f00",
    )
    go_repository(
        name = "io_k8s_sigs_json",
        importpath = "sigs.k8s.io/json",
        sum = "h1:EDPBXCAspyGV4jQlpZSudPeMmr1bNJefnuqLsRAsHZo=",
        version = "v0.0.0-20221116044647-bc3834ca7abd",
    )
    go_repository(
        name = "io_k8s_sigs_structured_merge_diff_v4",
        importpath = "sigs.k8s.io/structured-merge-diff/v4",
        sum = "h1:150L+0vs/8DA78h1u02ooW1/fFq/Lwr+sGiqlzvrtq4=",
        version = "v4.4.1",
    )
    go_repository(
        name = "io_k8s_sigs_yaml",
        importpath = "sigs.k8s.io/yaml",
        sum = "h1:a2VclLzOGrwOHDiV8EfBGhvjHvP46CtW5j6POvhYGGo=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_k8s_utils",
        importpath = "k8s.io/utils",
        sum = "h1:sgn3ZU783SCgtaSJjpcVVlRqd6GSnlTLKgpAAttJvpI=",
        version = "v0.0.0-20230726121419-3b25d923346b",
    )
    go_repository(
        name = "io_opencensus_go",
        importpath = "go.opencensus.io",
        sum = "h1:y73uSU6J157QMP2kn2r30vwW1A2W2WFwSCGnAVxeaD0=",
        version = "v0.24.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_google_golang_org_grpc_otelgrpc",
        importpath = "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
        sum = "h1:4Pp6oUg3+e/6M4C0A/3kJ2VYa++dsWVTtGgLVj5xtHg=",
        version = "v0.49.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_net_http_otelhttp",
        importpath = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
        sum = "h1:jq9TW8u3so/bN+JPT166wjOI6/vQPF6Xe7nMNIltagk=",
        version = "v0.49.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel",
        importpath = "go.opentelemetry.io/otel",
        sum = "h1:0LAOdjNmQeSTzGBzduGe/rU4tZhMwL5rWgtp9Ku5Jfo=",
        version = "v1.24.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_metric",
        importpath = "go.opentelemetry.io/otel/metric",
        sum = "h1:6EhoGWWK28x1fbpA4tYTOWBkPefTDQnb8WSGXlc88kI=",
        version = "v1.24.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_sdk",
        importpath = "go.opentelemetry.io/otel/sdk",
        sum = "h1:FTt8qirL1EysG6sTQRZ5TokkU8d0ugCj8htOgThZXQ8=",
        version = "v1.21.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_trace",
        importpath = "go.opentelemetry.io/otel/trace",
        sum = "h1:CsKnnL4dUAr/0llH9FKuc698G04IrpWV0MQA/Y1YELI=",
        version = "v1.24.0",
    )
    go_repository(
        name = "io_opentelemetry_go_proto_otlp",
        importpath = "go.opentelemetry.io/proto/otlp",
        sum = "h1:rwOQPCuKAKmwGKq2aVNnYIibI6wnV7EvzgfTCzcdGg8=",
        version = "v0.7.0",
    )
    go_repository(
        name = "io_rsc_binaryregexp",
        importpath = "rsc.io/binaryregexp",
        sum = "h1:HfqmD5MEmC0zvwBuF187nq9mdnXjXsSivRiXN7SmRkE=",
        version = "v0.2.0",
    )
    go_repository(
        name = "io_rsc_quote_v3",
        importpath = "rsc.io/quote/v3",
        sum = "h1:9JKUTTIUgS6kzR9mK1YuGKv6Nl+DijDNIc0ghT58FaY=",
        version = "v3.1.0",
    )
    go_repository(
        name = "io_rsc_sampler",
        importpath = "rsc.io/sampler",
        sum = "h1:7uVkIFmeBqHfdjD+gZwtXXI+RODJ2Wc4O7MPEh/QiW4=",
        version = "v1.3.0",
    )
    go_repository(
        name = "io_temporal_go_api",
        importpath = "go.temporal.io/api",
        sum = "h1:TN/PQNywCnOG/hXLHKkKKOQQtpi7JHBDD8fpv8H8JiA=",
        version = "v1.6.1-0.20211110205628-60c98e9cbfe2",
    )
    go_repository(
        name = "io_temporal_go_sdk",
        importpath = "go.temporal.io/sdk",
        sum = "h1:4LRIe1WLM+m2pN6sNod4sMV+0bV8WTscVfRsipOP8N8=",
        version = "v1.13.1",
    )
    go_repository(
        name = "org_golang_google_api",
        importpath = "google.golang.org/api",
        sum = "h1:6m4NUwrZYhAaVIHZWxaKjw1L1vNAjtMwORmKRyEEo24=",
        version = "v0.166.0",
    )
    go_repository(
        name = "org_golang_google_appengine",
        importpath = "google.golang.org/appengine",
        sum = "h1:IhEN5q69dyKagZPYMSdIjS2HqprW324FRQZJcGqPAsM=",
        version = "v1.6.8",
    )
    go_repository(
        name = "org_golang_google_genproto",
        importpath = "google.golang.org/genproto",
        sum = "h1:Zmyn5CV/jxzKnF+3d+xzbomACPwLQqVpLTpyXN5uTaQ=",
        version = "v0.0.0-20240221002015-b0ce06bbee7c",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_api",
        importpath = "google.golang.org/genproto/googleapis/api",
        sum = "h1:9g7erC9qu44ks7UK4gDNlnk4kOxZG707xKm4jVniy6o=",
        version = "v0.0.0-20240221002015-b0ce06bbee7c",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_bytestream",
        importpath = "google.golang.org/genproto/googleapis/bytestream",
        sum = "h1:SlwxrP5Neh11ftq/nw7ooqXcJNSqf0VM7wFiGRk8J9Y=",
        version = "v0.0.0-20240213162025-012b6fc9bca9",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_rpc",
        importpath = "google.golang.org/genproto/googleapis/rpc",
        sum = "h1:NUsgEN92SQQqzfA+YtqYNqYmB3DMMYLlIwUZAQFVFbo=",
        version = "v0.0.0-20240221002015-b0ce06bbee7c",
    )
    go_repository(
        name = "org_golang_google_protobuf",
        importpath = "google.golang.org/protobuf",
        sum = "h1:pPC6BG5ex8PDFnkbrGU3EixyhKcQ2aDuBS36lqK/C7I=",
        version = "v1.32.0",
    )
    go_repository(
        name = "org_golang_x_crypto",
        importpath = "golang.org/x/crypto",
        sum = "h1:ENy+Az/9Y1vSrlrvBSyna3PITt4tiZLf7sgCjZBX7Wo=",
        version = "v0.19.0",
    )
    go_repository(
        name = "org_golang_x_exp",
        importpath = "golang.org/x/exp",
        sum = "h1:QE6XYQK6naiK1EPAe1g/ILLxN5RBoH5xkJk3CqlMI/Y=",
        version = "v0.0.0-20200224162631-6cc2880d07d6",
    )
    go_repository(
        name = "org_golang_x_image",
        importpath = "golang.org/x/image",
        sum = "h1:+qEpEAPhDZ1o0x3tHzZTQDArnOixOzGD9HUJfcg0mb4=",
        version = "v0.0.0-20190802002840-cff245a6509b",
    )
    go_repository(
        name = "org_golang_x_lint",
        importpath = "golang.org/x/lint",
        sum = "h1:VLliZ0d+/avPrXXH+OakdXhpJuEoBZuwh1m2j7U6Iug=",
        version = "v0.0.0-20210508222113-6edffad5e616",
    )
    go_repository(
        name = "org_golang_x_mobile",
        importpath = "golang.org/x/mobile",
        sum = "h1:OVJ6QQUBAesB8CZijKDSsXX7xYVtUhrkY0gwMfbi4p4=",
        version = "v0.0.0-20200801112145-973feb4309de",
    )
    go_repository(
        name = "org_golang_x_mod",
        importpath = "golang.org/x/mod",
        sum = "h1:dGoOF9QVLYng8IHTm7BAyWqCqSheQ5pYWGhzW00YJr0=",
        version = "v0.14.0",
    )
    go_repository(
        name = "org_golang_x_sync",
        importpath = "golang.org/x/sync",
        sum = "h1:5BMeUDZ7vkXGfEr1x9B4bRcTH4lpkTkpdh0T/J+qjbQ=",
        version = "v0.6.0",
    )
    go_repository(
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:25cE3gD+tdBA7lp7QfhuV+rJiE9YXTcS3VG1SqssI/Y=",
        version = "v0.17.0",
    )
    go_repository(
        name = "org_golang_x_term",
        importpath = "golang.org/x/term",
        sum = "h1:mkTF7LCd6WGJNL3K1Ad7kwxNfYAW6a8a8QqtMblp/4U=",
        version = "v0.17.0",
    )
    go_repository(
        name = "org_golang_x_time",
        importpath = "golang.org/x/time",
        sum = "h1:o7cqy6amK/52YcAKIPlM3a+Fpj35zvRj2TP+e1xFSfk=",
        version = "v0.5.0",
    )
    go_repository(
        name = "org_golang_x_tools",
        importpath = "golang.org/x/tools",
        sum = "h1:FvmRgNOcs3kOa+T20R1uhfP9F6HgG2mfxDv1vrx1Htc=",
        version = "v0.17.0",
    )
    go_repository(
        name = "org_golang_x_xerrors",
        importpath = "golang.org/x/xerrors",
        sum = "h1:H2TDz8ibqkAF6YGhCdN3jS9O0/s90v0rJh3X/OLHEUk=",
        version = "v0.0.0-20220907171357-04be3eba64a2",
    )
    go_repository(
        name = "org_uber_go_atomic",
        importpath = "go.uber.org/atomic",
        sum = "h1:ECmE8Bn/WFTYwEW/bpKD3M8VtR/zQVbavAoalC1PYyE=",
        version = "v1.9.0",
    )
    go_repository(
        name = "org_uber_go_goleak",
        importpath = "go.uber.org/goleak",
        sum = "h1:z+mqJhf6ss6BSfSM671tgKyZBFPTTJM+HLxnhPC3wu0=",
        version = "v1.1.10",
    )
    go_repository(
        name = "org_uber_go_multierr",
        importpath = "go.uber.org/multierr",
        sum = "h1:y6IPFStTAIT5Ytl7/XYmHvzXQ7S3g/IeZW9hyZ5thw4=",
        version = "v1.6.0",
    )
    go_repository(
        name = "org_uber_go_zap",
        importpath = "go.uber.org/zap",
        sum = "h1:CSUJ2mjFszzEWt4CdKISEuChVIXGBn3lAPwkRGyVrc4=",
        version = "v1.18.1",
    )
