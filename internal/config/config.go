package config

import (
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

type ImageConfig struct {
	ID              string
	Version         string
	ZpoolName       string
	RootDS          string
	PkgLists        []string
	Overlays        []string
	DatasetOverlays []DatasetOverlay
}

// DatasetOverlay represents a ZFS dataset snapshot to receive into the image.
type DatasetOverlay struct {
	Name       string
	Source     string
	Mountpoint string
	CanMount   string
	Properties map[string]string
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
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("image config file not found: %s", path)
		}
		return nil, fmt.Errorf("cannot access image config file %s: %w", path, err)
	}

	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile(path); err != nil {
		return nil, fmt.Errorf("failed to parse Lua config %s: %w\nHint: Check for syntax errors in the Lua file", path, err)
	}

	// The Lua file should return a table
	ret := L.Get(-1)
	if ret.Type() != lua.LTTable {
		return nil, fmt.Errorf("invalid image config %s: must return a Lua table\nExample: return { id = \"my-image\", ... }", path)
	}

	tbl := ret.(*lua.LTable)

	cfg := &ImageConfig{
		ID:              getStringField(tbl, "id"),
		Version:         getStringField(tbl, "version"),
		ZpoolName:       getStringField(tbl, "zpool_name"),
		RootDS:          getStringField(tbl, "root_dataset"),
		PkgLists:        getStringArrayField(tbl, "pkg_lists"),
		Overlays:        getStringArrayField(tbl, "overlays"),
		DatasetOverlays: getDatasetOverlays(tbl, "dataset_overlays"),
	}

	// Validate required fields
	if err := validateImageConfig(cfg, path); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadVariantConfig loads a variant configuration from a Lua file.
func LoadVariantConfig(path string) (*VariantConfig, error) {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("variant config file not found: %s", path)
		}
		return nil, fmt.Errorf("cannot access variant config file %s: %w", path, err)
	}

	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile(path); err != nil {
		return nil, fmt.Errorf("failed to parse Lua config %s: %w\nHint: Check for syntax errors in the Lua file", path, err)
	}

	// The Lua file should return a table
	ret := L.Get(-1)
	if ret.Type() != lua.LTTable {
		return nil, fmt.Errorf("invalid variant config %s: must return a Lua table\nExample: return { id = \"my-variant\", ... }", path)
	}

	tbl := ret.(*lua.LTable)

	cfg := &VariantConfig{
		ID:        getStringField(tbl, "id"),
		Name:      getStringField(tbl, "name"),
		PkgLists:  getStringArrayField(tbl, "pkg_lists"),
		Overlays:  getStringArrayField(tbl, "overlays"),
		ImagesDir: getStringField(tbl, "images_dir"),
	}

	// Validate required fields
	if err := validateVariantConfig(cfg, path); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ListImages finds all .lua files in the images directory and loads them.
func ListImages(imagesDir string) ([]*ImageConfig, error) {
	// Check if directory exists
	if _, err := os.Stat(imagesDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("images directory not found: %s\nHint: Create the directory and add image .lua files", imagesDir)
		}
		return nil, fmt.Errorf("cannot access images directory %s: %w", imagesDir, err)
	}

	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read images directory %s: %w", imagesDir, err)
	}

	var configs []*ImageConfig
	var errors []string

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".lua" {
			continue
		}

		path := filepath.Join(imagesDir, entry.Name())
		cfg, err := LoadImageConfig(path)
		if err != nil {
			// Collect all errors instead of failing on first one
			errors = append(errors, fmt.Sprintf("  - %s: %v", entry.Name(), err))
			continue
		}
		configs = append(configs, cfg)
	}

	if len(errors) > 0 {
		return configs, fmt.Errorf("errors loading some image configs:\n%s", joinErrors(errors))
	}

	return configs, nil
}

// ListVariants finds all .lua files in the variants directory and loads them.
func ListVariants(variantsDir string) ([]*VariantConfig, error) {
	// Check if directory exists
	if _, err := os.Stat(variantsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("variants directory not found: %s\nHint: Create the directory and add variant .lua files", variantsDir)
		}
		return nil, fmt.Errorf("cannot access variants directory %s: %w", variantsDir, err)
	}

	entries, err := os.ReadDir(variantsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read variants directory %s: %w", variantsDir, err)
	}

	var configs []*VariantConfig
	var errors []string

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".lua" {
			continue
		}

		path := filepath.Join(variantsDir, entry.Name())
		cfg, err := LoadVariantConfig(path)
		if err != nil {
			// Collect all errors instead of failing on first one
			errors = append(errors, fmt.Sprintf("  - %s: %v", entry.Name(), err))
			continue
		}
		configs = append(configs, cfg)
	}

	if len(errors) > 0 {
		return configs, fmt.Errorf("errors loading some variant configs:\n%s", joinErrors(errors))
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

// getDatasetOverlays extracts dataset_overlays from Lua table
func getDatasetOverlays(tbl *lua.LTable, key string) []DatasetOverlay {
	lv := tbl.RawGetString(key)
	if lv.Type() != lua.LTTable {
		return nil
	}

	arr := lv.(*lua.LTable)
	var result []DatasetOverlay

	arr.ForEach(func(_, v lua.LValue) {
		if v.Type() != lua.LTTable {
			return
		}
		t := v.(*lua.LTable)

		// Parse properties sub-table
		props := map[string]string{}
		if p := t.RawGetString("properties"); p.Type() == lua.LTTable {
			p.(*lua.LTable).ForEach(func(k, v lua.LValue) {
				props[k.String()] = v.String()
			})
		}

		result = append(result, DatasetOverlay{
			Name:       getStringField(t, "name"),
			Source:     getStringField(t, "source"),
			Mountpoint: getStringField(t, "mountpoint"),
			CanMount:   getStringField(t, "canmount"),
			Properties: props,
		})
	})

	return result
}

// validateImageConfig validates an image configuration
func validateImageConfig(cfg *ImageConfig, path string) error {
	var missing []string

	if cfg.ID == "" {
		missing = append(missing, "id")
	}
	if cfg.Version == "" {
		missing = append(missing, "version")
	}
	if cfg.ZpoolName == "" {
		missing = append(missing, "zpool_name")
	}
	if cfg.RootDS == "" {
		missing = append(missing, "root_dataset")
	}

	if len(missing) > 0 {
		return fmt.Errorf("image config %s missing required fields: %v\nExample: return { id = \"my-image\", version = \"1.0\", zpool_name = \"mypool\", root_dataset = \"mypool/ROOT/default\", ... }",
			path, missing)
	}

	// Validate field formats
	if len(cfg.ID) > 64 {
		return fmt.Errorf("image config %s: id too long (max 64 characters)", path)
	}

	return nil
}

// validateVariantConfig validates a variant configuration
func validateVariantConfig(cfg *VariantConfig, path string) error {
	var missing []string

	if cfg.ID == "" {
		missing = append(missing, "id")
	}
	if cfg.Name == "" {
		missing = append(missing, "name")
	}

	if len(missing) > 0 {
		return fmt.Errorf("variant config %s missing required fields: %v\nExample: return { id = \"my-variant\", name = \"My Variant\", ... }",
			path, missing)
	}

	// Validate field formats
	if len(cfg.ID) > 64 {
		return fmt.Errorf("variant config %s: id too long (max 64 characters)", path)
	}

	return nil
}

// joinErrors joins error messages with newlines
func joinErrors(errors []string) string {
	result := ""
	for _, err := range errors {
		result += err + "\n"
	}
	return result
}
