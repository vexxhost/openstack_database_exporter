package nova

// DefaultQuotas holds configurable Nova quota defaults used as fallback
// when no explicit per-project quota and no quota_classes entry exists in the DB.
type DefaultQuotas struct {
	Instances   int
	Cores       int
	PinnedCores int
	RAM         int
}

// DefaultDefaultQuotas returns the Nova upstream defaults.
func DefaultDefaultQuotas() DefaultQuotas {
	return DefaultQuotas{
		Instances:   10,
		Cores:       20,
		PinnedCores: 0,
		RAM:         51200,
	}
}

// unifiedLimitToNova maps Keystone unified limit resource names to Nova quota
// resource names. Returns false for resources not relevant to Nova limits.
func unifiedLimitToNova(resource string) (string, bool) {
	switch resource {
	case "servers":
		return "instances", true
	case "class:VCPU":
		return "cores", true
	case "class:PCPU":
		return "pcpus", true
	case "class:MEMORY_MB":
		return "ram", true
	default:
		return "", false
	}
}
