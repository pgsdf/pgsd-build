package config

import (
    "fmt"
    "os"
    "path/filepath"

    lua "github.com/yuin/gopher-lua"
)

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

// LoadImageConfig loads an image configuration from a Lua file.
func LoadImageConfig(path string) (*ImageConfig, error) {
    L := lua.NewState()
    defer L.Close()

    if err := L.DoFile(path); err != nil {
        return nil, fmt.Errorf("failed to execute Lua file %s: %w", path, err)
    }

    // The Lua file should return a table
    ret := L.Get(-1)
    if ret.Type() != lua.LTTable {
        return nil, fmt.Errorf("Lua file %s did not return a table", path)
    }

    tbl := ret.(*lua.LTable)

    cfg := &ImageConfig{
        ID:        getStringField(tbl, "id"),
        Version:   getStringField(tbl, "version"),
        ZpoolName: getStringField(tbl, "zpool_name"),
        RootDS:    getStringField(tbl, "root_dataset"),
        PkgLists:  getStringArrayField(tbl, "pkg_lists"),
        Overlays:  getStringArrayField(tbl, "overlays"),
    }

    if cfg.ID == "" {
        return nil, fmt.Errorf("image config missing required field 'id' in %s", path)
    }

    return cfg, nil
}

// LoadVariantConfig loads a variant configuration from a Lua file.
func LoadVariantConfig(path string) (*VariantConfig, error) {
    L := lua.NewState()
    defer L.Close()

    if err := L.DoFile(path); err != nil {
        return nil, fmt.Errorf("failed to execute Lua file %s: %w", path, err)
    }

    // The Lua file should return a table
    ret := L.Get(-1)
    if ret.Type() != lua.LTTable {
        return nil, fmt.Errorf("Lua file %s did not return a table", path)
    }

    tbl := ret.(*lua.LTable)

    cfg := &VariantConfig{
        ID:        getStringField(tbl, "id"),
        Name:      getStringField(tbl, "name"),
        PkgLists:  getStringArrayField(tbl, "pkg_lists"),
        Overlays:  getStringArrayField(tbl, "overlays"),
        ImagesDir: getStringField(tbl, "images_dir"),
    }

    if cfg.ID == "" {
        return nil, fmt.Errorf("variant config missing required field 'id' in %s", path)
    }

    return cfg, nil
}

// ListImages finds all .lua files in the images directory and loads them.
func ListImages(imagesDir string) ([]*ImageConfig, error) {
    entries, err := os.ReadDir(imagesDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read images directory %s: %w", imagesDir, err)
    }

    var configs []*ImageConfig
    for _, entry := range entries {
        if entry.IsDir() || filepath.Ext(entry.Name()) != ".lua" {
            continue
        }

        path := filepath.Join(imagesDir, entry.Name())
        cfg, err := LoadImageConfig(path)
        if err != nil {
            return nil, fmt.Errorf("failed to load image config %s: %w", path, err)
        }
        configs = append(configs, cfg)
    }

    return configs, nil
}

// ListVariants finds all .lua files in the variants directory and loads them.
func ListVariants(variantsDir string) ([]*VariantConfig, error) {
    entries, err := os.ReadDir(variantsDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read variants directory %s: %w", variantsDir, err)
    }

    var configs []*VariantConfig
    for _, entry := range entries {
        if entry.IsDir() || filepath.Ext(entry.Name()) != ".lua" {
            continue
        }

        path := filepath.Join(variantsDir, entry.Name())
        cfg, err := LoadVariantConfig(path)
        if err != nil {
            return nil, fmt.Errorf("failed to load variant config %s: %w", path, err)
        }
        configs = append(configs, cfg)
    }

    return configs, nil
}

// getStringField extracts a string value from a Lua table.
func getStringField(tbl *lua.LTable, key string) string {
    lv := tbl.RawGetString(key)
    if lv.Type() == lua.LTString {
        return lv.String()
    }
    return ""
}

// getStringArrayField extracts a string array from a Lua table.
func getStringArrayField(tbl *lua.LTable, key string) []string {
    lv := tbl.RawGetString(key)
    if lv.Type() != lua.LTTable {
        return nil
    }

    arr := lv.(*lua.LTable)
    var result []string
    arr.ForEach(func(_, v lua.LValue) {
        if v.Type() == lua.LTString {
            result = append(result, v.String())
        }
    })

    return result
}
