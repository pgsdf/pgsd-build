package iso

import (
	"archive/tar"
	"compress/gzip"
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

// Builder builds bootable ISO images.
type Builder struct {
	config      *build.Config
	logger      *util.Logger
	freebsdRoot string // Root directory for FreeBSD files (for cross-building)
}

// NewBuilder creates a new ISO Builder.
func NewBuilder(cfg *build.Config, logger *util.Logger) *Builder {
	// Support cross-building from non-FreeBSD systems
	// Set FREEBSD_ROOT env var to point to extracted FreeBSD distribution
	freebsdRoot := os.Getenv("FREEBSD_ROOT")
	if freebsdRoot == "" {
		freebsdRoot = "/" // Default to system root for native builds
	}

	return &Builder{
		config:      cfg,
		logger:      logger,
		freebsdRoot: freebsdRoot,
	}
}

// Build implements the boot environment ISO build pipeline.
func (b *Builder) Build(cfg config.VariantConfig) error {
	b.logger.Info("Starting bootenv ISO build for %s", cfg.ID)

	// Create working directories
	workPath := filepath.Join(b.config.GetWorkDir(), "iso", cfg.ID)
	if err := util.EnsureDir(workPath); err != nil {
		return err
	}

	if !b.config.KeepWork {
		defer util.CleanupDir(workPath)
	}

	outputPath := filepath.Join(b.config.GetISODir(), cfg.ID+".iso")
	if err := util.EnsureDir(b.config.GetISODir()); err != nil {
		return err
	}

	// Step 1: Build root filesystem for ISO
	isoRoot := filepath.Join(workPath, "root")
	if err := util.EnsureDir(isoRoot); err != nil {
		return err
	}

	b.logger.Debug("Building root filesystem...")
	if err := b.buildISOFilesystem(cfg, isoRoot); err != nil {
		return fmt.Errorf("failed to build ISO filesystem: %w", err)
	}

	// Step 2: Install packages
	b.logger.Debug("Installing packages...")
	if err := b.installISOPackages(cfg, isoRoot); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	// Step 2.5: Install FreeBSD boot infrastructure
	b.logger.Debug("Installing boot infrastructure...")
	if err := b.installBootInfrastructure(isoRoot); err != nil {
		return fmt.Errorf("failed to install boot infrastructure: %w", err)
	}

	// Step 3: Apply overlays
	b.logger.Debug("Applying overlays...")
	if err := b.applyISOOverlays(cfg, isoRoot); err != nil {
		return fmt.Errorf("failed to apply overlays: %w", err)
	}

	// Step 4: Copy system images
	if cfg.ImagesDir != "" {
		b.logger.Debug("Copying system images...")
		if err := b.copySystemImages(cfg, isoRoot); err != nil {
			return fmt.Errorf("failed to copy system images: %w", err)
		}
	}

	// Step 5: Register Arcan target
	b.logger.Debug("Registering Arcan installer target...")
	if err := b.registerArcanTarget(isoRoot); err != nil {
		return fmt.Errorf("failed to register Arcan target: %w", err)
	}

	// Step 6: Assemble ISO image
	b.logger.Debug("Assembling ISO image...")
	if err := b.assembleISO(cfg, isoRoot, outputPath); err != nil {
		return fmt.Errorf("failed to assemble ISO: %w", err)
	}

	b.logger.Info("ISO build complete! Output: %s", outputPath)
	return nil
}

// buildISOFilesystem creates the base directory structure for the ISO.
func (b *Builder) buildISOFilesystem(cfg config.VariantConfig, isoRoot string) error {
	// Create standard FreeBSD directory structure
	dirs := []string{
		"bin", "boot", "dev", "etc", "lib", "libexec",
		"mnt", "proc", "rescue", "root", "sbin", "tmp",
		"usr/bin", "usr/lib", "usr/local/bin", "usr/local/etc",
		"usr/local/lib", "usr/local/share", "usr/sbin",
		"usr/share", "var/log", "var/run", "var/tmp",
	}

	for _, dir := range dirs {
		path := filepath.Join(isoRoot, dir)
		if err := util.EnsureDir(path); err != nil {
			return err
		}
	}

	b.logger.Debug("Created base filesystem structure")
	return nil
}

// installISOPackages installs packages into the ISO root.
func (b *Builder) installISOPackages(cfg config.VariantConfig, isoRoot string) error {
	b.logger.Info("Installing FreeBSD base system for bootable ISO...")

	// For bootable ISOs, copy essential FreeBSD base system from FREEBSD_ROOT
	// This supports both native FreeBSD builds and cross-building from Linux
	essentialDirs := []string{
		"bin",
		"sbin",
		"lib",
		"libexec",
		"usr/bin",
		"usr/sbin",
		"usr/lib",
		"usr/libexec",
		"rescue",
	}

	b.logger.Info("Copying essential base system directories from %s...", b.freebsdRoot)
	for _, dir := range essentialDirs {
		srcDir := filepath.Join(b.freebsdRoot, dir)
		destDir := filepath.Join(isoRoot, dir)

		if _, err := os.Stat(srcDir); err != nil {
			if os.IsNotExist(err) {
				b.logger.Warn("Source directory not found (skipping): %s", srcDir)
				continue
			}
			return fmt.Errorf("failed to access %s: %w", srcDir, err)
		}

		b.logger.Debug("Copying: %s -> %s", srcDir, destDir)
		if err := util.CopyDir(srcDir, destDir); err != nil {
			b.logger.Warn("Failed to copy %s: %v (continuing)", srcDir, err)
			// Don't fail build, continue with other dirs
		}
	}

	// Ensure critical system directories exist
	criticalDirs := []string{
		filepath.Join(isoRoot, "dev"),
		filepath.Join(isoRoot, "tmp"),
		filepath.Join(isoRoot, "var"),
		filepath.Join(isoRoot, "var/run"),
		filepath.Join(isoRoot, "var/log"),
		filepath.Join(isoRoot, "root"),
		filepath.Join(isoRoot, "proc"),
		filepath.Join(isoRoot, "mnt"),
	}

	for _, dir := range criticalDirs {
		if err := util.EnsureDir(dir); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	b.logger.Info("FreeBSD base system installed successfully")
	return nil
}

// applyISOOverlays applies filesystem overlays to the ISO root.
func (b *Builder) applyISOOverlays(cfg config.VariantConfig, isoRoot string) error {
	overlaysDir := b.config.GetOverlaysDir()

	for _, overlay := range cfg.Overlays {
		overlayPath := filepath.Join(overlaysDir, overlay)

		// Check if overlay exists
		if !util.DirExists(overlayPath) {
			return fmt.Errorf("overlay %s not found at %s", overlay, overlayPath)
		}

		// Copy overlay contents to isoRoot
		if err := util.CopyOverlay(overlayPath, isoRoot); err != nil {
			return fmt.Errorf("failed to copy overlay %s: %w", overlay, err)
		}

		b.logger.Debug("Applied overlay: %s", overlay)
	}
	return nil
}

// copySystemImages copies built system images into the ISO.
func (b *Builder) copySystemImages(cfg config.VariantConfig, isoRoot string) error {
	imagesDestDir := filepath.Join(isoRoot, cfg.ImagesDir[1:]) // Remove leading /
	if err := util.EnsureDir(imagesDestDir); err != nil {
		return err
	}

	// Look for artifacts in the artifacts directory
	artifactsDir := b.config.GetArtifactsDir()
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		// If no artifacts directory exists, that's okay
		if os.IsNotExist(err) {
			b.logger.Warn("No system images found in %s", artifactsDir)
			return nil
		}
		return fmt.Errorf("failed to read artifacts directory: %w", err)
	}

	imageCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		imageName := entry.Name()
		imageArtifactDir := filepath.Join(artifactsDir, imageName)
		imageDestDir := filepath.Join(imagesDestDir, imageName)

		// Copy the entire artifact directory
		if err := util.CopyDir(imageArtifactDir, imageDestDir); err != nil {
			return fmt.Errorf("failed to copy image %s: %w", imageName, err)
		}

		b.logger.Debug("Copied system image: %s", imageName)
		imageCount++
	}

	b.logger.Info("Copied %d system image(s) to %s", imageCount, cfg.ImagesDir)
	return nil
}

