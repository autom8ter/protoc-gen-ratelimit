package module

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/autom8ter/proto/gen/ratelimit"
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

// Module is the protoc-gen-ratelimitr module
// implements the protoc-gen-star module interface
type module struct {
	*pgs.ModuleBase
	pgsgo.Context
	limiter string
}

func New() pgs.Module {
	return &module{ModuleBase: &pgs.ModuleBase{}}
}

func (m *module) Name() string {
	return "ratelimit"
}

func (m *module) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.Context = pgsgo.InitContext(c.Parameters())
	m.limiter = c.Parameters().Str("limiter")
}

func (m *module) Execute(targets map[string]pgs.File, packages map[string]pgs.Package) []pgs.Artifact {
	for _, f := range targets {
		if f.BuildTarget() {
			m.generate(f)
		}
	}
	return m.Artifacts()
}

func (m *module) generate(f pgs.File) {
	var optsMap = map[string][]*ratelimit.RateLimitOptions{}
	for _, s := range f.Services() {
		for _, method := range s.Methods() {
			var opts []*ratelimit.RateLimitOptions
			ok, err := method.Extension(ratelimit.E_Options, &opts)
			if err != nil {
				m.AddError(err.Error())
				continue
			}
			if !ok {
				continue
			}
			name := fmt.Sprintf("%s_%s_FullMethodName", s.Name().UpperCamelCase(), method.Name().UpperCamelCase())
			optsMap[name] = opts
		}
	}
	if len(optsMap) == 0 {
		return
	}
	name := f.InputPath().SetExt(".pb.ratelimit.go").String()
	var (
		t   *template.Template
		err error
	)
	switch m.limiter {
	case "redis":
		t, err = template.New("ratelimit").Parse(redisTmpl)
	case "inmem":
		t, err = template.New("ratelimit").Parse(inMemTmpl)
	default:
		m.AddError("ratelimit: invalid limiter option(must be one of: redis, inmem)")
	}
	if err != nil {
		m.AddError(fmt.Sprintf("ratelimit: failed to parse template: %v", err))
		return
	}
	buffer := &bytes.Buffer{}
	if err := t.Execute(buffer, templateData{
		Package: m.Context.PackageName(f).String(),
		Opts:    optsMap,
	}); err != nil {
		m.AddError(err.Error())
		return
	}
	m.AddGeneratorFile(name, buffer.String())
}

type templateData struct {
	Package string
	Opts    map[string][]*ratelimit.RateLimitOptions
}

var redisTmpl = `
package {{ .Package }}

import (
	"github.com/autom8ter/proto/gen/ratelimit"

	"github.com/redis/go-redis/v9"

	"github.com/autom8ter/protoc-gen-ratelimit/limiter"
	redis_limiter "github.com/autom8ter/protoc-gen-ratelimit/redis"
)


// NewRateLimiter returns a new ratelimiter using the provided redis client
func NewRateLimiter(client *redis.Client) (limiter.Limiter, error) {
	limit := redis_limiter.NewLimiter(client, map[string][]*ratelimit.RateLimitOptions{
	{{- range $key, $value := .Opts }}
	{{$key}}: {
		{{- range $value }}
		{
			Limit: {{ .Limit }},
			MetadataKey: "{{ .MetadataKey }}",
			Message: "{{ .Message }}",
		},
		{{- end }}
	},
	{{- end }}
	})
	return limit.Limit, nil
}
`

var inMemTmpl = `
package {{ .Package }}

import (
	"github.com/autom8ter/proto/gen/ratelimit"

	"github.com/autom8ter/protoc-gen-ratelimit/limiter"
	"github.com/autom8ter/protoc-gen-ratelimit/inmem"
)


// NewRateLimiter returns a new inmemory ratelimiter
func NewRateLimiter() (limiter.Limiter, error) {
	limit := inmem.NewLimiter(map[string][]*ratelimit.RateLimitOptions{
	{{- range $key, $value := .Opts }}
	{{$key}}: {
		{{- range $value }}
		{
			Limit: {{ .Limit }},
			MetadataKey: "{{ .MetadataKey }}",
			Message: "{{ .Message }}",
		},
		{{- end }}
	},
	{{- end }}
	})
	return limit.Limit, nil
}
`
