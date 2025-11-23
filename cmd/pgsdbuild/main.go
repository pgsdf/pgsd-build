package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pgsdf/pgsdbuild/internal/build"
	"github.com/pgsdf/pgsdbuild/internal/config"
	"github.com/pgsdf/pgsdbuild/internal/image"
	"github.com/pgsdf/pgsdbuild/internal/iso"
	"github.com/pgsdf/pgsdbuild/internal/util"
)

var (
	// Global build configuration
	buildConfig *build.Config
	logger      *util.Logger
)

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize configuration
	buildConfig = build.NewDefaultConfig()
	buildConfig.LoadFromEnv()

	// Parse global flags
	var (
		verbose  = flag.Bool("v", false, "Enable verbose output")
		quiet    = flag.Bool("q", false, "Suppress all output except errors")
		keepWork = flag.Bool("keep-work", false, "Keep work directory after build")
		version  = flag.Bool("version", false, "Show version information")
		help     = flag.Bool("h", false, "Show help information")

		// Directory flags
		imagesDir    = flag.String("images-dir", buildConfig.ImagesDir, "Directory containing image configurations")
		variantsDir  = flag.String("variants-dir", buildConfig.VariantsDir, "Directory containing variant configurations")
		artifactsDir = flag.String("artifacts-dir", buildConfig.ArtifactsDir, "Directory for build artifacts")
		workDir      = flag.String("work-dir", buildConfig.WorkDir, "Working directory for builds")
		isoDir       = flag.String("iso-dir", buildConfig.ISODir, "Directory for ISO outputs")
	)

	flag.Usage = usage
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Println(VersionInfo())
		return 0
	}

	// Handle help flag
	if *help {
		usage()
		return 0
	}

	// Apply configuration from flags
	buildConfig.ImagesDir = *imagesDir
	buildConfig.VariantsDir = *variantsDir
	buildConfig.ArtifactsDir = *artifactsDir
	buildConfig.WorkDir = *workDir
	buildConfig.ISODir = *isoDir
	buildConfig.KeepWork = *keepWork
	buildConfig.Verbose = *verbose

	// Initialize logger
	logLevel := util.LevelInfo
	if *verbose {
		logLevel = util.LevelDebug
	} else if *quiet {
		logLevel = util.LevelError
	}
	logger = util.NewDefaultLogger(logLevel)
	util.DefaultLogger = logger

	// Validate configuration
	if err := buildConfig.Validate(); err != nil {
		logger.Error("Configuration validation failed: %v", err)
		return 1
	}

	// Parse command
	args := flag.Args()
	if len(args) < 1 {
		usage()
		return 1
	}

	cmd := args[0]
	switch cmd {
	case "image":
		return cmdImage(args[1:])
	case "iso":
		return cmdISO(args[1:])
	case "list-images":
		return cmdListImages(args[1:])
	case "list-variants":
		return cmdListVariants(args[1:])
	case "version":
		fmt.Println(VersionInfo())
		return 0
	case "help":
		usage()
		return 0
	default:
		logger.Error("Unknown command: %s", cmd)
		usage()
		return 1
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "pgsdbuild - PGSD Distribution Build Tool\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  pgsdbuild [flags] <command> [arguments]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  image <image-id>         Build a system image\n")
	fmt.Fprintf(os.Stderr, "  iso <variant-id>         Build a bootable ISO\n")
	fmt.Fprintf(os.Stderr, "  list-images              List available images\n")
	fmt.Fprintf(os.Stderr, "  list-variants            List available variants\n")
	fmt.Fprintf(os.Stderr, "  version                  Show version information\n")
	fmt.Fprintf(os.Stderr, "  help                     Show this help message\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
	fmt.Fprintf(os.Stderr, "  PGSD_IMAGES_DIR          Override images directory\n")
	fmt.Fprintf(os.Stderr, "  PGSD_VARIANTS_DIR        Override variants directory\n")
	fmt.Fprintf(os.Stderr, "  PGSD_ARTIFACTS_DIR       Override artifacts directory\n")
	fmt.Fprintf(os.Stderr, "  PGSD_WORK_DIR            Override work directory\n")
	fmt.Fprintf(os.Stderr, "  PGSD_ISO_DIR             Override ISO directory\n")
	fmt.Fprintf(os.Stderr, "  PGSD_VERBOSE             Enable verbose output (1|true)\n")
	fmt.Fprintf(os.Stderr, "  PGSD_KEEP_WORK           Keep work directory (1|true)\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  pgsdbuild image base\n")
	fmt.Fprintf(os.Stderr, "  pgsdbuild -v iso desktop\n")
	fmt.Fprintf(os.Stderr, "  pgsdbuild --keep-work image server\n")
	fmt.Fprintf(os.Stderr, "  pgsdbuild list-images\n\n")
}

