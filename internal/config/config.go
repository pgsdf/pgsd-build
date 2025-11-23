package config

// Placeholder for Lua-based image and variant loading.

type ImageConfig struct {
    ID         string
    Version    string
    ZpoolName  string
    RootDS     string
    PkgLists   []string
    Overlays   []string
}

type VariantConfig struct {
    ID        string
    Name      string
    PkgLists  []string
    Overlays  []string
    ImagesDir string
}
