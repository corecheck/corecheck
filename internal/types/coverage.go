package types

type CoverageLine struct {
	File               string
	OriginalLineNumber int
	NewLineNumber      int
}

type DifferentialCoverage struct {
	// Newly added code is not tested
	UncoveredNewCode []CoverageLine
	// Preexisting code is no longer tested
	LostBaselineCoverage []CoverageLine
	// Previously unused code is not covered
	UncoveredIncludedCode []CoverageLine
	// Unchanged code is covered now
	GainedBaselineCoverage []CoverageLine
	// Previously unused code is covered now
	GainedCoverageIncludedCode []CoverageLine
	// Newly added code is exercised
	GainedCoverageNewCode []CoverageLine
	// Previously un-exercised code is unused now.
	ExcludedUncoveredBaselineCode []CoverageLine
	// Previously exercised code is unused now.
	ExcludedCoveredBaselineCode []CoverageLine
	// Previously un-exercised code has been deleted.
	DeletedUncoveredBaselineCode []CoverageLine
	// Previously exercised code has been deleted
	DeletedCoveredBaselineCode []CoverageLine
}
