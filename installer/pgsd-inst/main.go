package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/pgsdf/pgsdbuild/installer/internal/install"
)

// Installation states
const (
    stateWelcome = iota
    stateImageSelect
    stateDiskSelect
    stateConfirm
    stateInstalling
    stateComplete
    stateError
)

// ImageInfo represents a system image available for installation
type ImageInfo struct {
    ID          string
    Path        string
    ManifestPath string
}

// DiskInfo represents a disk available for installation
type DiskInfo struct {
    Device string
    Size   string
    Model  string
}

type model struct {
    state        int
    images       []ImageInfo
    disks        []DiskInfo
    selectedImg  int
    selectedDisk int
    cursor       int
    err          error
    installLog   []string
}

func initialModel() model {
    return model{
        state: stateWelcome,
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKey(msg)
    }
    return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "ctrl+c", "q":
        if m.state == stateInstalling {
            return m, nil // Don't allow quit during installation
        }
        return m, tea.Quit
    }

    switch m.state {
    case stateWelcome:
        if msg.String() == "enter" {
            // Load available images
            images, err := loadImages()
            if err != nil {
                m.err = err
                m.state = stateError
                return m, nil
            }
            m.images = images
            if len(images) == 0 {
                m.err = fmt.Errorf("no system images found")
                m.state = stateError
                return m, nil
            }
            m.state = stateImageSelect
        }

    case stateImageSelect:
        switch msg.String() {
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.images)-1 {
                m.cursor++
            }
        case "enter":
            m.selectedImg = m.cursor
            m.cursor = 0
            // Load available disks
            disks, err := loadDisks()
            if err != nil {
                m.err = err
                m.state = stateError
                return m, nil
            }
            m.disks = disks
            if len(disks) == 0 {
                m.err = fmt.Errorf("no disks found")
                m.state = stateError
                return m, nil
            }
            m.state = stateDiskSelect
        }

    case stateDiskSelect:
        switch msg.String() {
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.disks)-1 {
                m.cursor++
            }
        case "enter":
            m.selectedDisk = m.cursor
            m.state = stateConfirm
        }

    case stateConfirm:
        switch msg.String() {
        case "y", "Y":
            m.state = stateInstalling
            go m.performInstallation()
        case "n", "N":
            m.state = stateDiskSelect
            m.cursor = m.selectedDisk
        }

    case stateComplete, stateError:
        if msg.String() == "enter" {
            return m, tea.Quit
        }
    }

    return m, nil
}

func (m model) View() string {
    switch m.state {
    case stateWelcome:
        return m.viewWelcome()
    case stateImageSelect:
        return m.viewImageSelect()
    case stateDiskSelect:
        return m.viewDiskSelect()
    case stateConfirm:
        return m.viewConfirm()
    case stateInstalling:
        return m.viewInstalling()
    case stateComplete:
        return m.viewComplete()
    case stateError:
        return m.viewError()
    }
    return ""
}

func (m model) viewWelcome() string {
    var b strings.Builder
    b.WriteString("╔════════════════════════════════════════╗\n")
    b.WriteString("║   PGSD System Installer (Prototype)    ║\n")
    b.WriteString("╚════════════════════════════════════════╝\n\n")
    b.WriteString("This installer will guide you through\n")
    b.WriteString("installing PGSD to your system.\n\n")
    b.WriteString("Press Enter to continue\n")
    b.WriteString("Press q to quit\n")
    return b.String()
}

func (m model) viewImageSelect() string {
    var b strings.Builder
    b.WriteString("╔════════════════════════════════════════╗\n")
    b.WriteString("║        Select System Image             ║\n")
    b.WriteString("╚════════════════════════════════════════╝\n\n")

    for i, img := range m.images {
        cursor := " "
        if i == m.cursor {
            cursor = ">"
        }
        b.WriteString(fmt.Sprintf(" %s %s\n", cursor, img.ID))
    }

    b.WriteString("\n")
    b.WriteString("↑/↓ or k/j: Navigate\n")
    b.WriteString("Enter: Select\n")
    b.WriteString("q: Quit\n")
    return b.String()
}

func (m model) viewDiskSelect() string {
    var b strings.Builder
    b.WriteString("╔════════════════════════════════════════╗\n")
    b.WriteString("║        Select Target Disk              ║\n")
    b.WriteString("╚════════════════════════════════════════╝\n\n")
    b.WriteString("WARNING: All data on the selected disk\n")
    b.WriteString("will be DESTROYED!\n\n")

    for i, disk := range m.disks {
        cursor := " "
        if i == m.cursor {
            cursor = ">"
        }
        b.WriteString(fmt.Sprintf(" %s %s (%s) - %s\n",
            cursor, disk.Device, disk.Size, disk.Model))
    }

    b.WriteString("\n")
    b.WriteString("↑/↓ or k/j: Navigate\n")
    b.WriteString("Enter: Select\n")
    b.WriteString("q: Quit\n")
    return b.String()
}