func cmdImage(args []string) int {
	if len(args) < 1 {
		logger.Error("Missing image-id argument")
		fmt.Fprintf(os.Stderr, "Usage: pgsdbuild image <image-id>\n")
		return 1
	}

	imageID := args[0]
	imagePath := filepath.Join(buildConfig.GetImagesDir(), imageID+".lua")

	// Check if image config exists
	if !util.FileExists(imagePath) {
		logger.Error("Image %q not found at %s", imageID, imagePath)
		logger.Info("Run 'pgsdbuild list-images' to see available images")
		return 1
	}

	// Load the image configuration
	logger.Info("Loading image configuration: %s", imageID)
	cfg, err := config.LoadImageConfig(imagePath)
	if err != nil {
		logger.Error("Failed to load image config: %v", err)
		return 1
	}

	// Display build information
	logger.Info("Building image: %s (version %s)", cfg.ID, cfg.Version)
	logger.Debug("  ZFS pool: %s", cfg.ZpoolName)
	logger.Debug("  Root dataset: %s", cfg.RootDS)
	logger.Debug("  Package lists: %v", cfg.PkgLists)
	logger.Debug("  Overlays: %v", cfg.Overlays)

	// Build the image
	builder := image.NewBuilder(buildConfig, logger)
	if err := builder.Build(*cfg); err != nil {
		logger.Error("Image build failed: %v", err)
		return 1
	}

	logger.Info("Image %s built successfully", cfg.ID)
	artifactPath := filepath.Join(buildConfig.GetArtifactsDir(), cfg.ID)
	logger.Info("Artifacts available in: %s", artifactPath)
	return 0
}

func cmdISO(args []string) int {
	if len(args) < 1 {
		logger.Error("Missing variant-id argument")
		fmt.Fprintf(os.Stderr, "Usage: pgsdbuild iso <variant-id>\n")
		return 1
	}

	variantID := args[0]
	variantPath := filepath.Join(buildConfig.GetVariantsDir(), variantID+".lua")

	// Check if variant config exists
	if !util.FileExists(variantPath) {
		logger.Error("Variant %q not found at %s", variantID, variantPath)
		logger.Info("Run 'pgsdbuild list-variants' to see available variants")
		return 1
	}

	// Load the variant configuration
	logger.Info("Loading variant configuration: %s", variantID)
	cfg, err := config.LoadVariantConfig(variantPath)
	if err != nil {
		logger.Error("Failed to load variant config: %v", err)
		return 1
	}

	// Display build information
	logger.Info("Building bootenv ISO: %s (%s)", cfg.ID, cfg.Name)
	logger.Debug("  Package lists: %v", cfg.PkgLists)
	logger.Debug("  Overlays: %v", cfg.Overlays)
	logger.Debug("  Images dir: %s", cfg.ImagesDir)

	// Build the ISO
	builder := iso.NewBuilder(buildConfig, logger)
	if err := builder.Build(*cfg); err != nil {
		logger.Error("ISO build failed: %v", err)
		return 1
	}

	logger.Info("Bootenv ISO %s built successfully", cfg.ID)
	isoPath := filepath.Join(buildConfig.GetISODir(), cfg.ID+".iso")
	logger.Info("ISO available at: %s", isoPath)
	return 0
}

func cmdListImages(args []string) int {
	imagesDir := buildConfig.GetImagesDir()
	logger.Debug("Scanning for images in: %s", imagesDir)

	configs, err := config.ListImages(imagesDir)
	if err != nil {
		logger.Error("Failed to list images: %v", err)
		return 1
	}

	if len(configs) == 0 {
		logger.Info("No images found in %s", imagesDir)
		return 0
	}

	fmt.Printf("Available images (%d):\n\n", len(configs))
	for _, cfg := range configs {
		fmt.Printf("  %s (version %s)\n", cfg.ID, cfg.Version)
		fmt.Printf("    ZFS pool: %s\n", cfg.ZpoolName)
		fmt.Printf("    Root dataset: %s\n", cfg.RootDS)
		if len(cfg.PkgLists) > 0 {
			fmt.Printf("    Package lists: %v\n", cfg.PkgLists)
		}
		if len(cfg.Overlays) > 0 {
			fmt.Printf("    Overlays: %v\n", cfg.Overlays)
		}
		fmt.Println()
	}

	return 0
}

func cmdListVariants(args []string) int {
	variantsDir := buildConfig.GetVariantsDir()
	logger.Debug("Scanning for variants in: %s", variantsDir)

	configs, err := config.ListVariants(variantsDir)
	if err != nil {
		logger.Error("Failed to list variants: %v", err)
		return 1
	}

	if len(configs) == 0 {
		logger.Info("No variants found in %s", variantsDir)
		return 0
	}

	fmt.Printf("Available variants (%d):\n\n", len(configs))
	for _, cfg := range configs {
		fmt.Printf("  %s (%s)\n", cfg.ID, cfg.Name)
		if len(cfg.PkgLists) > 0 {
			fmt.Printf("    Package lists: %v\n", cfg.PkgLists)
		}
		if len(cfg.Overlays) > 0 {
			fmt.Printf("    Overlays: %v\n", cfg.Overlays)
		}
		fmt.Printf("    Images dir: %s\n", cfg.ImagesDir)
		fmt.Println()
	}

	return 0
}
