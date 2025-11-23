package install

import (
    "fmt"
    "os/exec"
    "path/filepath"
    "strings"
)

// LogFunc is a function that logs installation progress
type LogFunc func(string)

// Config holds installation configuration
type Config struct {
    ImagePath    string // Path to the image directory (containing root.zfs.xz, efi.img, manifest.toml)
    TargetDisk   string // Target disk device (e.g., "ada0")
    ZpoolName    string // Name of the ZFS pool to create
    LogFunc      LogFunc
}

// Install performs the full ZFS-based installation pipeline
func Install(cfg Config) error {
    log := cfg.LogFunc
    if log == nil {
        log = func(s string) {} // No-op logger
    }

    // Step 1: Partition the disk
    log("Partitioning disk...")
    if err := partitionDisk(cfg.TargetDisk); err != nil {
        return fmt.Errorf("failed to partition disk: %w", err)
    }

    // Step 2: Create EFI filesystem
    log("Creating EFI system partition...")
    efiPart := cfg.TargetDisk + "p1"
    if err := createEFIFilesystem(efiPart); err != nil {
        return fmt.Errorf("failed to create EFI filesystem: %w", err)
    }

    // Step 3: Create ZFS pool
    log("Creating ZFS pool...")
    zfsPart := cfg.TargetDisk + "p2"
    if err := createZFSPool(cfg.ZpoolName, zfsPart); err != nil {
        return fmt.Errorf("failed to create ZFS pool: %w", err)
    }

    // Step 4: Extract root filesystem
    log("Extracting root filesystem...")
    rootZFS := filepath.Join(cfg.ImagePath, "root.zfs.xz")
    if err := extractZFSStream(rootZFS, cfg.ZpoolName); err != nil {
        return fmt.Errorf("failed to extract ZFS stream: %w", err)
    }

    // Step 5: Copy EFI partition
    log("Installing EFI partition...")
    efiImg := filepath.Join(cfg.ImagePath, "efi.img")
    if err := copyEFIPartition(efiImg, efiPart); err != nil {
        return fmt.Errorf("failed to copy EFI partition: %w", err)
    }

    // Step 6: Install bootloader
    log("Installing bootloader...")
    if err := installBootloader(cfg.TargetDisk, cfg.ZpoolName); err != nil {
        return fmt.Errorf("failed to install bootloader: %w", err)
    }

    // Step 7: Finalize
    log("Finalizing installation...")
    if err := finalizeInstallation(cfg.ZpoolName); err != nil {
        return fmt.Errorf("failed to finalize installation: %w", err)
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

    // Build the pipeline: xzcat | zfs receive
    xzcat := exec.Command("xzcat", rootZFS)
    zfsRecv := exec.Command("zfs", "receive", "-F",
        fmt.Sprintf("%s/ROOT/default", poolName))

    // Connect the pipeline
    pipe, err := xzcat.StdoutPipe()
    if err != nil {
        return fmt.Errorf("failed to create pipe: %w", err)
    }
    zfsRecv.Stdin = pipe

    // Start both commands
    if err := zfsRecv.Start(); err != nil {
        return fmt.Errorf("failed to start zfs receive: %w", err)
    }

    if err := xzcat.Start(); err != nil {
        return fmt.Errorf("failed to start xzcat: %w", err)
    }

    // Wait for completion
    if err := xzcat.Wait(); err != nil {
        return fmt.Errorf("xzcat failed: %w", err)
    }

    if err := zfsRecv.Wait(); err != nil {
        return fmt.Errorf("zfs receive failed: %w", err)
    }

    return nil
}

// copyEFIPartition copies the EFI image to the EFI partition
func copyEFIPartition(efiImg, efiPart string) error {
    // On FreeBSD:
    // dd if=efiImg of=efiPart bs=1M

    cmd := exec.Command("dd",
        "if="+efiImg,
        "of="+efiPart,
        "bs=1M")

    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("dd failed: %w\nOutput: %s", err, output)
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
