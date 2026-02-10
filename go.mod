module github.com/Permify/permify

go 1.25.7

tool (
	github.com/bufbuild/buf/cmd/buf
	github.com/envoyproxy/protoc-gen-validate
	github.com/golangci/golangci-lint/v2/cmd/golangci-lint
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto
	github.com/securego/gosec/v2/cmd/gosec
	google.golang.org/grpc/cmd/protoc-gen-go-grpc
	google.golang.org/protobuf/cmd/protoc-gen-go
	mvdan.cc/gofumpt
)

//
require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/Permify/sloggcp v0.0.3
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/dgraph-io/ristretto v1.0.0
	github.com/dustin/go-humanize v1.0.1
	github.com/envoyproxy/protoc-gen-validate v1.3.0
	github.com/exaring/otelpgx v0.9.3
	github.com/fatih/color v1.18.0
	github.com/go-jose/go-jose/v3 v3.0.4
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/google/cel-go v0.26.1
	github.com/gookit/color v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/hashicorp/go-memdb v1.3.5
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-retryablehttp v0.7.8
	github.com/jackc/pgio v1.0.0
	github.com/jackc/pgtype v1.14.4
	github.com/jackc/pgx/v5 v5.8.0
	github.com/juju/ratelimit v1.0.2
	github.com/lestrrat-go/jwx v1.2.31
	github.com/onsi/ginkgo/v2 v2.28.1
	github.com/onsi/gomega v1.39.0
	github.com/pkg/errors v0.9.1
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10
	github.com/pressly/goose/v3 v3.26.0
	github.com/rs/cors v1.11.1
	github.com/rs/xid v1.6.0
	github.com/sercand/kuberesolver/v5 v5.1.1
	github.com/sony/gobreaker v1.0.0
	github.com/spf13/cobra v1.10.1
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/testcontainers/testcontainers-go v0.40.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.64.0
	go.opentelemetry.io/contrib/instrumentation/host v0.63.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.63.0
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.39.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.39.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
	go.opentelemetry.io/otel/exporters/zipkin v1.38.0
	go.opentelemetry.io/otel/metric v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
	go.opentelemetry.io/otel/sdk/metric v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
	golang.org/x/exp v0.0.0-20251002181428-27f1f14c8bb9
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0
	google.golang.org/genproto/googleapis/api v0.0.0-20251202230838-ff82c1b0f217
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
	gopkg.in/yaml.v3 v3.0.1
	resenje.org/singleflight v0.4.3
)

