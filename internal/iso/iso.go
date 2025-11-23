package iso

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/pgsdf/pgsdbuild/internal/config"
)

const (
    isoWorkDir  = "work/iso"
    isoOutputDir = "iso"
)

// BuildBootenvISO implements the boot environment ISO build pipeline.
func BuildBootenvISO(cfg config.VariantConfig) error {
    fmt.Printf("[iso] Starting bootenv ISO build for %s\n", cfg.ID)

    // Create working directories
    workPath := filepath.Join(isoWorkDir, cfg.ID)
    if err := os.MkdirAll(workPath, 0755); err != nil {
        return fmt.Errorf("failed to create work directory: %w", err)
    }
    defer os.RemoveAll(workPath) // Clean up work directory

    outputPath := filepath.Join(isoOutputDir, cfg.ID+".iso")
    if err := os.MkdirAll(isoOutputDir, 0755); err != nil {
        return fmt.Errorf("failed to create output directory: %w", err)
    }

    // Step 1: Build root filesystem for ISO
    isoRoot := filepath.Join(workPath, "root")
    if err := os.MkdirAll(isoRoot, 0755); err != nil {
        return fmt.Errorf("failed to create ISO root: %w", err)
    }

    fmt.Println("[iso] Building root filesystem...")
    if err := buildISOFilesystem(cfg, isoRoot); err != nil {
        return fmt.Errorf("failed to build ISO filesystem: %w", err)
    }

    // Step 2: Install packages
    fmt.Println("[iso] Installing packages...")
    if err := installISOPackages(cfg, isoRoot); err != nil {
        return fmt.Errorf("failed to install packages: %w", err)
    }

    // Step 3: Apply overlays
    fmt.Println("[iso] Applying overlays...")
    if err := applyISOOverlays(cfg, isoRoot); err != nil {
        return fmt.Errorf("failed to apply overlays: %w", err)
    }

    // Step 4: Copy system images
    if cfg.ImagesDir != "" {
        fmt.Println("[iso] Copying system images...")
        if err := copySystemImages(cfg, isoRoot); err != nil {
            return fmt.Errorf("failed to copy system images: %w", err)
        }
    }

    // Step 5: Register Arcan target
    fmt.Println("[iso] Registering Arcan installer target...")
    if err := registerArcanTarget(isoRoot); err != nil {
        return fmt.Errorf("failed to register Arcan target: %w", err)
    }

    // Step 6: Assemble ISO image
    fmt.Println("[iso] Assembling ISO image...")
    if err := assembleISO(cfg, isoRoot, outputPath); err != nil {
        return fmt.Errorf("failed to assemble ISO: %w", err)
    }

    fmt.Printf("[iso] ISO build complete! Output: %s\n", outputPath)
    return nil
}

// buildISOFilesystem creates the base directory structure for the ISO.
func buildISOFilesystem(cfg config.VariantConfig, isoRoot string) error {
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
        if err := os.MkdirAll(path, 0755); err != nil {
            return err
        }
    }

    fmt.Println("[iso] Created base filesystem structure")
    return nil
}

// installISOPackages installs packages into the ISO root.
func installISOPackages(cfg config.VariantConfig, isoRoot string) error {
    // On FreeBSD:
    // pkg -r isoRoot install -y <packages>

    // For prototype, create a marker file showing what packages would be installed
    pkgList := filepath.Join(isoRoot, "installed-packages.txt")
    f, err := os.Create(pkgList)
    if err != nil {
        return err
    }
    defer f.Close()

    fmt.Fprintf(f, "# Bootenv ISO: %s\n", cfg.ID)
    fmt.Fprintf(f, "# Package sets installed:\n")
    for _, pkgSet := range cfg.PkgLists {
        fmt.Fprintf(f, "# - %s\n", pkgSet)
    }

    fmt.Printf("[iso] Installed package lists: %v\n", cfg.PkgLists)
    return nil
}

