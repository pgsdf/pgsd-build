package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"

    "github.com/pgsdf/pgsdbuild/internal/config"
    "github.com/pgsdf/pgsdbuild/internal/image"
    "github.com/pgsdf/pgsdbuild/internal/iso"
)

var (
    imagesDir  = "images"
    variantsDir = "variants"
)

func usage() {
    fmt.Println("pgsdbuild - PGSD build tool (prototype)")
    fmt.Println()
    fmt.Println("Usage:")
    fmt.Println("  pgsdbuild image <image-id>")
    fmt.Println("  pgsdbuild iso <variant-id>")
    fmt.Println("  pgsdbuild list-images")
    fmt.Println("  pgsdbuild list-variants")
    fmt.Println()
}

func main() {
    flag.Usage = usage
    flag.Parse()

    args := flag.Args()
    if len(args) < 1 {
        usage()
        os.Exit(1)
    }

    cmd := args[0]
    switch cmd {
    case "image":
        if len(args) < 2 {
            fmt.Println("missing image-id")
            os.Exit(1)
        }
        cmdImage(args[1])
    case "iso":
        if len(args) < 2 {
            fmt.Println("missing variant-id")
            os.Exit(1)
        }
        cmdISO(args[1])
    case "list-images":
        cmdListImages()
    case "list-variants":
        cmdListVariants()
    default:
        fmt.Println("unknown command:", cmd)
        usage()
        os.Exit(1)
    }
}

func cmdImage(imageID string) {
    // Find the image config file
    imagePath := filepath.Join(imagesDir, imageID+".lua")
    if _, err := os.Stat(imagePath); os.IsNotExist(err) {
        fmt.Printf("error: image %q not found at %s\n", imageID, imagePath)
        os.Exit(1)
    }

    // Load the image configuration
    cfg, err := config.LoadImageConfig(imagePath)
    if err != nil {
        fmt.Printf("error loading image config: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Building image: %s (version %s)\n", cfg.ID, cfg.Version)
    fmt.Printf("  ZFS pool: %s\n", cfg.ZpoolName)
    fmt.Printf("  Root dataset: %s\n", cfg.RootDS)
    fmt.Printf("  Package lists: %v\n", cfg.PkgLists)
    fmt.Printf("  Overlays: %v\n", cfg.Overlays)
    fmt.Println()

    // Build the image
    if err := image.BuildImage(*cfg); err != nil {
        fmt.Printf("error building image: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Image %s built successfully\n", cfg.ID)
}

func cmdISO(variantID string) {
    // Find the variant config file
    variantPath := filepath.Join(variantsDir, variantID+".lua")
    if _, err := os.Stat(variantPath); os.IsNotExist(err) {
        fmt.Printf("error: variant %q not found at %s\n", variantID, variantPath)
        os.Exit(1)
    }

    // Load the variant configuration
    cfg, err := config.LoadVariantConfig(variantPath)
    if err != nil {
        fmt.Printf("error loading variant config: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Building bootenv ISO: %s (%s)\n", cfg.ID, cfg.Name)
    fmt.Printf("  Package lists: %v\n", cfg.PkgLists)
    fmt.Printf("  Overlays: %v\n", cfg.Overlays)
    fmt.Printf("  Images dir: %s\n", cfg.ImagesDir)
    fmt.Println()

    // Build the ISO
    if err := iso.BuildBootenvISO(*cfg); err != nil {
        fmt.Printf("error building ISO: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Bootenv ISO %s built successfully\n", cfg.ID)
}

func cmdListImages() {
    configs, err := config.ListImages(imagesDir)
    if err != nil {
        fmt.Printf("error listing images: %v\n", err)
        os.Exit(1)
    }

    if len(configs) == 0 {
        fmt.Println("No images found in", imagesDir)
        return
    }

    fmt.Println("Available images:")
    fmt.Println()
    for _, cfg := range configs {
        fmt.Printf("  %s (version %s)\n", cfg.ID, cfg.Version)
        fmt.Printf("    ZFS pool: %s\n", cfg.ZpoolName)
        fmt.Printf("    Root dataset: %s\n", cfg.RootDS)
        fmt.Printf("    Package lists: %v\n", cfg.PkgLists)
        fmt.Printf("    Overlays: %v\n", cfg.Overlays)
        fmt.Println()
    }
}

func cmdListVariants() {
    configs, err := config.ListVariants(variantsDir)
    if err != nil {
        fmt.Printf("error listing variants: %v\n", err)
        os.Exit(1)
    }

    if len(configs) == 0 {
        fmt.Println("No variants found in", variantsDir)
        return
    }

    fmt.Println("Available variants:")
    fmt.Println()
    for _, cfg := range configs {
        fmt.Printf("  %s (%s)\n", cfg.ID, cfg.Name)
        fmt.Printf("    Package lists: %v\n", cfg.PkgLists)
        fmt.Printf("    Overlays: %v\n", cfg.Overlays)
        fmt.Printf("    Images dir: %s\n", cfg.ImagesDir)
        fmt.Println()
    }
}
