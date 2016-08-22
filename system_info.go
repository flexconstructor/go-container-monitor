package container_monitor

// System info value object.
type SystemInfo struct {
	CPUusage          float64        // Total CPU usage info.
	VirtualMemoryInfo *MemoryInfo    // Total virtual memory usage info.
	SWAPmemoryInfo    *MemoryInfo    // Total swap memory usage info.
	Top               []*ProcessInfo // Processes info array.
}

// Returns new system info value object.
func NewSystemInfo() *SystemInfo {
	return &SystemInfo{
		CPUusage: 0.0,
	}
}
