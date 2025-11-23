package iso

import "github.com/pgsdf/pgsdbuild/internal/config"

// BuildBootenvISO is a placeholder for the boot environment ISO build pipeline.
func BuildBootenvISO(cfg config.VariantConfig) error {
    // In the real implementation, this would:
    // - build a root filesystem for the ISO
    // - install Arcan, Durden, and pgsd-inst
    // - copy system images into /usr/local/share/pgsd/images
    // - register the Arcan target
    // - assemble the final ISO image
    return nil
}
