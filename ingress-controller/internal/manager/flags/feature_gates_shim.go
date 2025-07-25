package flags

import (
	"github.com/samber/lo"
	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"

	managercfg "github.com/kong/kong-operator/ingress-controller/pkg/manager/config"
)

// NewMapStringBoolForFeatureGatesWithDefaults takes a pointer to a FeatureGates (map[string]bool) and returns the
// MapStringBool flag parsing shim for that map which populates it with defaults feature gates in case of missing keys.
func NewMapStringBoolForFeatureGatesWithDefaults(m *managercfg.FeatureGates) *FeatureGatesVar {
	*m = lo.Must(managercfg.NewFeatureGates(nil)) // For nil it never returns an error.
	return &FeatureGatesVar{fg: m}
}

type FeatureGatesVar struct {
	fg *managercfg.FeatureGates
}

var _ pflag.Value = &FeatureGatesVar{}

func (f *FeatureGatesVar) Set(value string) error {
	tmp := make(map[string]bool)
	if err := flag.NewMapStringBool(&tmp).Set(value); err != nil {
		return err
	}
	var err error
	*f.fg, err = managercfg.NewFeatureGates(tmp)
	if err != nil {
		return err
	}
	return nil
}

func (f *FeatureGatesVar) Type() string {
	return "mapStringBool"
}

func (f *FeatureGatesVar) String() string {
	return ""
}
