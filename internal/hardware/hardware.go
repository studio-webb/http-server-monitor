package hardware

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"runtime"
)

func GetSystemSection() (string, error) {
	runTimeOs := runtime.GOOS

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}

	hostStat, err := host.Info()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("Hostname: %s\nTotal Memory: %d\nUsed Memory: %d\nOS: %s\n",
		hostStat.Hostname, vmStat.Total, vmStat.Used, runTimeOs)

	return output, nil
}

func FetCPUSection() (string, error) {
	cpuStat, err := cpu.Info()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("CPU: %s\nCores: %d\n", cpuStat[0].ModelName, len(cpuStat))

	return output, nil
}

func GetDiskSection() (string, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("Total Disk Spase:%d\nFree Disk Space: %d\n", diskStat.Total, diskStat.Used)
	return output, nil
}