// installBootInfrastructure installs FreeBSD boot files needed for bootable ISO
func (b *Builder) installBootInfrastructure(isoRoot string) error {
	// Copy essential boot files from FREEBSD_ROOT
	// This makes the ISO bootable in both BIOS and UEFI modes

	bootDir := filepath.Join(isoRoot, "boot")

	// Required boot files for BIOS boot (relative paths from FREEBSD_ROOT)
	biosBootFiles := []string{
		"boot/cdboot",               // CD/DVD boot loader (required for El Torito)
		"boot/isoboot",              // Hybrid ISO boot (required for USB boot support)
		"boot/loader",               // Boot loader (stage 3)
		"boot/loader.rc",            // Loader configuration
		"boot/defaults/loader.conf", // Default loader settings
	}

	b.logger.Debug("Copying BIOS boot files from %s...", b.freebsdRoot)
	for _, relPath := range biosBootFiles {
		srcPath := filepath.Join(b.freebsdRoot, relPath)

		if _, err := os.Stat(srcPath); err != nil {
			if os.IsNotExist(err) {
				b.logger.Warn("Boot file not found (skipping): %s", srcPath)
				continue
			}
			return fmt.Errorf("failed to access boot file %s: %w", srcPath, err)
		}

		// Determine destination path
		bootRelPath := strings.TrimPrefix(relPath, "boot/")
		dstPath := filepath.Join(bootDir, bootRelPath)

		// Ensure destination directory exists
		dstDir := filepath.Dir(dstPath)
		if err := util.EnsureDir(dstDir); err != nil {
			return err
		}

		// Copy the file
		if err := util.CopyFile(srcPath, dstPath, 0644); err != nil {
			return fmt.Errorf("failed to copy %s: %w", srcPath, err)
		}
	}

	// Copy kernel (REQUIRED for boot)
	b.logger.Info("Copying FreeBSD kernel (required for boot)...")
	kernelSrc := filepath.Join(b.freebsdRoot, "boot/kernel/kernel")
	if _, err := os.Stat(kernelSrc); err != nil {
		return fmt.Errorf("kernel not found at %s - cannot create bootable ISO: %w", kernelSrc, err)
	}

	kernelDir := filepath.Join(bootDir, "kernel")
	if err := util.EnsureDir(kernelDir); err != nil {
		return fmt.Errorf("failed to create kernel directory: %w", err)
	}

	kernelDst := filepath.Join(kernelDir, "kernel")
	b.logger.Debug("Copying kernel: %s -> %s", kernelSrc, kernelDst)
	if err := util.CopyFile(kernelSrc, kernelDst, 0755); err != nil {
		return fmt.Errorf("failed to copy kernel: %w", err)
	}

	// Verify kernel was copied
	if info, err := os.Stat(kernelDst); err != nil {
		return fmt.Errorf("kernel copy verification failed: %w", err)
	} else {
		b.logger.Info("Kernel copied successfully (%d bytes)", info.Size())
	}

	// Copy essential kernel modules
	b.logger.Debug("Copying kernel modules...")
	modules := []string{
		"zfs.ko",        // ZFS filesystem
		"geom_label.ko", // GEOM labels
		"ahci.ko",       // AHCI disk controller
	}

	for _, module := range modules {
		srcPath := filepath.Join(b.freebsdRoot, "boot/kernel", module)
		if _, err := os.Stat(srcPath); err == nil {
			dstPath := filepath.Join(bootDir, "kernel", module)
			if err := util.CopyFile(srcPath, dstPath, 0644); err != nil {
				b.logger.Warn("Failed to copy module %s: %v", module, err)
			}
		}
	}

	// Set up EFI boot directory for UEFI support
	b.logger.Debug("Setting up EFI boot...")
	efiDir := filepath.Join(isoRoot, "EFI", "BOOT")
	if err := util.EnsureDir(efiDir); err != nil {
		return err
	}

	// Copy EFI boot loader if available
	efiBootSrc := filepath.Join(b.freebsdRoot, "boot/boot1.efi")
	if _, err := os.Stat(efiBootSrc); err == nil {
		efiBootDst := filepath.Join(efiDir, "BOOTX64.EFI")
		if err := util.CopyFile(efiBootSrc, efiBootDst, 0644); err != nil {
			b.logger.Warn("Failed to copy EFI bootloader: %v", err)
		}
	} else {
		b.logger.Warn("EFI bootloader not found - UEFI boot may not work")
	}

	b.logger.Info("Boot infrastructure installed")
	return nil
}

