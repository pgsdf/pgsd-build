package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LogFunc is a function that logs installation progress
type LogFunc func(string)

// Config holds installation configuration
type Config struct {
	ImagePath  string // Path to the image directory (containing root.zfs.xz, efi.img, manifest.toml)
	TargetDisk string // Target disk device (e.g., "ada0")
	ZpoolName  string // Name of the ZFS pool to create
	LogFunc    LogFunc
}

// Install performs the full ZFS-based installation pipeline
func Install(cfg Config) error {
	log := cfg.LogFunc
	if log == nil {
		log = func(s string) {} // No-op logger
	}

	// Validate configuration
	log("Validating installation configuration...")
	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("invalid installation configuration: %w", err)
	}

	// Check for required tools
	log("Checking system requirements...")
	if err := checkRequirements(); err != nil {
		return fmt.Errorf("system requirements not met: %w", err)
	}

	// Step 1: Partition the disk
	log("Partitioning disk...")
	if err := partitionDisk(cfg.TargetDisk); err != nil {
		return fmt.Errorf("disk partitioning failed: %w\nHint: Ensure the disk is not in use and you have root privileges", err)
	}

	// Step 2: Create EFI filesystem
	log("Creating EFI system partition...")
	efiPart := cfg.TargetDisk + "p1"
	if err := createEFIFilesystem(efiPart); err != nil {
		return fmt.Errorf("EFI filesystem creation failed: %w\nHint: The partition may not be properly created", err)
	}

	// Step 3: Create ZFS pool
	log("Creating ZFS pool...")
	zfsPart := cfg.TargetDisk + "p2"
	if err := createZFSPool(cfg.ZpoolName, zfsPart); err != nil {
		return fmt.Errorf("ZFS pool creation failed: %w\nHint: Ensure ZFS kernel module is loaded (kldload zfs)", err)
	}

	// Step 4: Extract root filesystem
	log("Extracting root filesystem (this may take several minutes)...")
	rootZFS := filepath.Join(cfg.ImagePath, "root.zfs.xz")
	if err := extractZFSStream(rootZFS, cfg.ZpoolName); err != nil {
		return fmt.Errorf("root filesystem extraction failed: %w\nHint: Ensure the ZFS stream file is not corrupted", err)
	}

	// Step 5: Copy EFI partition
	log("Installing EFI partition...")
	efiImg := filepath.Join(cfg.ImagePath, "efi.img")
	if err := copyEFIPartition(efiImg, efiPart); err != nil {
		return fmt.Errorf("EFI partition installation failed: %w", err)
	}

	// Step 6: Install bootloader
	log("Installing bootloader...")
	if err := installBootloader(cfg.TargetDisk, cfg.ZpoolName); err != nil {
		return fmt.Errorf("bootloader installation failed: %w\nHint: Ensure /boot/boot1.efifat exists on the system", err)
	}

	// Step 7: Finalize
	log("Finalizing installation...")
	if err := finalizeInstallation(cfg.ZpoolName); err != nil {
		return fmt.Errorf("installation finalization failed: %w", err)
	}

	log("Installation complete!")
	return nil
}

// partitionDisk creates a GPT partition table with EFI and ZFS partitions
func partitionDisk(disk string) error {
	// On FreeBSD:
	// gpart destroy -F disk (if exists)
	// gpart create -s gpt disk
	// gpart add -t efi -s 200M -l efiboot0 disk
	// gpart add -t freebsd-zfs -l zfsroot0 disk

	commands := [][]string{
		{"gpart", "destroy", "-F", disk},
		{"gpart", "create", "-s", "gpt", disk},
		{"gpart", "add", "-t", "efi", "-s", "200M", "-l", "efiboot0", disk},
		{"gpart", "add", "-t", "freebsd-zfs", "-l", "zfsroot0", disk},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if output, err := cmd.CombinedOutput(); err != nil {
			// Ignore error on destroy if partition table doesn't exist
			if args[1] != "destroy" {
				return fmt.Errorf("command %v failed: %w\nOutput: %s",
					strings.Join(args, " "), err, output)
			}
		}
	}

	return nil
}

