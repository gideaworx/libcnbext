package libcnbext

type ExtensionProvides struct {
	Name string `toml:"name"`
}

type DetectResult struct {
	Provides []ExtensionProvides `toml:"provides"`
}

type DetectFunc func() (DetectResult, error)