// registerArcanTarget creates the Arcan target registration metadata.
func (b *Builder) registerArcanTarget(isoRoot string) error {
	// On a real system, this would use arcan_db to register the target
	// For prototype, we'll create a marker file

	arcanDir := filepath.Join(isoRoot, "usr/local/share/arcan")
	if err := util.EnsureDir(arcanDir); err != nil {
		return err
	}

	targetFile := filepath.Join(arcanDir, "pgsd-installer-target.txt")
	content := fmt.Sprintf("# PGSD Installer Arcan Target\n"+
		"# Registered at: %s\n\n"+
		"target_name: pgsd-installer\n"+
		"target_type: BINARY\n"+
		"target_path: /usr/local/bin/pgsd-inst\n"+
		"config: default\n",
		time.Now().Format(time.RFC3339))

	if err := util.WriteStringToFile(targetFile, content, 0644); err != nil {
		return err
	}

	b.logger.Debug("Registered pgsd-installer as Arcan target")
	return nil
}

// assembleISO creates the final ISO image.
func (b *Builder) assembleISO(cfg config.VariantConfig, isoRoot, outputPath string) error {
	// Create bootable ISO using makefs
	// This requires FreeBSD with makefs utility

	// Determine ISO label (max 32 characters, alphanumeric only for cd9660)
	// makefs cd9660 requires "d-characters" (digits/letters) only in labels
	label := cfg.ID
	label = strings.ReplaceAll(label, "-", "")
	label = strings.ReplaceAll(label, "_", "")
	if len(label) > 32 {
		label = label[:32]
	}
	// Ensure label is uppercase for consistency
	label = strings.ToUpper(label)

	b.logger.Debug("Creating ISO filesystem...")
	b.logger.Debug("ISO label: %s", label)

	// Check if cdboot file exists before configuring boot
	cdbootPath := filepath.Join(isoRoot, "boot/cdboot")
	hasCdboot := false
	if stat, err := os.Stat(cdbootPath); err == nil && !stat.IsDir() {
		hasCdboot = true
		b.logger.Info("BIOS boot file found: boot/cdboot (size: %d bytes)", stat.Size())
	} else {
		b.logger.Warn("BIOS boot file not found at %s - ISO will NOT be bootable in BIOS mode", cdbootPath)
		b.logger.Warn("To create bootable ISOs, this must run on FreeBSD with boot files installed")
	}

	// Detect which ISO creation tool is available
	isoTool, isoToolPath := b.detectISOTool()
	if isoTool == "" {
		b.logger.Warn("No ISO creation tool found (tried: makefs, xorriso, genisoimage, mkisofs)")
		b.logger.Warn("Creating tar archive instead - convert to ISO on a system with ISO tools")

		// Fallback: create a tar.gz archive of the ISO contents
		tarPath := strings.TrimSuffix(outputPath, ".iso") + ".tar.gz"
		if err := b.createTarArchive(tarPath, isoRoot); err != nil {
			return fmt.Errorf("tar archive creation failed: %w", err)
		}

		b.logger.Info("Created tar archive: %s", tarPath)
		b.logger.Info("To convert to ISO, use: genisoimage -r -V %s -o %s -graft-points <extracted-contents>", label, outputPath)

		return nil
	}

	b.logger.Info("Using ISO creation tool: %s (%s)", isoTool, isoToolPath)

	// Convert paths to absolute to avoid working directory issues
	absIsoRoot, err := filepath.Abs(isoRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for ISO root: %w", err)
	}

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for output: %w", err)
	}

	b.logger.Debug("ISO root (absolute): %s", absIsoRoot)
	b.logger.Debug("Output path (absolute): %s", absOutputPath)

	// Create ISO using the detected tool
	if err := b.createISOWithTool(isoTool, isoToolPath, absOutputPath, absIsoRoot, label, hasCdboot); err != nil {
		return fmt.Errorf("%s failed: %w", isoTool, err)
	}

	b.logger.Debug("Created ISO image: %s", outputPath)

	// Verify the ISO was created
	if info, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("ISO verification failed: %w", err)
	} else {
		b.logger.Info("ISO size: %.2f MB", float64(info.Size())/(1024*1024))
	}

	return nil
}

