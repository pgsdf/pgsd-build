package image

import "github.com/pgsdf/pgsdbuild/internal/config"

// BuildImage is a placeholder for the ZFS image build pipeline.
func BuildImage(cfg config.ImageConfig) error {
    // In the real implementation, this would:
    // - create an md-backed disk
    // - partition it
    // - create a ZFS pool and datasets
    // - install packages and overlays
    // - snapshot and export as root.zfs.xz + efi.img + manifest.toml
    return nil
}
