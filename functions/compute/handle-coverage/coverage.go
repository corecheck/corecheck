package main

type RawCoverageData struct {
	Files []struct {
		File      string `json:"file"`
		Functions []struct {
			ExecutionCount int    `json:"execution_count"`
			Lineno         int    `json:"lineno"`
			Name           string `json:"name"`
		} `json:"functions"`
		Lines []struct {
			Branches   []any `json:"branches"`
			Count      int   `json:"count"`
			LineNumber int   `json:"line_number"`
		} `json:"lines"`
	} `json:"files"`
}

type RawLineCoverage struct {
	LineNumber int
	Count      int
}

type CoverageMap map[string]map[int]RawLineCoverage

func (c RawCoverageData) ToMap() CoverageMap {
	m := make(CoverageMap)

	for _, file := range c.Files {
		m[file.File] = make(map[int]RawLineCoverage)
		for _, l := range file.Lines {
			m[file.File][l.LineNumber] = RawLineCoverage{
				LineNumber: l.LineNumber,
				Count:      l.Count,
			}
		}
	}

	return m
}

func (c CoverageMap) ListFiles() []string {
	var files []string

	for file := range c {
		files = append(files, file)
	}

	return files
}

func (c CoverageMap) IsTested(filename string, line int) bool {
	if _, ok := c[filename]; !ok {
		return false
	}

	if _, ok := c[filename][line]; !ok {
		return false
	}

	return true
}

func (c CoverageMap) IsCovered(filename string, line int) bool {
	if !c.IsTested(filename, line) {
		return false
	}

	return c[filename][line].Count > 0
}