// detectISOTool detects which ISO creation tool is available on the system
// Returns the tool name and full path to the executable
func (b *Builder) detectISOTool() (string, string) {
	// Try common tools with explicit paths for FreeBSD/Linux systems
	toolPaths := map[string][]string{
		"makefs":      {"/usr/sbin/makefs", "/sbin/makefs"},
		"xorriso":     {"/usr/bin/xorriso"},
		"genisoimage": {"/usr/bin/genisoimage"},
		"mkisofs":     {"/usr/bin/mkisofs"},
	}

	// Check explicit paths first (for FreeBSD where tools may not be in PATH)
	for tool, paths := range toolPaths {
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				b.logger.Debug("Found ISO tool at %s", path)
				return tool, path
			}
		}
	}

	// Fall back to PATH search
	tools := []string{"makefs", "xorriso", "genisoimage", "mkisofs"}
	for _, tool := range tools {
		if path, err := exec.LookPath(tool); err == nil {
			return tool, path
		}
	}

	return "", ""
}

// createISOWithTool creates an ISO using the specified tool
func (b *Builder) createISOWithTool(tool, toolPath, outputPath, isoRoot, label string, hasCdboot bool) error {
	switch tool {
	case "makefs":
		return b.createISOWithMakefs(toolPath, outputPath, isoRoot, label, hasCdboot)
	case "xorriso":
		return b.createISOWithXorriso(toolPath, outputPath, isoRoot, label, hasCdboot)
	case "genisoimage", "mkisofs":
		return b.createISOWithGenisoimage(toolPath, outputPath, isoRoot, label, hasCdboot)
	default:
		return fmt.Errorf("unsupported ISO tool: %s", tool)
	}
}

