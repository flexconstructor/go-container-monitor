package container_monitor

// Sorts process info array by CPU usage and memory usage.
type ByCPU []*ProcessInfo

// Returns length of sortable array.
func (b ByCPU) Len() int {
	return len(b)
}

// Swaps array indexes.
func (b ByCPU) Swap(i, j int) {
	if b[i] == nil || b[j] == nil {
		return
	}
	b[i], b[j] = b[j], b[i]
}

// Returns sort rule.
func (b ByCPU) Less(i, j int) bool {
	if b[i] == nil || b[j] == nil {
		return false
	}
	if b[i].CPUPercent > 0 || b[j].CPUPercent > 0 {
		return b[i].CPUPercent > b[j].CPUPercent
	}
	return b[i].MemoryPercent > b[j].MemoryPercent
}