// applyISOOverlays applies filesystem overlays to the ISO root.
func applyISOOverlays(cfg config.VariantConfig, isoRoot string) error {
    for _, overlay := range cfg.Overlays {
        overlayPath := filepath.Join("overlays", overlay)

        // Check if overlay exists
        if _, err := os.Stat(overlayPath); os.IsNotExist(err) {
            return fmt.Errorf("overlay %s not found at %s", overlay, overlayPath)
        }

        // Copy overlay contents to isoRoot
        if err := copyOverlay(overlayPath, isoRoot); err != nil {
            return fmt.Errorf("failed to copy overlay %s: %w", overlay, err)
        }

        fmt.Printf("[iso] Applied overlay: %s\n", overlay)
    }
    return nil
}

// copySystemImages copies built system images into the ISO.
func copySystemImages(cfg config.VariantConfig, isoRoot string) error {
    imagesDestDir := filepath.Join(isoRoot, cfg.ImagesDir[1:]) // Remove leading /
    if err := os.MkdirAll(imagesDestDir, 0755); err != nil {
        return err
    }

    // Look for artifacts in the artifacts directory
    artifactsDir := "artifacts"
    entries, err := os.ReadDir(artifactsDir)
    if err != nil {
        // If no artifacts directory exists, that's okay
        if os.IsNotExist(err) {
            fmt.Println("[iso] No system images found in artifacts/")
            return nil
        }
        return err
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
        if err := copyDir(imageArtifactDir, imageDestDir); err != nil {
            return fmt.Errorf("failed to copy image %s: %w", imageName, err)
        }

        fmt.Printf("[iso] Copied system image: %s\n", imageName)
        imageCount++
    }

    fmt.Printf("[iso] Copied %d system image(s) to %s\n", imageCount, cfg.ImagesDir)
    return nil
}

// registerArcanTarget creates the Arcan target registration metadata.
func registerArcanTarget(isoRoot string) error {
    // On a real system, this would use arcan_db to register the target
    // For prototype, we'll create a marker file

    arcanDir := filepath.Join(isoRoot, "usr/local/share/arcan")
    if err := os.MkdirAll(arcanDir, 0755); err != nil {
        return err
    }

    targetFile := filepath.Join(arcanDir, "pgsd-installer-target.txt")
    f, err := os.Create(targetFile)
    if err != nil {
        return err
    }
    defer f.Close()

    fmt.Fprintf(f, "# PGSD Installer Arcan Target\n")
    fmt.Fprintf(f, "# Registered at: %s\n\n", time.Now().Format(time.RFC3339))
    fmt.Fprintf(f, "target_name: pgsd-installer\n")
    fmt.Fprintf(f, "target_type: BINARY\n")
    fmt.Fprintf(f, "target_path: /usr/local/bin/pgsd-inst\n")
    fmt.Fprintf(f, "config: default\n")

    fmt.Println("[iso] Registered pgsd-installer as Arcan target")
    return nil
}

// assembleISO creates the final ISO image.
func assembleISO(cfg config.VariantConfig, isoRoot, outputPath string) error {
    // On FreeBSD, we would use:
    // makefs -t cd9660 -o rockridge -o label=PGSD_BOOT outputPath isoRoot

    // For prototype, create a marker ISO file
    f, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer f.Close()

    fmt.Fprintf(f, "# PGSD Bootenv ISO (prototype)\n")
    fmt.Fprintf(f, "# Variant: %s (%s)\n", cfg.ID, cfg.Name)
    fmt.Fprintf(f, "# Created: %s\n", time.Now().Format(time.RFC3339))
    fmt.Fprintf(f, "# Package lists: %v\n", cfg.PkgLists)
    fmt.Fprintf(f, "# Overlays: %v\n", cfg.Overlays)
    fmt.Fprintf(f, "# Images dir: %s\n", cfg.ImagesDir)

    fmt.Printf("[iso] Created ISO image: %s\n", outputPath)
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

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
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
