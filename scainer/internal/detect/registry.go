package detect

type Registry struct{ stages []Stage }

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Add(s Stage) { r.stages = append(r.stages, s) }

func (r *Registry) Stages() []Stage { return r.stages }
