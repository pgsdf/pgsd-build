package image

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pgsdf/pgsdbuild/internal/build"
	"github.com/pgsdf/pgsdbuild/internal/config"
	"github.com/pgsdf/pgsdbuild/internal/util"
)

// Builder builds ZFS-based system images.
type Builder struct {
	config *build.Config
	logger *util.Logger
}

// NewBuilder creates a new image Builder.
func NewBuilder(cfg *build.Config, logger *util.Logger) *Builder {
	return &Builder{
		config: cfg,
		logger: logger,
	}
}

// Build implements the ZFS image build pipeline.
func (b *Builder) Build(cfg config.ImageConfig) error {
	b.logger.Info("Starting build for %s", cfg.ID)

	// Create working directories
	artifactPath := filepath.Join(b.config.GetArtifactsDir(), cfg.ID)
	if err := util.EnsureDir(artifactPath); err != nil {
		return err
	}

	workPath := filepath.Join(b.config.GetWorkDir(), cfg.ID)
	if err := util.EnsureDir(workPath); err != nil {
		return err
	}

	if !b.config.KeepWork {
		defer func() {
			if err := util.CleanupDir(workPath); err != nil {
				b.logger.Warn("Failed to cleanup work directory %s: %v", workPath, err)
			}
		}()
	}

	// Step 1: Create md-backed disk
	b.logger.Debug("Creating memory-backed disk...")
	mdDevice, err := b.createMemoryDisk(workPath, b.config.DiskSizeGB)
	if err != nil {
		return fmt.Errorf("failed to create memory disk: %w", err)
	}
	defer b.destroyMemoryDisk(mdDevice)

	// Step 2: Partition the disk
	b.logger.Debug("Partitioning disk...")
	efiPart, zfsPart, err := b.partitionDisk(mdDevice)
	if err != nil {
		return fmt.Errorf("failed to partition disk: %w", err)
	}

	// Step 3: Create filesystems
	b.logger.Debug("Creating EFI filesystem...")
	if err := b.createEFIFilesystem(efiPart); err != nil {
		return fmt.Errorf("failed to create EFI filesystem: %w", err)
	}

	b.logger.Debug("Creating ZFS pool and datasets...")
	if err := b.createZFSPool(cfg, zfsPart); err != nil {
		return fmt.Errorf("failed to create ZFS pool: %w", err)
	}
	defer b.destroyZFSPool(cfg.ZpoolName)

	// Step 4: Install packages and overlays
	rootMount := fmt.Sprintf("/%s/ROOT/default", cfg.ZpoolName)

	b.logger.Debug("Installing packages...")
	if err := b.installPackages(cfg, rootMount); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	b.logger.Debug("Applying overlays...")
	if err := b.applyOverlays(cfg, rootMount); err != nil {
		return fmt.Errorf("failed to apply overlays: %w", err)
	}

	// Apply ZFS dataset overlays
	b.logger.Debug("Applying dataset overlays...")
	if err := b.applyDatasetOverlays(cfg); err != nil {
		return fmt.Errorf("failed to apply dataset overlays: %w", err)
	}

	// Step 5: Create snapshot
	b.logger.Debug("Creating ZFS snapshot...")
	snapshot := fmt.Sprintf("%s@install", cfg.RootDS)
	if err := b.createSnapshot(snapshot); err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Step 6: Export artifacts
	b.logger.Debug("Exporting ZFS stream...")
	zfsArtifact := filepath.Join(artifactPath, "root.zfs.xz")
	if err := b.exportZFSStream(snapshot, zfsArtifact); err != nil {
		return fmt.Errorf("failed to export ZFS stream: %w", err)
	}

	b.logger.Debug("Exporting EFI partition...")
	efiArtifact := filepath.Join(artifactPath, "efi.img")
	if err := b.exportEFIPartition(efiPart, efiArtifact); err != nil {
		return fmt.Errorf("failed to export EFI partition: %w", err)
	}

	// Step 7: Create manifest
	b.logger.Debug("Creating manifest...")
	manifestPath := filepath.Join(artifactPath, "manifest.toml")
	if err := b.createManifest(cfg, manifestPath); err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	b.logger.Info("Build complete! Artifacts in %s/", artifactPath)
	return nil
}

// createMemoryDisk creates an md-backed disk and returns the device name.
func (b *Builder) createMemoryDisk(workPath string, sizeGB int) (string, error) {
	diskPath := filepath.Join(workPath, "disk.img")

	// Create a sparse file
	size := int64(sizeGB) * 1024 * 1024 * 1024
	if err := util.CreateSparseFile(diskPath, size); err != nil {
		return "", err
	}

	// On FreeBSD, we would use: mdconfig -a -t vnode -f diskPath
	// For this prototype, we'll simulate it
	b.logger.Debug("Created sparse disk: %s (%d GB)", diskPath, sizeGB)
	return diskPath, nil
}

