package main

import "syscall"
import "os"
import "fmt"
import "strings"
import "os/exec"

func rsync(args ...string) {
	env := os.Environ()
	execErr := syscall.Exec("rsync", args, env)
	if execErr != nil {
		panic(execErr)
	}
}

// https://bogner.sh/2012/12/os-x-create-a-ram-disk-the-easy-way/
func convertGigabytesToSectors(gigabytes int) int {
	return gigabytes * 1024 * 1024 * 1024 / 512
}

func getRamdiskBackupDir() string {
	ramdiskBackupDir := os.Getenv("RAMDISK_BACKUP_DIR")
	if ramdiskBackupDir == "" {
		fmt.Println("Must set RAMDISK_BACKUP_DIR")
		os.Exit(1)
	}
	fmt.Println("RAMDISK_BACKUP_DIR:", ramdiskBackupDir)
	return ramdiskBackupDir
}

func start() {
	sectors := convertGigabytesToSectors(4)
	device := verboseExec("hdiutil", "attach", "-nomount", fmt.Sprintf("ram://%d", sectors))

	verboseExec("diskutil", "erasevolume", "HFS+", "ramdisk", device)

	ramdiskBackupDir := getRamdiskBackupDir()
	verboseExec("rsync", "-av", ramdiskBackupDir+"/", "/Volumes/ramdisk")
}

func stop() {
	sync()
	verboseExec("diskutil", "eject", "ramdisk")
}

func sync() {
	ramdiskBackupDir := getRamdiskBackupDir()
	verboseExec("rsync", "-av", "/Volumes/ramdisk/", ramdiskBackupDir)
}

func verboseExec(args ...string) string {
	fmt.Printf("==> Executing: %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}

	if len(output) > 0 {
		fmt.Printf("==> Output: %s\n", string(output))
	}

	return strings.TrimSpace(string(output))
}

// http://www.observium.org/wiki/Persistent_RAM_disk_RRD_storage
func main() {

	if len(os.Args) == 1 {
		fmt.Println("Usage: /etc/init.d/ramdisk {start|stop|sync}")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		start()
	case "stop":
		stop()
	case "sync":
		sync()
	}

}
