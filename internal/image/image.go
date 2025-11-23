package image

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/pgsdf/pgsdbuild/internal/config"
)

const (
    diskSizeGB      = 10
    artifactsDir    = "artifacts"
    workDir         = "work"
)

// BuildImage implements the ZFS image build pipeline.
func BuildImage(cfg config.ImageConfig) error {
    fmt.Printf("[image] Starting build for %s\n", cfg.ID)

    // Create working directories
    artifactPath := filepath.Join(artifactsDir, cfg.ID)
    if err := os.MkdirAll(artifactPath, 0755); err != nil {
        return fmt.Errorf("failed to create artifact directory: %w", err)
    }

    workPath := filepath.Join(workDir, cfg.ID)
    if err := os.MkdirAll(workPath, 0755); err != nil {
        return fmt.Errorf("failed to create work directory: %w", err)
    }
    defer os.RemoveAll(workPath) // Clean up work directory

    // Step 1: Create md-backed disk
    fmt.Println("[image] Creating memory-backed disk...")
    mdDevice, err := createMemoryDisk(workPath, diskSizeGB)
    if err != nil {
        return fmt.Errorf("failed to create memory disk: %w", err)
    }
    defer destroyMemoryDisk(mdDevice)

    // Step 2: Partition the disk
    fmt.Println("[image] Partitioning disk...")
    efiPart, zfsPart, err := partitionDisk(mdDevice)
    if err != nil {
        return fmt.Errorf("failed to partition disk: %w", err)
    }

    // Step 3: Create filesystems
    fmt.Println("[image] Creating EFI filesystem...")
    if err := createEFIFilesystem(efiPart); err != nil {
        return fmt.Errorf("failed to create EFI filesystem: %w", err)
    }

    fmt.Println("[image] Creating ZFS pool and datasets...")
    if err := createZFSPool(cfg, zfsPart); err != nil {
        return fmt.Errorf("failed to create ZFS pool: %w", err)
    }
    defer destroyZFSPool(cfg.ZpoolName)

    // Step 4: Install packages and overlays
    rootMount := fmt.Sprintf("/%s/ROOT/default", cfg.ZpoolName)

    fmt.Println("[image] Installing packages...")
    if err := installPackages(cfg, rootMount); err != nil {
        return fmt.Errorf("failed to install packages: %w", err)
    }

    fmt.Println("[image] Applying overlays...")
    if err := applyOverlays(cfg, rootMount); err != nil {
        return fmt.Errorf("failed to apply overlays: %w", err)
    }

    // Step 5: Create snapshot
    fmt.Println("[image] Creating ZFS snapshot...")
    snapshot := fmt.Sprintf("%s@install", cfg.RootDS)
    if err := createSnapshot(snapshot); err != nil {
        return fmt.Errorf("failed to create snapshot: %w", err)
    }

    // Step 6: Export artifacts
    fmt.Println("[image] Exporting ZFS stream...")
    zfsArtifact := filepath.Join(artifactPath, "root.zfs.xz")
    if err := exportZFSStream(snapshot, zfsArtifact); err != nil {
        return fmt.Errorf("failed to export ZFS stream: %w", err)
    }

    fmt.Println("[image] Exporting EFI partition...")
    efiArtifact := filepath.Join(artifactPath, "efi.img")
    if err := exportEFIPartition(efiPart, efiArtifact); err != nil {
        return fmt.Errorf("failed to export EFI partition: %w", err)
    }

    // Step 7: Create manifest
    fmt.Println("[image] Creating manifest...")
    manifestPath := filepath.Join(artifactPath, "manifest.toml")
    if err := createManifest(cfg, manifestPath); err != nil {
        return fmt.Errorf("failed to create manifest: %w", err)
    }

    fmt.Printf("[image] Build complete! Artifacts in %s/\n", artifactPath)
    return nil
}