// createISOWithMakefs creates an ISO using FreeBSD's makefs utility
func (b *Builder) createISOWithMakefs(toolPath, outputPath, isoRoot, label string, hasCdboot bool) error {
	// Verify boot files exist in isoRoot before calling makefs
	if hasCdboot {
		bootPath := filepath.Join(isoRoot, "boot/cdboot")
		if stat, err := os.Stat(bootPath); err != nil {
			b.logger.Error("Boot file check failed: %v", err)
			b.logger.Error("Expected boot file at: %s", bootPath)
			return fmt.Errorf("boot file not found at %s: %w", bootPath, err)
		} else {
			b.logger.Debug("Verified boot file exists: %s (size: %d bytes)", bootPath, stat.Size())
		}

		// List contents of boot directory for debugging
		bootDir := filepath.Join(isoRoot, "boot")
		if entries, err := os.ReadDir(bootDir); err == nil {
			b.logger.Debug("Contents of boot directory:")
			for _, entry := range entries {
				info, _ := entry.Info()
				if info != nil {
					b.logger.Debug("  - %s (%d bytes)", entry.Name(), info.Size())
				} else {
					b.logger.Debug("  - %s", entry.Name())
				}
			}
		}
	}

	// Build makefs command arguments
	// -t cd9660: ISO 9660 filesystem
	// -o rockridge (R): Rock Ridge extensions (long filenames, permissions)
	// -o L=<label>: Volume label (must be d-characters: alphanumeric only)
	// -o B=<bootimage>: Boot image for BIOS boot
	// -o no-emul-boot: No emulation boot mode
	// -o no-trailing-padding: Omit padding for smaller file
	args := []string{
		"-t", "cd9660",
		"-o", "rockridge",
		"-o", fmt.Sprintf("L=%s", label),
	}

	// Add boot image configuration if cdboot exists
	if hasCdboot {
		args = append(args,
			"-o", "B=i386;boot/cdboot", // Boot image specification
			"-o", "no-emul-boot", // No emulation mode
		)
		b.logger.Info("Configured for BIOS boot with boot/cdboot")
	} else {
		b.logger.Info("Creating non-bootable ISO (no boot files available)")
	}

	// makefs needs to run from the isoRoot directory so it can find boot/cdboot
	// The -B option expects paths relative to the working directory
	// So we:
	//   1. Change to isoRoot directory
	//   2. Use "." as the source directory (current dir = isoRoot)
	//   3. Make outputPath relative to isoRoot or absolute

	// Ensure outputPath is absolute or relative to isoRoot
	finalOutputPath := outputPath
	if !filepath.IsAbs(outputPath) {
		// outputPath is relative to current dir, need to make it relative to isoRoot
		// or convert to absolute
		absOutput, err := filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for output: %w", err)
		}
		finalOutputPath = absOutput
	}

	// Add final options and paths
	args = append(args,
		"-o", "no-trailing-padding",
		finalOutputPath, // Absolute path to output ISO
		".",             // Source directory (current dir = isoRoot)
	)

	b.logger.Debug("Calling makefs from isoRoot: %s", isoRoot)
	b.logger.Debug("Output will be written to: %s", finalOutputPath)

	if err := b.runCommandInDir(toolPath, isoRoot, args...); err != nil {
		return err
	}

	// For USB boot support, add MBR boot code to create a hybrid ISO
	// This makes the ISO bootable from both CD/DVD and USB drives
	if hasCdboot {
		if err := b.addMBRBootCode(finalOutputPath, isoRoot); err != nil {
			b.logger.Warn("Failed to add MBR boot code (USB boot may not work): %v", err)
			b.logger.Info("ISO is still bootable from CD/DVD")
		} else {
			b.logger.Info("Created hybrid ISO (bootable from CD/DVD and USB)")
		}
	}

	return nil
}