// destroyMemoryDisk destroys the memory disk.
func (b *Builder) destroyMemoryDisk(mdDevice string) {
	// On FreeBSD: mdconfig -d -u <unit>
	// For prototype, just remove the file
	os.Remove(mdDevice)
}

// partitionDisk partitions the disk with GPT, EFI, and ZFS partitions.
func (b *Builder) partitionDisk(mdDevice string) (efi, zfs string, err error) {
	// On FreeBSD, we would use gpart:
	// gpart create -s gpt mdDevice
	// gpart add -t efi -s 200M mdDevice
	// gpart add -t freebsd-zfs mdDevice

	// For prototype, we'll create dummy partition files
	efi = mdDevice + ".p1"
	zfs = mdDevice + ".p2"

	// Create dummy partition files
	if err := util.CreateSparseFile(efi, 200*1024*1024); err != nil {
		return "", "", err
	}
	if err := util.CreateSparseFile(zfs, 9*1024*1024*1024); err != nil {
		return "", "", err
	}

	b.logger.Debug("Created partitions: %s, %s", efi, zfs)
	return efi, zfs, nil
}

// createEFIFilesystem creates a FAT32 filesystem on the EFI partition.
func (b *Builder) createEFIFilesystem(efiPart string) error {
	// On FreeBSD: newfs_msdos -F 32 efiPart
	// For prototype, we'll just note it was created
	b.logger.Debug("EFI filesystem created on %s", efiPart)
	return nil
}

// createZFSPool creates a ZFS pool and datasets.
func (b *Builder) createZFSPool(cfg config.ImageConfig, zfsPart string) error {
	// On FreeBSD:
	// zpool create -o altroot=/mnt -O compression=lz4 -O atime=off poolName zfsPart
	// zfs create -o mountpoint=none poolName/ROOT
	// zfs create -o mountpoint=/ poolName/ROOT/default

	// For prototype, we'll create a directory structure
	rootMount := fmt.Sprintf("/%s/ROOT/default", cfg.ZpoolName)
	if err := util.EnsureDir(rootMount); err != nil {
		return err
	}

	b.logger.Debug("Created pool %s with root dataset %s", cfg.ZpoolName, cfg.RootDS)
	return nil
}

// destroyZFSPool destroys the ZFS pool.
func (b *Builder) destroyZFSPool(poolName string) {
	// On FreeBSD: zpool destroy -f poolName
	// For prototype, remove the directory
	_ = util.CleanupDir(fmt.Sprintf("/%s", poolName))
}

// installPackages installs packages into the root mount.
func (b *Builder) installPackages(cfg config.ImageConfig, rootMount string) error {
	// On FreeBSD:
	// pkg -r rootMount install -y <packages>

	// For prototype, we'll create a marker file showing what packages would be installed
	pkgList := filepath.Join(rootMount, "installed-packages.txt")
	content := fmt.Sprintf("# Image: %s\n# Package sets installed:\n", cfg.ID)
	for _, pkgSet := range cfg.PkgLists {
		content += fmt.Sprintf("# - %s\n", pkgSet)
	}

	if err := util.WriteStringToFile(pkgList, content, 0644); err != nil {
		return err
	}

	b.logger.Debug("Installed package lists: %v", cfg.PkgLists)
	return nil
}

// applyOverlays copies overlay files into the root mount.
func (b *Builder) applyOverlays(cfg config.ImageConfig, rootMount string) error {
	overlaysDir := b.config.GetOverlaysDir()

	for _, overlay := range cfg.Overlays {
		overlayPath := filepath.Join(overlaysDir, overlay)

		// Check if overlay exists
		if !util.DirExists(overlayPath) {
			return fmt.Errorf("overlay %s not found at %s", overlay, overlayPath)
		}

		// Copy overlay contents to rootMount
		if err := util.CopyOverlay(overlayPath, rootMount); err != nil {
			return fmt.Errorf("failed to copy overlay %s: %w", overlay, err)
		}

		b.logger.Debug("Applied overlay: %s", overlay)
	}
	return nil
}

