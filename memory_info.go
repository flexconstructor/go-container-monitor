package container_monitor

// Memory usage value object.
type MemoryInfo struct {
	Total       float64 // Total memory bytes
	Used        float64 // Used memory bytes
	Available   float64 // Available memory bytes.
	UsedPercent float64 // Used memory in percents.
}