// createMemoryDisk creates an md-backed disk and returns the device name.
func createMemoryDisk(workPath string, sizeGB int) (string, error) {
    diskPath := filepath.Join(workPath, "disk.img")

    // Create a sparse file
    size := int64(sizeGB) * 1024 * 1024 * 1024
    f, err := os.Create(diskPath)
    if err != nil {
        return "", err
    }
    if err := f.Truncate(size); err != nil {
        f.Close()
        return "", err
    }
    f.Close()

    // On FreeBSD, we would use: mdconfig -a -t vnode -f diskPath
    // For this prototype, we'll simulate it
    return diskPath, nil
}

// destroyMemoryDisk destroys the memory disk.
func destroyMemoryDisk(mdDevice string) {
    // On FreeBSD: mdconfig -d -u <unit>
    // For prototype, just remove the file
    os.Remove(mdDevice)
}

// partitionDisk partitions the disk with GPT, EFI, and ZFS partitions.
func partitionDisk(mdDevice string) (efi, zfs string, err error) {
    // On FreeBSD, we would use gpart:
    // gpart create -s gpt mdDevice
    // gpart add -t efi -s 200M mdDevice
    // gpart add -t freebsd-zfs mdDevice

    // For prototype, we'll create dummy partition files
    efi = mdDevice + ".p1"
    zfs = mdDevice + ".p2"

    // Create dummy partition files
    if err := createDummyPartition(efi, 200*1024*1024); err != nil {
        return "", "", err
    }
    if err := createDummyPartition(zfs, 9*1024*1024*1024); err != nil {
        return "", "", err
    }

    return efi, zfs, nil
}

// createDummyPartition creates a sparse file to simulate a partition.
func createDummyPartition(path string, size int64) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    return f.Truncate(size)
}

// createEFIFilesystem creates a FAT32 filesystem on the EFI partition.
func createEFIFilesystem(efiPart string) error {
    // On FreeBSD: newfs_msdos -F 32 efiPart
    // For prototype, we'll just note it was created
    return nil
}

// createZFSPool creates a ZFS pool and datasets.
func createZFSPool(cfg config.ImageConfig, zfsPart string) error {
    // On FreeBSD:
    // zpool create -o altroot=/mnt -O compression=lz4 -O atime=off poolName zfsPart
    // zfs create -o mountpoint=none poolName/ROOT
    // zfs create -o mountpoint=/ poolName/ROOT/default

    // For prototype, we'll create a directory structure
    rootMount := fmt.Sprintf("/%s/ROOT/default", cfg.ZpoolName)
    if err := os.MkdirAll(rootMount, 0755); err != nil {
        return err
    }

    fmt.Printf("[zfs] Created pool %s with root dataset %s\n", cfg.ZpoolName, cfg.RootDS)
    return nil
}

// destroyZFSPool destroys the ZFS pool.
func destroyZFSPool(poolName string) {
    // On FreeBSD: zpool destroy -f poolName
    // For prototype, remove the directory
    os.RemoveAll(fmt.Sprintf("/%s", poolName))
}

// installPackages installs packages into the root mount.
func installPackages(cfg config.ImageConfig, rootMount string) error {
    // On FreeBSD:
    // pkg -r rootMount install -y <packages>

    // For prototype, we'll create a marker file showing what packages would be installed
    pkgList := filepath.Join(rootMount, "installed-packages.txt")
    f, err := os.Create(pkgList)
    if err != nil {
        return err
    }
    defer f.Close()

    for _, pkgSet := range cfg.PkgLists {
        fmt.Fprintf(f, "# Package set: %s\n", pkgSet)
    }

    fmt.Printf("[pkg] Installed package lists: %v\n", cfg.PkgLists)
    return nil
}