// applyDatasetOverlays receives ZFS datasets into the image pool.
func (b *Builder) applyDatasetOverlays(cfg config.ImageConfig) error {
	if len(cfg.DatasetOverlays) == 0 {
		b.logger.Debug("No dataset overlays configured")
		return nil
	}

	pool := cfg.ZpoolName

	for _, o := range cfg.DatasetOverlays {
		if o.Source == "" || o.Name == "" {
			return fmt.Errorf("dataset overlay missing source or name")
		}

		target := fmt.Sprintf("%s/OVERLAYS/%s", pool, o.Name)

		b.logger.Info("Receiving dataset overlay: %s -> %s", o.Source, target)

		// Build zfs recv command with properties
		recvArgs := []string{"recv", "-u"} // -u = don't mount
		if o.Mountpoint != "" {
			recvArgs = append(recvArgs, "-o", "mountpoint="+o.Mountpoint)
		}
		if o.CanMount != "" {
			recvArgs = append(recvArgs, "-o", "canmount="+o.CanMount)
		}
		for k, v := range o.Properties {
			recvArgs = append(recvArgs, "-o", fmt.Sprintf("%s=%s", k, v))
		}
		recvArgs = append(recvArgs, target)

		// Pipe: zfs send source | zfs recv target
		send := exec.Command("zfs", "send", "-p", o.Source)
		recv := exec.Command("zfs", recvArgs...)

		pipe, err := send.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create pipe for %s: %w", o.Name, err)
		}
		recv.Stdin = pipe

		// Capture stderr for debugging
		sendErr, _ := send.StderrPipe()
		recvErr, _ := recv.StderrPipe()

		if err := send.Start(); err != nil {
			return fmt.Errorf("failed to start zfs send for %s: %w", o.Source, err)
		}
		if err := recv.Start(); err != nil {
			return fmt.Errorf("failed to start zfs recv for %s: %w", target, err)
		}

		// Wait for both commands to complete
		if err := send.Wait(); err != nil {
			stderr, _ := io.ReadAll(sendErr)
			return fmt.Errorf("zfs send failed for %s: %w\nOutput: %s", o.Source, err, string(stderr))
		}
		if err := recv.Wait(); err != nil {
			stderr, _ := io.ReadAll(recvErr)
			return fmt.Errorf("zfs recv failed for %s: %w\nOutput: %s", target, err, string(stderr))
		}

		b.logger.Debug("Applied dataset overlay: %s", o.Name)
	}

	return nil
}

// createSnapshot creates a ZFS snapshot.
func (b *Builder) createSnapshot(snapshot string) error {
	// On FreeBSD: zfs snapshot snapshot
	b.logger.Debug("Created snapshot: %s", snapshot)
	return nil
}

// exportZFSStream exports a ZFS snapshot as a compressed stream.
func (b *Builder) exportZFSStream(snapshot, outputPath string) error {
	// On FreeBSD: zfs send snapshot | xz -9 > outputPath

	// For prototype, create a dummy compressed file
	content := fmt.Sprintf("# ZFS snapshot: %s\n# Created: %s\n",
		snapshot, time.Now().Format(time.RFC3339))

	if err := util.WriteStringToFile(outputPath, content, 0644); err != nil {
		return err
	}

	b.logger.Debug("Exported ZFS stream to %s", outputPath)
	return nil
}

// exportEFIPartition exports the EFI partition as an image.
func (b *Builder) exportEFIPartition(efiPart, outputPath string) error {
	// On FreeBSD: dd if=efiPart of=outputPath bs=1M

	// For prototype, copy the dummy partition file
	data, err := os.ReadFile(efiPart)
	if err != nil {
		return fmt.Errorf("failed to read EFI partition: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write EFI image: %w", err)
	}

	b.logger.Debug("Exported EFI partition to %s", outputPath)
	return nil
}

// createManifest creates a TOML manifest file.
func (b *Builder) createManifest(cfg config.ImageConfig, manifestPath string) error {
	var sb strings.Builder

	sb.WriteString("# PGSD Image Manifest\n")
	sb.WriteString(fmt.Sprintf("# Generated: %s\n\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("[image]\n")
	sb.WriteString(fmt.Sprintf("id = %q\n", cfg.ID))
	sb.WriteString(fmt.Sprintf("version = %q\n", cfg.Version))
	sb.WriteString(fmt.Sprintf("zpool_name = %q\n", cfg.ZpoolName))
	sb.WriteString(fmt.Sprintf("root_dataset = %q\n", cfg.RootDS))
	sb.WriteString("\n[artifacts]\n")
	sb.WriteString("root_zfs = \"root.zfs.xz\"\n")
	sb.WriteString("efi_image = \"efi.img\"\n")
	sb.WriteString("\n[[package_lists]]\n")
	sb.WriteString(fmt.Sprintf("sets = %s\n", formatStringArray(cfg.PkgLists)))
	sb.WriteString("\n[[overlays]]\n")
	sb.WriteString(fmt.Sprintf("applied = %s\n", formatStringArray(cfg.Overlays)))

	if err := util.WriteStringToFile(manifestPath, sb.String(), 0644); err != nil {
		return err
	}

	b.logger.Debug("Created manifest: %s", manifestPath)
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
