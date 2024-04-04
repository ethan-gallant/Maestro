package reconciler

type DryRunType string

const (
	// DryRunWarn will attempt a dry-run on a mismatch and log a warning if the object is identical after the dry-run (default)
	DryRunWarn DryRunType = "warn"
	// DryRunSilent will perform a dry-run and not log anything if the object is identical after the dry-run
	DryRunSilent DryRunType = "silent"
	// DryRunNone will not perform a dry-run, and will always update the object if it is different
	DryRunNone DryRunType = "none"
)