//
require (
	4d63.com/gocheckcompilerdirectives v1.3.0 // indirect
	4d63.com/gochecknoglobals v0.2.2 // indirect
	buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go v1.36.10-20250718181942-e35f9b667443.1 // indirect
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250717185734-6c6e0d3c608e.1 // indirect
	buf.build/gen/go/bufbuild/registry/connectrpc/go v1.19.0-20250924144421-cb55f06efbd2.1 // indirect
	buf.build/gen/go/bufbuild/registry/protocolbuffers/go v1.36.10-20250924144421-cb55f06efbd2.1 // indirect
	buf.build/gen/go/pluginrpc/pluginrpc/protocolbuffers/go v1.36.9-20241007202033-cf42259fcbfc.1 // indirect
	buf.build/go/app v0.1.0 // indirect
	buf.build/go/bufplugin v0.9.0 // indirect
	buf.build/go/interrupt v1.1.0 // indirect
	buf.build/go/protovalidate v1.0.0 // indirect
	buf.build/go/protoyaml v0.6.0 // indirect
	buf.build/go/spdx v0.2.0 // indirect
	buf.build/go/standard v0.1.0 // indirect
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.121.2 // indirect
	cloud.google.com/go/ai v0.12.1 // indirect
	cloud.google.com/go/auth v0.16.5 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/logging v1.13.0 // indirect
	cloud.google.com/go/longrunning v0.6.7 // indirect
	cloud.google.com/go/monitoring v1.24.2 // indirect
	cloud.google.com/go/trace v1.11.6 // indirect
	codeberg.org/chavacava/garif v0.2.0 // indirect
	connectrpc.com/connect v1.19.1 // indirect
	connectrpc.com/otelconnect v0.8.0 // indirect
	dev.gaijin.team/go/exhaustruct/v4 v4.0.0 // indirect
	dev.gaijin.team/go/golib v0.6.0 // indirect
	github.com/4meepo/tagalign v1.4.3 // indirect
	github.com/Abirdcfly/dupword v0.1.6 // indirect
	github.com/AdminBenni/iota-mixing v1.0.0 // indirect
	github.com/AlwxSin/noinlineerr v1.0.5 // indirect
	github.com/Antonboom/errname v1.1.1 // indirect
	github.com/Antonboom/nilnil v1.1.1 // indirect
	github.com/Antonboom/testifylint v1.6.4 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/Djarvur/go-err113 v0.1.1 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.55.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/MirrexOne/unqueryvet v1.2.1 // indirect
	github.com/OpenPeeDeeP/depguard/v2 v2.2.1 // indirect
	github.com/alecthomas/chroma/v2 v2.20.0 // indirect
	github.com/alecthomas/go-check-sumtype v0.3.1 // indirect
	github.com/alexkohler/nakedret/v2 v2.0.6 // indirect
	github.com/alexkohler/prealloc v1.0.0 // indirect
	github.com/alfatraining/structtag v1.0.0 // indirect
	github.com/alingse/asasalint v0.0.11 // indirect
	github.com/alingse/nilnesserr v0.2.0 // indirect
	github.com/ashanbrown/forbidigo/v2 v2.1.0 // indirect
	github.com/ashanbrown/makezero/v2 v2.0.1 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/bkielbasa/cyclop v1.2.3 // indirect
	github.com/blizzy78/varnamelen v0.8.0 // indirect
	github.com/bombsimon/wsl/v4 v4.7.0 // indirect
	github.com/bombsimon/wsl/v5 v5.2.0 // indirect
	github.com/breml/bidichk v0.3.3 // indirect
	github.com/breml/errchkjson v0.4.1 // indirect
	github.com/bufbuild/buf v1.58.0 // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/bufbuild/protoplugin v0.0.0-20250218205857-750e09ce93e1 // indirect
	github.com/butuzov/ireturn v0.4.0 // indirect
	github.com/butuzov/mirror v1.3.0 // indirect
	github.com/catenacyber/perfsprint v0.9.1 // indirect
	github.com/ccojocar/zxcvbn-go v1.0.4 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/charithe/durationcheck v0.0.10 // indirect
	github.com/charmbracelet/colorprofile v0.2.3-0.20250311203215-f60798e515dc // indirect
	github.com/charmbracelet/lipgloss v1.1.0 // indirect
	github.com/charmbracelet/x/ansi v0.8.0 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13-0.20250311204145-2c3ea96c31dd // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/ckaznocha/intrange v0.3.1 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.17.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/curioswitch/go-reassign v0.3.0 // indirect
	github.com/daixiang0/gci v0.13.7 // indirect
	github.com/dave/dst v0.27.3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/denis-tingaikin/go-header v0.5.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/docker/cli v28.5.0+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.9.4 // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/firefart/nonamedreturns v1.0.6 // indirect
	github.com/fzipp/gocyclo v0.6.0 // indirect
	github.com/ghostiam/protogetter v0.3.16 // indirect
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	github.com/go-critic/go-critic v0.13.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-toolsmith/astcast v1.1.0 // indirect
	github.com/go-toolsmith/astcopy v1.1.0 // indirect
	github.com/go-toolsmith/astequal v1.2.0 // indirect
	github.com/go-toolsmith/astfmt v1.1.0 // indirect
	github.com/go-toolsmith/astp v1.1.0 // indirect
	github.com/go-toolsmith/strparse v1.1.0 // indirect
	github.com/go-toolsmith/typep v1.1.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/go-xmlfmt/xmlfmt v1.1.3 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/godoc-lint/godoc-lint v0.10.0 // indirect
	github.com/gofrs/flock v0.12.1 // indirect
	github.com/golangci/asciicheck v0.5.0 // indirect
	github.com/golangci/dupl v0.0.0-20250308024227-f665c8d69b32 // indirect
	github.com/golangci/go-printf-func-name v0.1.1 // indirect
	github.com/golangci/gofmt v0.0.0-20250106114630-d62b90e6713d // indirect
	github.com/golangci/golangci-lint/v2 v2.5.0 // indirect
	github.com/golangci/golines v0.0.0-20250217134842-442fd0091d95 // indirect
	github.com/golangci/misspell v0.7.0 // indirect
	github.com/golangci/nilerr v0.0.0-20250918000102-015671e622fe // indirect
	github.com/golangci/plugin-module-register v0.1.2 // indirect
	github.com/golangci/revgrep v0.8.0 // indirect
	github.com/golangci/swaggoswag v0.0.0-20250504205917-77f2aca3143e // indirect
	github.com/golangci/unconvert v0.0.0-20250410112200-a129a6e6413e // indirect
	github.com/google/generative-ai-go v0.20.1 // indirect
	github.com/google/go-containerregistry v0.20.6 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/gordonklaus/ineffassign v0.2.0 // indirect
	github.com/gostaticanalysis/analysisutil v0.7.1 // indirect
	github.com/gostaticanalysis/comment v1.5.0 // indirect
	github.com/gostaticanalysis/forcetypeassert v0.2.0 // indirect
	github.com/hashicorp/go-immutable-radix/v2 v2.1.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/jdx/go-netrc v1.0.0 // indirect
	github.com/jgautheron/goconst v1.8.2 // indirect
	github.com/jingyugao/rowserrcheck v1.1.1 // indirect
	github.com/jjti/go-spancheck v0.6.5 // indirect
	github.com/julz/importas v0.2.0 // indirect
	github.com/karamaru-alpha/copyloopvar v1.2.1 // indirect
	github.com/kisielk/errcheck v1.9.0 // indirect
	github.com/kkHAIKE/contextcheck v1.1.6 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/kulti/thelper v0.7.1 // indirect
	github.com/kunwardeep/paralleltest v1.0.14 // indirect
	github.com/lasiar/canonicalheader v1.1.2 // indirect
	github.com/ldez/exptostd v0.4.4 // indirect
	github.com/ldez/gomoddirectives v0.7.0 // indirect
	github.com/ldez/grignotin v0.10.1 // indirect
	github.com/ldez/tagliatelle v0.7.2 // indirect
	github.com/ldez/usetesting v0.5.0 // indirect
	github.com/leonklingele/grouper v1.1.2 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/lyft/protoc-gen-star/v2 v2.0.4-0.20230330145011-496ad1ac90a4 // indirect
	github.com/macabu/inamedparam v0.2.0 // indirect
	github.com/manuelarte/embeddedstructfieldcheck v0.4.0 // indirect
	github.com/manuelarte/funcorder v0.5.0 // indirect
	github.com/maratori/testableexamples v1.0.0 // indirect
	github.com/maratori/testpackage v1.1.1 // indirect
	github.com/matoous/godox v1.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mgechev/revive v1.12.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moricho/tparallel v0.3.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nakabonne/nestif v0.3.1 // indirect
	github.com/nishanths/exhaustive v0.12.0 // indirect
	github.com/nishanths/predeclared v0.2.2 // indirect
	github.com/nunnatsa/ginkgolinter v0.21.0 // indirect
	github.com/olekukonko/errors v1.1.0 // indirect
	github.com/olekukonko/ll v0.1.3 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/polyfloyd/go-errorlint v1.8.0 // indirect
	github.com/quasilyte/go-ruleguard v0.4.4 // indirect
	github.com/quasilyte/go-ruleguard/dsl v0.3.22 // indirect
	github.com/quasilyte/gogrep v0.5.0 // indirect
	github.com/quasilyte/regex/syntax v0.0.0-20210819130434-b3f0c404a727 // indirect
	github.com/quasilyte/stdinfo v0.0.0-20220114132959-f7386bf02567 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.57.0 // indirect
	github.com/raeperd/recvcheck v0.2.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryancurrah/gomodguard v1.4.1 // indirect
	github.com/ryanrolds/sqlclosecheck v0.5.1 // indirect
	github.com/samber/lo v1.52.0 // indirect
	github.com/samber/slog-common v0.19.0 // indirect
	github.com/sanposhiho/wastedassign/v2 v2.1.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/sashamelentyev/interfacebloat v1.1.0 // indirect
	github.com/sashamelentyev/usestdlibvars v1.29.0 // indirect
	github.com/securego/gosec/v2 v2.22.8 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/segmentio/encoding v0.5.3 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/shirou/gopsutil/v4 v4.25.8 // indirect
	github.com/sivchari/containedctx v1.0.3 // indirect
	github.com/sonatard/noctx v0.4.0 // indirect
	github.com/sourcegraph/go-diff v0.7.0 // indirect
	github.com/ssgreg/nlreturn/v2 v2.2.1 // indirect
	github.com/stbenjam/no-sprintf-host-port v0.2.0 // indirect
	github.com/tetafro/godot v1.5.4 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/timakin/bodyclose v0.0.0-20241222091800-1db5c5ca4d67 // indirect
	github.com/timonwong/loggercheck v0.11.0 // indirect
	github.com/tomarrell/wrapcheck/v2 v2.11.0 // indirect
	github.com/tommy-muehle/go-mnd/v2 v2.5.1 // indirect
	github.com/ultraware/funlen v0.2.0 // indirect
	github.com/ultraware/whitespace v0.2.0 // indirect
	github.com/uudashr/gocognit v1.2.0 // indirect
	github.com/uudashr/iface v1.4.1 // indirect
	github.com/vbatts/tar-split v0.12.1 // indirect
	github.com/xen0n/gosmopolitan v1.3.0 // indirect
	github.com/yagipy/maintidx v1.0.0 // indirect
	github.com/yeya24/promlinter v0.3.0 // indirect
	github.com/ykadowak/zerologlint v0.1.5 // indirect
	gitlab.com/bosi/decorder v0.4.2 // indirect
	go-simpler.org/musttag v0.14.0 // indirect
	go-simpler.org/sloglint v0.11.1 // indirect
	go.augendre.info/arangolint v0.2.0 // indirect
	go.augendre.info/fatcontext v0.8.1 // indirect
	go.lsp.dev/jsonrpc2 v0.10.0 // indirect
	go.lsp.dev/pkg v0.0.0-20210717090340-384b27a52fb2 // indirect
	go.lsp.dev/protocol v0.12.0 // indirect
	go.lsp.dev/uri v0.3.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp/typeparams v0.0.0-20250911091902-df9299821621 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/oauth2 v0.32.0 // indirect
	golang.org/x/term v0.39.0 // indirect
	golang.org/x/time v0.13.0 // indirect
	google.golang.org/api v0.249.0 // indirect
	google.golang.org/genproto v0.0.0-20250922171735-9219d122eba9 // indirect
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.5.1 // indirect
	honnef.co/go/tools v0.6.1 // indirect
	mvdan.cc/gofumpt v0.9.1 // indirect
	mvdan.cc/unparam v0.0.0-20250301125049-0df0534333a4 // indirect
	pluginrpc.com/pluginrpc v0.5.0 // indirect
)

//
require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.55.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.30.0
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/agoda-com/opentelemetry-go/otelslog v0.3.0
	github.com/agoda-com/opentelemetry-logs-go v0.6.0
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/docker v28.5.1+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260115054156-294ebfa9ad83 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250827001030-24949be3fa54 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/olekukonko/tablewriter v1.1.2
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.20.3 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.59.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/samber/slog-multi v1.6.0
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10
	github.com/stoewer/go-strcase v1.3.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)

require (
	github.com/clipperhouse/displaywidth v0.6.0 // indirect
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.3.0 // indirect
	github.com/olekukonko/cat v0.0.0-20250911104152-50322a0618f6 // indirect
)
