package vcr

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/generator"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"

	"github.com/crossnokaye/goa-vcr/plugin/vcr/internal/vcrgen"
)

func init() {
	generator.RegisterPlugin("vcr", "gen", func() generator.Plugin {
		p := &plugin{}
		return generator.Plugin{Prepare: p.prepare, Generate: p.generate}
	})
}

// plugin holds the state of one generation run. Goa's generator.Plan does not
// expose its roots, so prepare records them for generate.
type plugin struct {
	genpkg string
	roots  []eval.Root
}

func (p *plugin) prepare(genpkg string, roots []eval.Root) error {
	p.genpkg = genpkg
	p.roots = roots
	return nil
}

// generate appends the VCR glue for every HTTP service in every Goa root.
func (p *plugin) generate(plan *generator.Plan, files []*codegen.File) ([]*codegen.File, error) {
	if os.Getenv("GOA_VCR_DEBUG") != "" {
		files = append(files, debugFile(p.genpkg, p.roots))
	}

	seen := make(map[string]struct{})
	for _, r := range p.roots {
		root, ok := r.(*expr.RootExpr)
		if !ok || root.API == nil || root.API.HTTP == nil {
			continue
		}
		httpPlan, ok := plan.HTTP(root)
		if !ok {
			continue
		}

		// Deterministic service ordering per root.
		services := append([]*expr.HTTPServiceExpr(nil), root.API.HTTP.Services...)
		sort.Slice(services, func(i, j int) bool { return services[i].Name() < services[j].Name() })

		for _, hs := range services {
			name := hs.Name()
			if _, ok := seen[name]; ok {
				continue
			}
			svc, ok := httpPlan.Service(hs)
			if !ok || svc.Service == nil {
				continue
			}
			spec := vcrgen.BuildServiceSpec(p.genpkg, svc)
			f := vcrgen.RenderServiceVCR(spec)
			cli := vcrgen.RenderServiceVCRCLI(spec)
			// Goa retains imports per file, so borrow the service package's own
			// imports: they cover user types placed in other packages via
			// "struct:pkg:path" and meta types via "struct:field:type". Goa drops
			// the ones the VCR file does not use.
			addServiceImports(f, files, svc.Service.PathName)
			files = append(files, f, cli)
			seen[name] = struct{}{}
		}
	}
	return files, nil
}

func addServiceImports(f *codegen.File, files []*codegen.File, servicePath string) {
	if len(f.SectionTemplates) == 0 {
		return
	}
	svcFile := filepath.Join(codegen.Gendir, servicePath, "service.go")
	for _, src := range files {
		if src.Path != svcFile || len(src.SectionTemplates) == 0 {
			continue
		}
		if data, ok := src.SectionTemplates[0].Data.(map[string]any); ok {
			if imports, ok := data["Imports"].([]*codegen.ImportSpec); ok {
				codegen.AddImport(f.SectionTemplates[0], imports...)
			}
		}
	}
}

func debugFile(genpkg string, roots []eval.Root) *codegen.File {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "goa-vcr plugin: genpkg=%s\n", genpkg)
	_, _ = fmt.Fprintf(&b, "roots=%d\n", len(roots))
	for i, r := range roots {
		if rr, ok := r.(*expr.RootExpr); ok {
			hasHTTP := rr.API != nil && rr.API.HTTP != nil
			_, _ = fmt.Fprintf(&b, "%d: root=%s services=%d hasHTTP=%v\n", i, rr.EvalName(), len(rr.Services), hasHTTP)
			for _, s := range rr.Services {
				_, _ = fmt.Fprintf(&b, "  - %s\n", s.Name)
			}
		} else {
			_, _ = fmt.Fprintf(&b, "%d: root=%T\n", i, r)
		}
	}
	return &codegen.File{
		Path: filepath.Join(codegen.Gendir, "vcr_debug.txt"),
		SectionTemplates: []*codegen.SectionTemplate{
			{Name: "debug", Source: b.String()},
		},
	}
}
