package container_monitor

import (
	"errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"log"
	"math"
	"sort"
	"time"
	"gopkg.in/redis.v4"
	"fmt"
	"strconv"
	"encoding/json"
)

// Collects system information, format this and marshal/unmarshal system info
// from/to JSON object.
type SystemInfoFactory struct {
	system_info *SystemInfo // System info value object.
	redis_client *redis.Client // Redis client instance.
}

// Returns new instance of system info factory.
//
// param: The instance of Redis client.
func NewSystemInfoFactory(client *redis.Client) *SystemInfoFactory {
	return &SystemInfoFactory{
		system_info: NewSystemInfo(),
		redis_client:client,
	}
}

// Writes system information to redis db.
//
// param: test_id string   ID of current test.
func(f *SystemInfoFactory)UpdateSystemInfo(test_id string){
	pref:= "system:"+test_id
	err:=f.redis_client.Incr(pref+":steps").Err()
	if(err != nil){
		log.Printf("can not increment steps: %s",err.Error())
	}

	cpu,err:= f.getCPUPercent()
	if(err != nil){
		log.Printf("can not get cpu: %s",err.Error())
		cpu=0.0
	}

	err= f.redis_client.IncrByFloat(pref+":cpu",cpu).Err()
	if(err != nil){
		log.Printf("can not increment cpu: %s",err.Error())
	}

 	swap, err := mem.SwapMemory()
	if(err != nil){
		log.Printf("can not get swap: %s",err.Error())
	}

	err= f.redis_client.HIncrByFloat(
		pref+":swap","percent",swap.UsedPercent).Err()
	if(err != nil){
		log.Printf("can not increment swap percent: %s",err.Error())
	}

	err= f.redis_client.HIncrBy(pref+":swap","total",int64(swap.Total)).Err()
	if(err != nil){
		log.Printf("can not increment swap total: %s",err.Error())
	}

	err= f.redis_client.HIncrBy(pref+":swap","used",int64(swap.Used)).Err()
	if(err != nil){
		log.Printf("can not increment swap used: %s",err.Error())
	}

	err= f.redis_client.HIncrBy(pref+":swap","available",int64(swap.Free)).Err()
	if(err != nil){
		log.Printf("can not increment swap available: %s",err.Error())
	}

	virtual_memory, err := mem.VirtualMemory()
	if(err != nil){
		log.Printf("can not get virtual memory: %s",err.Error())
	}

	err= f.redis_client.HIncrByFloat(
		pref+":vm","percent",virtual_memory.UsedPercent).Err()
	if(err != nil){
		log.Printf("can not increment vm percent: %s",err.Error())
	}

	err= f.redis_client.HIncrBy(pref+":vm","total",int64(
		virtual_memory.Total)).Err()
	if(err != nil){
		log.Printf("can not increment vm total: %s",err.Error())
	}

	err= f.redis_client.HIncrBy(pref+":vm","used",int64(
		virtual_memory.Used)).Err()
	if(err != nil){
		log.Printf("can not increment vm used: %s",err.Error())
	}

	err= f.redis_client.HIncrBy(
		pref+":vm","available",int64(virtual_memory.Free)).Err()
	if(err != nil){
		log.Printf("can not increment vm available: %s",err.Error())
	}

	pids, err := process.Pids()
	if err != nil {
		log.Printf("can not get top: %s",err.Error())
	}

	for _, pid := range pids {
		process, err := process.NewProcess(pid)
		if(err != nil){
			log.Printf("can not get process ID %v info: %s",pid,err.Error())
			continue
		}
		pid_string:= fmt.Sprintf("%v",pid)

		process_name,err:= process.Name()
		if(err != nil){
			log.Printf("can not get process ID %v name: %s",pid,err.Error())
			process_name="undefined"
		}
		err= f.redis_client.HSetNX(
			pref+":pids:names",pid_string,process_name).Err()
		if(err != nil){
			log.Printf("can not write process %v name: %s",pid,err.Error())
		}
		status,err:= process.Status()
		if(err != nil){
			log.Printf("can not get process ID  %v status: %s",pid,err.Error())
			status = "undefined"
		}

		err= f.redis_client.HSet(pref+":pids:status",pid_string,status).Err()
		if(err != nil){
			log.Printf("can not write process ID %v status: %s",pid,err.Error())
		}

		cwd,err:= process.Cwd()
		if(err != nil){
			cwd= err.Error()
			log.Printf("can not get process ID %v cwd: %s",pid,err.Error())
		}

		err= f.redis_client.HSetNX(pref+":pids:cwd",pid_string,cwd).Err()
		if(err != nil){
			log.Printf("can not write process ID %v cwd: %s",pid,err.Error())
		}

		create_time,err:= process.CreateTime()
		if(err != nil){
			create_time = time.Unix(0,0).Unix()
			log.Printf(
				"can not get process ID %v creation time: %s",pid,err.Error())
		}
		create_time_string:= time.Unix(create_time,0).Format(
			"Jan 02, 2006 15:04:05")
		err= f.redis_client.HSet(
			pref+":pids:creation_time",pid_string,create_time_string).Err()
		if(err != nil){
			log.Printf("can not write process ID %v name: %s",pid,err.Error())
		}

		memory_info,err:= process.MemoryInfo()
		if(err != nil){
			log.Printf(
				"can not get process ID %v memory info: %s",pid,err.Error())
		}
		err= f.redis_client.HSet(
			pref+":pids:memory_info",pid_string,memory_info.String()).Err()
		if(err != nil){
			log.Printf(
				"can not write process ID %v memory info: %s",pid,err.Error())
		}

		num_threads,err:= process.NumThreads()
		if(err != nil){
			log.Printf(
				"can not get process ID %v num threads: %s",pid,err.Error())
		}
		num_treads_string:= strconv.Itoa(int(num_threads))
		err= f.redis_client.HSet(
			pref+":pids:num_threads",pid_string,num_treads_string).Err()
		if(err != nil){
			log.Printf(
				"can not write process ID %v num threads: %s",pid,err.Error())
		}

		mem_percent,err:= process.MemoryPercent()
		if(err != nil){
			mem_percent = 0.0
			log.Printf(
				"can not get process ID %v memory percent %s",pid,err.Error())
		}
		err= f.redis_client.HIncrByFloat(
			pref+":pids:mem_percent",pid_string,float64(mem_percent)).Err()
		if(err != nil){
			log.Printf(
				"can not write process ID %v memory percent: %s",pid,err.Error())
		}

		cpu_percent,err:= process.Percent(time.Second)
		if(err != nil){
			cpu_percent = 0.0
			log.Printf(
				"can not get process ID %v cpu percent %s",pid,err.Error())
		}

		err= f.redis_client.HIncrByFloat(
			pref+":pids:cpu_percent",pid_string,cpu_percent).Err()
		if(err != nil){
			log.Printf(
				"can not write process ID %v cpu percent: %s",pid,err.Error())
		}
	}
}

