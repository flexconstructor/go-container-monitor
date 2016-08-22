package container_monitor

import (
	"encoding/json"
	"errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"log"
	"math"
	"sort"
	"time"
)

// Collects system information, format this and marshal/unmarshal system info
// from/to JSON object.
type SystemInfoFactory struct {
	system_info *SystemInfo // System info value object.
}

// Returns new instance of system info factory.
func NewSystemInfoFactory() *SystemInfoFactory {
	return &SystemInfoFactory{
		system_info: &SystemInfo{},
	}
}

// Returns system information as JSON bytes.
func (f *SystemInfoFactory) GetSystemInfo() []byte {
	err := errors.New("System info error")
	log.Println("----------- Get System info -----------")
	f.system_info.CPUusage, err = f.getCPUPercent()

	if err != nil {
		f.system_info.CPUusage = 0.0
	} else {
		f.system_info.CPUusage = f.roundInterest64(f.system_info.CPUusage)
	}
	f.system_info.SWAPmemoryInfo = f.getSwapMemoryInfo()
	f.system_info.VirtualMemoryInfo = f.getVirtualMemoryInfo()
	log.Printf(
		"CPU: %v memory %v",
		f.system_info.CPUusage, f.system_info.VirtualMemoryInfo.UsedPercent)
	f.updateTop()
	bytes, err := json.Marshal(f.system_info)
	if err != nil {
		log.Printf("Can not marshal system info")
		return nil
	}
	return bytes
}

// Sets new system info value object from JSON
func (f *SystemInfoFactory) UnMarshalInfo(data []byte) (*SystemInfo, error) {
	err := json.Unmarshal(data, f.system_info)
	return f.system_info, err
}

// Returns total CPU usage in percents or error if the information is not found.
func (f *SystemInfoFactory) getCPUPercent() (float64, error) {
	arr, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0.0, err
	}
	var cpu_count float64 = 0.0
	var total_percent float64 = 0.0
	for _, cpu_percent := range arr {
		cpu_count += 1.0
		total_percent += cpu_percent
	}
	if cpu_count == 0.0 {
		return 0.0, errors.New("can not find cpu info")
	}
	return total_percent / cpu_count, nil
}

// Returns total swap memory usage info.
func (f *SystemInfoFactory) getSwapMemoryInfo() *MemoryInfo {
	info := &MemoryInfo{}
	swap, err := mem.SwapMemory()
	if err != nil {
		return info
	}
	info.Total = f.toMegaBytes(swap.Total)
	info.Available = f.toMegaBytes(swap.Free)
	info.Used = f.toMegaBytes(swap.Used)
	info.UsedPercent = f.roundInterest64(swap.UsedPercent)
	return info
}

// Returns total virtual memory usage info.
func (f *SystemInfoFactory) getVirtualMemoryInfo() *MemoryInfo {
	virtual_memory, err := mem.VirtualMemory()
	info := &MemoryInfo{}
	if err != nil {
		return info
	}
	info.Total = f.toMegaBytes(virtual_memory.Total)
	info.Available = f.toMegaBytes(virtual_memory.Available)
	info.Used = f.toMegaBytes(virtual_memory.Used)
	info.UsedPercent = f.roundInterest64(virtual_memory.UsedPercent)
	return info
}

// Casts bytes to megabytes.
func (f *SystemInfoFactory) toMegaBytes(value uint64) float64 {
	return math.Floor(float64(value)/(1024*1024)*100) / 100
}

// Rounds percents values.
//
// Returns float64
func (f *SystemInfoFactory) roundInterest64(value float64) float64 {
	return float64(math.Floor(value*100)) / 100
}

// Round percents values.
//
// Returns float32
func (f *SystemInfoFactory) roundInterest32(value float32) float32 {
	return float32(math.Floor(float64(value*100))) / 100
}

// Updates processes information.
func (f *SystemInfoFactory) updateTop() {
	pids, err := process.Pids()
	if err != nil {
		return
	}
	f.system_info.Top = make([]*ProcessInfo, len(pids))
	process_count := 0
	for _, pid := range pids {
		f.system_info.Top[process_count] = f.getProcessInfo(pid)
		if f.system_info.Top[process_count] != nil {
			log.Printf("- process PID: %d name: %s Memory usage: %v CPU: %v",
				f.system_info.Top[process_count].PID,
				f.system_info.Top[process_count].Name,
				f.system_info.Top[process_count].MemoryPersent,
				f.system_info.Top[process_count].CPUPersent)
		}
		process_count++
	}
	if f.system_info.Top != nil && len(f.system_info.Top) > 1 {
		sort.Sort(ByCPU(f.system_info.Top))
	}
}

// Returns process info value object.
func (f *SystemInfoFactory) getProcessInfo(pid int32) *ProcessInfo {
	process, err := process.NewProcess(pid)

	if err != nil {
		return nil
	}
	p_name, err := process.Name()

	if err != nil {
		p_name = err.Error()
	}
	status, err := process.Status()

	if err != nil {
		status = err.Error()
	}
	cwd, err := process.Cwd()

	if err != nil {
		cwd = err.Error()
	}
	createTime, err := process.CreateTime()

	if err != nil {
		createTime = 0
	}
	createTime = createTime / 1000
	memory_info, err := process.MemoryInfo()

	if err != nil {
		log.Printf("Can not get memory info for process %s", p_name)
	}
	memory_percent, err := process.MemoryPercent()

	if err != nil {
		memory_percent = 0.0
	} else {
		memory_percent = f.roundInterest32(memory_percent)
	}
	num_treads, err := process.NumThreads()

	if err != nil {
		num_treads = 0
	}
	cpu_percent, err := process.Percent(time.Second)

	if err != nil {
		log.Printf("can not get process %s percent of cpu")
		cpu_percent = 0.0
	} else {
		cpu_percent = f.roundInterest64(cpu_percent)
	}
	return &ProcessInfo{
		Name:          p_name,
		PID:           pid,
		Status:        status,
		Cwd:           cwd,
		CreateTime:    time.Unix(createTime, 0).Format("Jan 02, 2006 15:04:05"),
		MemoryInfo:    memory_info,
		MemoryPersent: memory_percent,
		NumThreads:    num_treads,
		CPUPersent:    cpu_percent,
	}
}