// addMBRBootCode adds MBR boot code to an ISO to make it USB-bootable (hybrid ISO)
// This is how FreeBSD creates hybrid ISOs that boot from both CD/DVD and USB
func (b *Builder) addMBRBootCode(isoPath, isoRoot string) error {
	// FreeBSD uses /boot/isoboot for hybrid ISOs
	// isoboot contains the MBR boot code that allows USB boot
	isobootPaths := []string{
		filepath.Join(b.freebsdRoot, "boot/isoboot"), // From FREEBSD_ROOT
		filepath.Join(isoRoot, "boot/isoboot"),       // From ISO root
		filepath.Join(b.freebsdRoot, "boot/cdboot"),  // Fallback
	}

	var isobootPath string
	for _, path := range isobootPaths {
		if _, err := os.Stat(path); err == nil {
			isobootPath = path
			b.logger.Debug("Found isoboot at: %s", path)
			break
		}
	}

	if isobootPath == "" {
		return fmt.Errorf("isoboot file not found (checked: %v)", isobootPaths)
	}

	// Read the isoboot file (contains MBR boot code)
	isoboot, err := os.ReadFile(isobootPath)
	if err != nil {
		return fmt.Errorf("failed to read isoboot: %w", err)
	}

	// Open the ISO file for writing at the beginning
	iso, err := os.OpenFile(isoPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ISO: %w", err)
	}
	defer iso.Close()

	// Write the first 432 bytes of isoboot to the ISO (MBR boot code)
	// This doesn't overwrite the ISO 9660 structures, only adds boot code
	bootCodeSize := 432
	if len(isoboot) < bootCodeSize {
		bootCodeSize = len(isoboot)
	}

	n, err := iso.WriteAt(isoboot[:bootCodeSize], 0)
	if err != nil {
		return fmt.Errorf("failed to write boot code: %w", err)
	}
	b.logger.Debug("Wrote %d bytes of MBR boot code to ISO", n)

	// Get ISO file size for partition table
	isoInfo, err := iso.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat ISO: %w", err)
	}
	isoSize := isoInfo.Size()
	isoSectors := (isoSize + 511) / 512 // Round up to sector boundary

	// Create MBR partition table for USB boot detection
	// Partition table starts at byte 446 (0x1BE)
	partitionTable := make([]byte, 64) // 4 partition entries * 16 bytes each

	// First partition entry (bytes 0-15 of partition table)
	// This tells BIOS where the bootable data is
	partitionTable[0] = 0x80 // Bootable flag (0x80 = bootable, 0x00 = not bootable)
	partitionTable[1] = 0x00 // Starting head
	partitionTable[2] = 0x01 // Starting sector
	partitionTable[3] = 0x00 // Starting cylinder
	partitionTable[4] = 0xcd // Partition type (0xcd = ISO 9660 filesystem)
	partitionTable[5] = 0xFE // Ending head
	partitionTable[6] = 0xFF // Ending sector
	partitionTable[7] = 0xFF // Ending cylinder

	// LBA of first sector (little-endian, 4 bytes)
	partitionTable[8] = 0x00
	partitionTable[9] = 0x00
	partitionTable[10] = 0x00
	partitionTable[11] = 0x00

	// Number of sectors (little-endian, 4 bytes)
	// Limit to 32-bit max for compatibility
	sectors := uint32(isoSectors)
	if isoSectors > 0xFFFFFFFF {
		sectors = 0xFFFFFFFF
	}
	partitionTable[12] = byte(sectors & 0xFF)
	partitionTable[13] = byte((sectors >> 8) & 0xFF)
	partitionTable[14] = byte((sectors >> 16) & 0xFF)
	partitionTable[15] = byte((sectors >> 24) & 0xFF)

	// Write partition table at offset 446 (0x1BE)
	if _, err := iso.WriteAt(partitionTable, 446); err != nil {
		return fmt.Errorf("failed to write partition table: %w", err)
	}
	b.logger.Debug("Created MBR partition table (ISO size: %d MB, sectors: %d)", isoSize/(1024*1024), sectors)

	// Write MBR boot signature at bytes 510-511 (0x55 0xAA)
	bootSig := []byte{0x55, 0xAA}
	if _, err := iso.WriteAt(bootSig, 510); err != nil {
		return fmt.Errorf("failed to write boot signature: %w", err)
	}
	b.logger.Debug("Wrote MBR boot signature (0x55AA)")

	return nil
}