// createEFIFilesystem creates a FAT32 filesystem on the EFI partition
func createEFIFilesystem(efiPart string) error {
	// On FreeBSD:
	// newfs_msdos -F 32 -c 1 efiPart

	cmd := exec.Command("newfs_msdos", "-F", "32", "-c", "1", efiPart)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("newfs_msdos failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// createZFSPool creates a ZFS pool on the ZFS partition
func createZFSPool(poolName, zfsPart string) error {
	// On FreeBSD:
	// zpool create -f -o altroot=/mnt -O compression=lz4 -O atime=off poolName zfsPart

	cmd := exec.Command("zpool", "create", "-f",
		"-o", "altroot=/mnt",
		"-O", "compression=lz4",
		"-O", "atime=off",
		poolName, zfsPart)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("zpool create failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// extractZFSStream extracts a compressed ZFS stream to the pool
func extractZFSStream(rootZFS, poolName string) error {
	// On FreeBSD:
	// xzcat rootZFS | zfs receive -F poolName/ROOT/default

	// Check if source file exists
	if _, err := os.Stat(rootZFS); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ZFS stream file not found: %s", rootZFS)
		}
		return fmt.Errorf("cannot access ZFS stream file: %w", err)
	}

	// Build the pipeline: xzcat | zfs receive
	xzcat := exec.Command("xzcat", rootZFS)
	zfsRecv := exec.Command("zfs", "receive", "-F",
		fmt.Sprintf("%s/ROOT/default", poolName))

	// Connect the pipeline
	pipe, err := xzcat.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe between xzcat and zfs receive: %w", err)
	}
	zfsRecv.Stdin = pipe

	// Capture stderr for better error messages
	xzcatStderr, _ := xzcat.StderrPipe()
	zfsStderr, _ := zfsRecv.StderrPipe()

	// Start both commands
	if err := zfsRecv.Start(); err != nil {
		return fmt.Errorf("failed to start zfs receive command: %w", err)
	}

	if err := xzcat.Start(); err != nil {
		return fmt.Errorf("failed to start xzcat command: %w", err)
	}

	// Wait for completion
	if err := xzcat.Wait(); err != nil {
		stderr := readPipe(xzcatStderr)
		return fmt.Errorf("xzcat decompression failed: %w\nDetails: %s", err, stderr)
	}

	if err := zfsRecv.Wait(); err != nil {
		stderr := readPipe(zfsStderr)
		return fmt.Errorf("zfs receive failed: %w\nDetails: %s", err, stderr)
	}

	return nil
}

// copyEFIPartition copies the EFI image to the EFI partition
func copyEFIPartition(efiImg, efiPart string) error {
	// On FreeBSD:
	// dd if=efiImg of=efiPart bs=1M

	// Check if source file exists
	if _, err := os.Stat(efiImg); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("EFI image file not found: %s", efiImg)
		}
		return fmt.Errorf("cannot access EFI image file: %w", err)
	}

	cmd := exec.Command("dd",
		"if="+efiImg,
		"of="+efiPart,
		"bs=1M")

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy EFI partition: %w\nOutput: %s", err, output)
	}

	return nil
}

// installBootloader installs the FreeBSD bootloader
func installBootloader(disk, poolName string) error {
	// On FreeBSD:
	// gpart bootcode -p /boot/boot1.efifat -i 1 disk

	cmd := exec.Command("gpart", "bootcode",
		"-p", "/boot/boot1.efifat",
		"-i", "1",
		disk)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gpart bootcode failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// finalizeInstallation performs final cleanup and configuration
func finalizeInstallation(poolName string) error {
	// Set bootfs property
	cmd := exec.Command("zpool", "set",
		fmt.Sprintf("bootfs=%s/ROOT/default", poolName),
		poolName)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("zpool set bootfs failed: %w\nOutput: %s", err, output)
	}

	// Export the pool
	cmd = exec.Command("zpool", "export", poolName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("zpool export failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// validateConfig validates the installation configuration
func validateConfig(cfg *Config) error {
	if cfg.ImagePath == "" {
		return fmt.Errorf("image path is required")
	}
	if cfg.TargetDisk == "" {
		return fmt.Errorf("target disk is required")
	}
	if cfg.ZpoolName == "" {
		return fmt.Errorf("zpool name is required")
	}

	// Check if image directory exists
	if _, err := os.Stat(cfg.ImagePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("image directory not found: %s", cfg.ImagePath)
		}
		return fmt.Errorf("cannot access image directory: %w", err)
	}

	// Check for required files
	requiredFiles := []string{"root.zfs.xz", "efi.img", "manifest.toml"}
	for _, file := range requiredFiles {
		path := filepath.Join(cfg.ImagePath, file)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("required file missing: %s\nThe image directory must contain: %v", file, requiredFiles)
			}
			return fmt.Errorf("cannot access required file %s: %w", file, err)
		}
	}

	// Validate zpool name
	if len(cfg.ZpoolName) > 63 {
		return fmt.Errorf("zpool name too long (max 63 characters): %s", cfg.ZpoolName)
	}
	if strings.ContainsAny(cfg.ZpoolName, " /\\") {
		return fmt.Errorf("zpool name contains invalid characters (no spaces or slashes): %s", cfg.ZpoolName)
	}

	return nil
}

// checkRequirements checks if required system commands are available
func checkRequirements() error {
	required := []string{"gpart", "newfs_msdos", "zpool", "zfs", "xzcat", "dd"}
	var missing []string

	for _, cmd := range required {
		if _, err := exec.LookPath(cmd); err != nil {
			missing = append(missing, cmd)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required commands not found: %v\nPlease ensure these tools are installed and in PATH", missing)
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("installation must be run as root\nTry: sudo pgsd-inst")
	}

	return nil
}

// readPipe reads all data from a pipe and returns it as a string
func readPipe(pipe interface{}) string {
	if pipe == nil {
		return ""
	}
	// This is a simplified version; in production we'd use io.ReadAll
	return ""
}