// applyOverlays copies overlay files into the root mount.
func applyOverlays(cfg config.ImageConfig, rootMount string) error {
    for _, overlay := range cfg.Overlays {
        overlayPath := filepath.Join("overlays", overlay)

        // Check if overlay exists
        if _, err := os.Stat(overlayPath); os.IsNotExist(err) {
            return fmt.Errorf("overlay %s not found at %s", overlay, overlayPath)
        }

        // Copy overlay contents to rootMount
        // We would use: cp -a overlayPath/* rootMount/
        if err := copyOverlay(overlayPath, rootMount); err != nil {
            return fmt.Errorf("failed to copy overlay %s: %w", overlay, err)
        }

        fmt.Printf("[overlay] Applied overlay: %s\n", overlay)
    }
    return nil
}

// copyOverlay recursively copies overlay files.
func copyOverlay(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Calculate relative path
        relPath, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }

        dstPath := filepath.Join(dst, relPath)

        if info.IsDir() {
            return os.MkdirAll(dstPath, info.Mode())
        }

        // Copy file
        return copyFile(path, dstPath, info.Mode())
    })
}

// copyFile copies a single file.
func copyFile(src, dst string, mode os.FileMode) error {
    data, err := os.ReadFile(src)
    if err != nil {
        return err
    }
    return os.WriteFile(dst, data, mode)
}

// createSnapshot creates a ZFS snapshot.
func createSnapshot(snapshot string) error {
    // On FreeBSD: zfs snapshot snapshot
    fmt.Printf("[zfs] Created snapshot: %s\n", snapshot)
    return nil
}

// exportZFSStream exports a ZFS snapshot as a compressed stream.
func exportZFSStream(snapshot, outputPath string) error {
    // On FreeBSD: zfs send snapshot | xz -9 > outputPath

    // For prototype, create a dummy compressed file
    f, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer f.Close()

    // Write a minimal header indicating this is a ZFS stream
    _, err = f.WriteString(fmt.Sprintf("# ZFS snapshot: %s\n# Created: %s\n",
        snapshot, time.Now().Format(time.RFC3339)))
    return err
}

// exportEFIPartition exports the EFI partition as an image.
func exportEFIPartition(efiPart, outputPath string) error {
    // On FreeBSD: dd if=efiPart of=outputPath bs=1M

    // For prototype, copy the dummy partition file
    data, err := os.ReadFile(efiPart)
    if err != nil {
        return err
    }
    return os.WriteFile(outputPath, data, 0644)
}

// createManifest creates a TOML manifest file.
func createManifest(cfg config.ImageConfig, manifestPath string) error {
    f, err := os.Create(manifestPath)
    if err != nil {
        return err
    }
    defer f.Close()

    // Write TOML manifest
    fmt.Fprintf(f, "# PGSD Image Manifest\n")
    fmt.Fprintf(f, "# Generated: %s\n\n", time.Now().Format(time.RFC3339))
    fmt.Fprintf(f, "[image]\n")
    fmt.Fprintf(f, "id = %q\n", cfg.ID)
    fmt.Fprintf(f, "version = %q\n", cfg.Version)
    fmt.Fprintf(f, "zpool_name = %q\n", cfg.ZpoolName)
    fmt.Fprintf(f, "root_dataset = %q\n", cfg.RootDS)
    fmt.Fprintf(f, "\n[artifacts]\n")
    fmt.Fprintf(f, "root_zfs = \"root.zfs.xz\"\n")
    fmt.Fprintf(f, "efi_image = \"efi.img\"\n")
    fmt.Fprintf(f, "\n[[package_lists]]\n")
    fmt.Fprintf(f, "sets = %s\n", formatStringArray(cfg.PkgLists))
    fmt.Fprintf(f, "\n[[overlays]]\n")
    fmt.Fprintf(f, "applied = %s\n", formatStringArray(cfg.Overlays))

    return nil
}

// formatStringArray formats a string slice as a TOML array.
func formatStringArray(arr []string) string {
    quoted := make([]string, len(arr))
    for i, s := range arr {
        quoted[i] = fmt.Sprintf("%q", s)
    }
    return "[" + strings.Join(quoted, ", ") + "]"
}