// createISOWithXorriso creates an ISO using xorriso (modern Linux ISO tool)
func (b *Builder) createISOWithXorriso(toolPath, outputPath, isoRoot, label string, hasCdboot bool) error {
	// xorriso command arguments for creating a hybrid BIOS+UEFI bootable ISO
	// This creates an ISO that boots from CD/DVD/USB in both BIOS and UEFI modes
	// -as mkisofs: Compatibility mode
	// -r: Rock Ridge extensions
	// -J: Joliet extensions (Windows compatibility)
	// -joliet-long: Long filenames in Joliet
	// -V <label>: Volume label
	// -o <output>: Output file
	args := []string{
		"-as", "mkisofs",
		"-r",            // Rock Ridge
		"-J",            // Joliet
		"-joliet-long",  // Long filenames
		"-cache-inodes", // Optimize hard links
		"-V", label,     // Volume label
		"-o", outputPath, // Output file
	}

	if hasCdboot {
		// Check if EFI bootloader exists for UEFI support
		efiBootPath := filepath.Join(isoRoot, "EFI/BOOT/BOOTX64.EFI")
		hasEFI := false
		if stat, err := os.Stat(efiBootPath); err == nil && !stat.IsDir() {
			hasEFI = true
		}

		// Check if isoboot exists for hybrid USB boot
		isobootPath := filepath.Join(isoRoot, "boot/isoboot")
		hasIsoboot := false
		if stat, err := os.Stat(isobootPath); err == nil && !stat.IsDir() {
			hasIsoboot = true
		}

		// Add hybrid MBR boot code for USB boot (BIOS mode)
		if hasIsoboot {
			args = append(args, "-isohybrid-mbr", filepath.Join(isoRoot, "boot/isoboot"))
			b.logger.Info("Configured for hybrid USB boot (BIOS mode)")
		}

		// Add BIOS boot configuration (CD/DVD and USB)
		args = append(args,
			"-b", "boot/cdboot", // Boot image
			"-c", "boot.catalog", // Boot catalog
			"-boot-load-size", "4", // Load size
			"-boot-info-table", // Create boot info table
			"-no-emul-boot",    // No emulation
		)
		b.logger.Info("Configured for BIOS boot with boot/cdboot")

		// Add UEFI boot configuration
		if hasEFI {
			args = append(args,
				"-eltorito-alt-boot",         // Alternative boot entry
				"-e", "EFI/BOOT/BOOTX64.EFI", // EFI boot loader
				"-no-emul-boot",         // No emulation
				"-isohybrid-gpt-basdat", // GPT partition for hybrid boot
			)
			b.logger.Info("Configured for UEFI boot with EFI/BOOT/BOOTX64.EFI")
		}
	} else {
		b.logger.Info("Creating non-bootable ISO (no boot files available)")
	}

	args = append(args, isoRoot)

	return b.runCommand(toolPath, args...)
}

