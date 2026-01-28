package stdlib

import (
	"sync"

	"github.com/FrameworkOSS/portal/features/commands"
	"github.com/FrameworkOSS/portal/features/debugger"
	"github.com/FrameworkOSS/portal/features/files"
	"github.com/FrameworkOSS/portal/features/hellodolly"
	"github.com/FrameworkOSS/portal/features/shell"
	"github.com/FrameworkOSS/portal/features/wires"
	"github.com/FrameworkOSS/portal/portal"
)

var features = []portal.Feature{
	hellodolly.NewHelloDolly(),
	files.NewFiles(),
	wires.NewWires(),
}

// NewPortal is a convenient wrapper for importing the standard library.
// It is useful for prototyping but is not guaranteed to stay! Please switch to a manual stdlib import before switching to production.
func NewPortal(opts *portal.PortalOptions, cmds, shell, debug bool) *portal.Portal {
	p := portal.NewPortal(opts)
	p.FeatureAdd(NewStdlib(p, cmds, shell, debug))
	return p
}

// Stdlib is used as a way to transparently load the rest of the standard portal features if none replace them.
// Must be imported before anything which requires it in case of runtime dependency checks during Init.
type Stdlib struct {
	p       *portal.Portal
	c, s, d bool

	lockResp sync.Mutex
	resps    []*portal.Event
}

func NewStdlib(p *portal.Portal, c, s, d bool) *Stdlib {
	f := new(Stdlib)
	f.p = p
	f.c = c
	f.s = s
	f.d = d

	if f.c {
		cmds := commands.NewCommands(f.p)
		if err := f.p.FeatureAdd(cmds); err != nil {
			panic(err)
		}
	}
	if f.s {
		sh := shell.NewShell(f.p, "", false, !f.c)
		if err := f.p.FeatureAdd(sh); err != nil {
			panic(err)
		}
	}
	if f.d {
		dbg := debugger.NewDebugger(f.p)
		if err := f.p.FeatureAdd(dbg); err != nil {
			panic(err)
		}
		f.p.OptionsGet().SetDebugger(dbg.ID())
	}
	if err := f.p.BatchFeatureAdd(features...); err != nil {
		panic(err)
	}
	return f
}

func (f *Stdlib) API() int {
	return 0
}

func (f *Stdlib) ID() string {
	return "portal-stdlib"
}

func (f *Stdlib) Name() string {
	return "Portal Standard Features Library"
}

func (f *Stdlib) Authors() []string {
	return []string{"JoshuaDoes"}
}

func (f *Stdlib) Description() string {
	return "Transparently loads missing standard portal features if none replace them."
}

func (f *Stdlib) Version() string {
	return "v0.0.1"
}

func (f *Stdlib) Open() error {
	f.storeResp(portal.NewEventReady(f.ID(), true))
	return nil
}

func (f *Stdlib) Close() (errs []error, retry bool) {
	f.p = nil
	return
}

func (f *Stdlib) Input(_ *portal.Event) error {
	return nil
}

func (f *Stdlib) Output() (*portal.Event, error) {
	return f.readResp(), nil
}

func (f *Stdlib) storeResp(e *portal.Event) {
	f.lockResp.Lock()
	f.resps = append(f.resps, e)
	f.lockResp.Unlock()
}

func (f *Stdlib) readResp() (e *portal.Event) {
	if len(f.resps) > 0 {
		f.lockResp.Lock()
		e = f.resps[0]
		f.resps = f.resps[1:]
		f.lockResp.Unlock()
	}
	return
}