// Reads system information from redis.
//
// param: test_id string   ID of current test.
func(f *SystemInfoFactory)ReadSystemInfo(test_id string)(*SystemInfo){
	pref:= "system:"+test_id
	steps_count,err:= f.redis_client.Get(pref+":steps").Float64()
	if(err != nil || steps_count == 0.0){
		log.Printf("can not read steps count: %s",err.Error())
		return f.system_info
	}

	cpu,err:= f.redis_client.Get(pref+":cpu").Float64()
	f.system_info.CPUusage = f.roundPercents64(cpu/steps_count)

	swap_info:=  &MemoryInfo{}
	swap_total,err:= f.redis_client.HGet(pref+":swap","total").Float64()
	if(err != nil){
		log.Printf("can not read swap total: %s",err.Error())
		swap_total =0.0
	}
	swap_used, err:= f.redis_client.HGet(pref+":swap","used").Float64()
	if(err != nil){
		log.Printf("can not read swap used: %s",err.Error())
		swap_used= 0.0;
	}
	swap_available,err:= f.redis_client.HGet(pref+":swap","available").Float64()
	if(err != nil){
		log.Printf("can not read swap available: %s",err.Error())
		swap_available =0.0
	}
	swap_percent,err:= f.redis_client.HGet(pref+":swap","percent").Float64()
	if(err != nil){
		log.Printf("can not read swap percent: %s",err.Error())
		swap_percent =0.0
	}

	swap_info.Total = f.toMegaBytes(swap_total/steps_count)
	swap_info.Used = f.toMegaBytes(swap_used/steps_count)
	swap_info.Available = f.toMegaBytes(swap_available/steps_count)
	swap_info.UsedPercent = f.roundPercents64(swap_percent/steps_count)
	f.system_info.SWAPmemoryInfo=swap_info

	vm_info := &MemoryInfo{}
	vm_total,err:= f.redis_client.HGet(pref+":vm","total").Float64()
	if(err != nil){
		log.Printf("can not read swap total: %s",err.Error())
		vm_total =0.0
	}

	vm_used, err:= f.redis_client.HGet(pref+":vm","used").Float64()
	if(err != nil){
		log.Printf("can not read virtual memory used: %s",err.Error())
		vm_used= 0.0;
	}

	vm_available,err:= f.redis_client.HGet(pref+":vm","available").Float64()
	if(err != nil){
		log.Printf("can not read virtual memory available: %s",err.Error())
		vm_available =0.0
	}

	vm_percent,err:= f.redis_client.HGet(pref+":vm","percent").Float64()
	if(err != nil){
		log.Printf("can not read virtual memory percent: %s",err.Error())
		vm_percent =0.0
	}

	top_length,err:= f.redis_client.HLen(pref+":pids:names").Result()
	if(err != nil){
		log.Printf("can not read top list length: %s",err.Error())
		return f.system_info
	}

	top_list,err:= f.redis_client.HGetAll(pref+":pids:names").Result()
	if(err != nil){
		log.Printf("can not read top list: %s",err.Error())
		return f.system_info
	}

	f.system_info.Top = make([]*ProcessInfo, top_length)
	count:=0;
	for pid, process_name  := range top_list {
		process_info:= &ProcessInfo{}
		process_info.Name = process_name
		pid_int,err:= strconv.ParseInt(pid,10,32)
		if(err != nil){
			log.Printf("can not parse process pid %s",pid)
			continue
		}
		process_info.PID = int32(pid_int)

		status,err:= f.redis_client.HGet(pref+":pids:status",pid).Result()
		if(err != nil){
			log.Printf("can not read processID %v status %s",pid,err.Error())
		}
		cwd,err:= f.redis_client.HGet(pref+":pids:cwd",pid).Result()
		if(err != nil){
			cwd= "undefined"
		}

		creation_time_string,err:= f.redis_client.HGet(
			pref+":pids:creation_time",pid).Result()
		if(err != nil){
			log.Printf(
				"can not read process ID: %v creation date %s",pid,err.Error())
			creation_time_string = time.Unix(0,0).Format(
				"Jan 02, 2006 15:04:05")
		}
		memory_info_bytes, err:= f.redis_client.HGet(
			pref+":pids:memory_info",pid).Bytes()
		if(err != nil){
			log.Printf(
				"can not read process ID: %v memory info %s",pid,err.Error())
		}
		process_memory_info:= &process.MemoryInfoStat{}
		err= json.Unmarshal(memory_info_bytes,process_memory_info)
		if(err != nil){
			log.Printf(
				"can not unmarshall process ID: %v memory info %s",
				pid,err.Error())
		}

		num_threads,err:= f.redis_client.HGet(
			pref+":pids:num_threads",pid).Int64()
		if(err != nil){
			log.Printf(
				"can not read process ID: %v num threads %s",pid,err.Error())
		}

		mem_percent,err:= f.redis_client.HGet(
			pref+":pids:mem_percent",pid).Float64()
		if(err != nil){
			log.Printf(
				"can not read process ID: %v memory percent %s",pid,err.Error())
			mem_percent = 0.0
		}
		log.Printf("process pid %s memory percent: %d",pid,mem_percent)
		process_cpu_percent,err:= f.redis_client.HGet(
			pref+":pids:cpu_percent",pid).Float64()
		if(err != nil){
			process_cpu_percent = 0.0
			log.Printf(
				"can not read process ID: %v CPU percent %s",pid,err.Error())
		}
		process_info.CPUPercent = f.roundPercents64(
			process_cpu_percent/steps_count)
		process_info.MemoryPercent = f.roundPercents64(mem_percent/steps_count)
		process_info.NumThreads = num_threads
		process_info.MemoryInfo = process_memory_info
		process_info.CreateTime = creation_time_string
		process_info.Cwd = cwd
		process_info.Status = status
		f.system_info.Top[count]=process_info
		count++
	}

	vm_info.Total = f.toMegaBytes(vm_total/steps_count)
	vm_info.Used = f.toMegaBytes(vm_used/steps_count)
	vm_info.Available = f.toMegaBytes(vm_available/steps_count)
	vm_info.UsedPercent = f.roundPercents64(vm_percent/steps_count)
	f.system_info.VirtualMemoryInfo = vm_info
	if f.system_info.Top != nil && len(f.system_info.Top) > 1 {
		sort.Sort(ByCPU(f.system_info.Top))
	}
	return f.system_info
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

// Converts bytes to megabytes.
//
// param value float64    Value in bytes.
// return: float64        The value in megabytes.
func (f *SystemInfoFactory) toMegaBytes(value float64) float64 {
	return math.Floor(value/(1024*1024)*100) / 100
}

// Rounds percents values.
//
// return: float64   Rounded value.
func (f *SystemInfoFactory) roundPercents64(value float64) float64 {
	return float64(math.Floor(value*100)) / 100
}