// createISOWithGenisoimage creates an ISO using genisoimage or mkisofs (legacy Linux tools)
func (b *Builder) createISOWithGenisoimage(toolPath, outputPath, isoRoot, label string, hasCdboot bool) error {
	// genisoimage/mkisofs command arguments
	// -r: Rock Ridge extensions
	// -V <label>: Volume label
	// -o <output>: Output file
	args := []string{
		"-r",        // Rock Ridge
		"-V", label, // Volume label
		"-o", outputPath, // Output file
	}

	if hasCdboot {
		args = append(args,
			"-b", "boot/cdboot", // Boot image
			"-no-emul-boot",        // No emulation
			"-boot-load-size", "4", // Load size
			"-boot-info-table", // Create boot info table
		)
		b.logger.Info("Configured for BIOS boot with boot/cdboot")
	} else {
		b.logger.Info("Creating non-bootable ISO (no boot files available)")
	}

	args = append(args, isoRoot)

	return b.runCommand(toolPath, args...)
}

// createTarArchive creates a compressed tar archive of the ISO root directory
func (b *Builder) createTarArchive(tarPath, isoRoot string) error {
	b.logger.Debug("Creating tar archive: %s", tarPath)

	// Create the output file
	outFile, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk the ISO root directory and add all files
	return filepath.Walk(isoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path for the tar header
		relPath, err := filepath.Rel(isoRoot, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header from file info
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", path, err)
		}

		// Use the relative path as the name in the archive
		header.Name = relPath

		// Write the header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", path, err)
		}

		// If it's a regular file, write its contents
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("failed to write file %s to tar: %w", path, err)
			}
		}

		return nil
	})
}

// runCommand executes a command and returns an error if it fails
func (b *Builder) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	b.logger.Debug("Running: %s %v", name, args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		b.logger.Debug("Command output: %s", string(output))
	}

	return nil
}

// runCommandInDir executes a command in a specific working directory
func (b *Builder) runCommandInDir(name, dir string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	b.logger.Debug("Running in %s: %s %v", dir, name, args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		b.logger.Debug("Command output: %s", string(output))
	}

	return nil
}
