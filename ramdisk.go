package main

import "os"
import "fmt"
import "flag"
import "strings"
import "os/exec"

var ramdiskSizeInGigabytes int

func init() {
	flag.IntVar(&ramdiskSizeInGigabytes, "size", 4, "ramdisk size in gigabytes")
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
	sectors := convertGigabytesToSectors(ramdiskSizeInGigabytes)
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

	flag.Usage = func() {
		fmt.Printf("Usage: ramdisk [options] {start|sync|stop}>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
	}

	command := flag.Args()[0]

	switch command {
	case "start":
		start()
	case "stop":
		stop()
	case "sync":
		sync()
	default:
		flag.Usage()
	}

}