func (m model) viewConfirm() string {
    var b strings.Builder
    b.WriteString("╔════════════════════════════════════════╗\n")
    b.WriteString("║           Confirmation                 ║\n")
    b.WriteString("╚════════════════════════════════════════╝\n\n")

    b.WriteString("You are about to install:\n\n")
    b.WriteString(fmt.Sprintf("  Image: %s\n", m.images[m.selectedImg].ID))
    b.WriteString(fmt.Sprintf("  Disk:  %s (%s)\n\n",
        m.disks[m.selectedDisk].Device,
        m.disks[m.selectedDisk].Size))
    b.WriteString("WARNING: This will DESTROY all data on\n")
    b.WriteString("the target disk!\n\n")
    b.WriteString("Continue? (y/n)\n")
    return b.String()
}

func (m model) viewInstalling() string {
    var b strings.Builder
    b.WriteString("╔════════════════════════════════════════╗\n")
    b.WriteString("║         Installing...                  ║\n")
    b.WriteString("╚════════════════════════════════════════╝\n\n")

    if len(m.installLog) > 0 {
        // Show last 10 log lines
        start := 0
        if len(m.installLog) > 10 {
            start = len(m.installLog) - 10
        }
        for i := start; i < len(m.installLog); i++ {
            b.WriteString(m.installLog[i])
            b.WriteString("\n")
        }
    } else {
        b.WriteString("Please wait...\n")
    }

    return b.String()
}

func (m model) viewComplete() string {
    var b strings.Builder
    b.WriteString("╔════════════════════════════════════════╗\n")
    b.WriteString("║      Installation Complete!            ║\n")
    b.WriteString("╚════════════════════════════════════════╝\n\n")
    b.WriteString("PGSD has been successfully installed to\n")
    b.WriteString(fmt.Sprintf("%s\n\n", m.disks[m.selectedDisk].Device))
    b.WriteString("You may now reboot your system.\n\n")
    b.WriteString("Press Enter to exit\n")
    return b.String()
}

func (m model) viewError() string {
    var b strings.Builder
    b.WriteString("╔════════════════════════════════════════╗\n")
    b.WriteString("║             Error                      ║\n")
    b.WriteString("╚════════════════════════════════════════╝\n\n")
    b.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
    b.WriteString("Press Enter to exit\n")
    return b.String()
}

// loadImages scans for available system images
func loadImages() ([]ImageInfo, error) {
    imagesDir := "/usr/local/share/pgsd/images"

    // For prototype, also check local artifacts directory
    if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
        imagesDir = "artifacts"
    }

    entries, err := os.ReadDir(imagesDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read images directory: %w", err)
    }

    var images []ImageInfo
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        imgPath := filepath.Join(imagesDir, entry.Name())
        manifestPath := filepath.Join(imgPath, "manifest.toml")

        // Check if manifest exists
        if _, err := os.Stat(manifestPath); err == nil {
            images = append(images, ImageInfo{
                ID:          entry.Name(),
                Path:        imgPath,
                ManifestPath: manifestPath,
            })
        }
    }

    return images, nil
}

// loadDisks detects available disks for installation
func loadDisks() ([]DiskInfo, error) {
    // On FreeBSD, we would use: diskinfo -l to list disks
    // For prototype, we'll create dummy disks

    // Try to run diskinfo if available
    cmd := exec.Command("diskinfo", "-l")
    if output, err := cmd.Output(); err == nil {
        // Parse real diskinfo output
        return parseDiskinfo(string(output)), nil
    }

    // Fall back to dummy disks for prototype
    return []DiskInfo{
        {Device: "ada0", Size: "500GB", Model: "Virtual Disk"},
        {Device: "ada1", Size: "1TB", Model: "Virtual Disk"},
    }, nil
}

// parseDiskinfo parses diskinfo output
func parseDiskinfo(output string) []DiskInfo {
    var disks []DiskInfo
    lines := strings.Split(output, "\n")

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        fields := strings.Fields(line)
        if len(fields) >= 2 {
            disks = append(disks, DiskInfo{
                Device: fields[0],
                Size:   fields[1],
                Model:  strings.Join(fields[2:], " "),
            })
        }
    }

    return disks
}

// performInstallation executes the installation pipeline
func (m *model) performInstallation() {
    image := m.images[m.selectedImg]
    disk := m.disks[m.selectedDisk]

    m.addLog("Starting installation...")
    m.addLog(fmt.Sprintf("Image: %s", image.ID))
    m.addLog(fmt.Sprintf("Target disk: %s", disk.Device))

    // Perform the actual installation
    cfg := install.Config{
        ImagePath:  image.Path,
        TargetDisk: disk.Device,
        ZpoolName:  "pgsd", // Default pool name
        LogFunc:    m.addLog,
    }

    if err := install.Install(cfg); err != nil {
        m.addLog(fmt.Sprintf("Installation failed: %v", err))
        m.err = err
        m.state = stateError
        return
    }

    m.state = stateComplete
}

// addLog adds a message to the installation log
func (m *model) addLog(msg string) {
    m.installLog = append(m.installLog, msg)
}

func main() {
    if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
        fmt.Fprintln(os.Stderr, "error running TUI:", err)
        os.Exit(1)
    }
}
