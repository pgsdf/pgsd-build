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
	ID           string
	Path         string
	ManifestPath string
}

// DiskInfo represents a disk available for installation
type DiskInfo struct {
	Device string
	Size   string
	Model  string
}

// Installation messages
type installLogMsg string
type installCompleteMsg struct{}
type installErrorMsg struct{ err error }

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
	case installLogMsg:
		m.installLog = append(m.installLog, string(msg))
		return m, nil
	case installCompleteMsg:
		m.state = stateComplete
		return m, nil
	case installErrorMsg:
		m.err = msg.err
		m.state = stateError
		return m, nil
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
			return m, m.performInstallation()
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
				ID:           entry.Name(),
				Path:         imgPath,
				ManifestPath: manifestPath,
			})
		}
	}

	return images, nil
}

// loadDisks detects available disks for installation
func loadDisks() ([]DiskInfo, error) {
	// Try geom disk list first (more reliable on FreeBSD)
	if disks := loadDisksFromGeom(); len(disks) > 0 {
		return disks, nil
	}

	// Try sysctl kern.disks as fallback
	if disks := loadDisksFromSysctl(); len(disks) > 0 {
		return disks, nil
	}

	// Fall back to dummy disks for development/testing
	return []DiskInfo{
		{Device: "ada0", Size: "500GB", Model: "Virtual Disk"},
		{Device: "ada1", Size: "1TB", Model: "Virtual Disk"},
	}, nil
}

// loadDisksFromGeom uses geom disk list to discover disks
func loadDisksFromGeom() []DiskInfo {
	// Run: geom disk list
	cmd := exec.Command("geom", "disk", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var disks []DiskInfo
	var currentDisk DiskInfo
	inDisk := false

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// New disk section starts with "Geom name:"
		if strings.HasPrefix(line, "Geom name:") {
			if inDisk && currentDisk.Device != "" {
				disks = append(disks, currentDisk)
			}
			currentDisk = DiskInfo{}
			inDisk = true
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				currentDisk.Device = fields[2]
			}
		} else if inDisk {
			// Parse Mediasize field
			if strings.HasPrefix(line, "Mediasize:") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					// Format: "Mediasize: 21474836480 (20G)"
					if len(fields) >= 4 && strings.HasPrefix(fields[3], "(") {
						currentDisk.Size = strings.Trim(fields[3], "()")
					} else {
						currentDisk.Size = formatBytes(fields[1])
					}
				}
			}
			// Parse descr field (model)
			if strings.HasPrefix(line, "descr:") {
				currentDisk.Model = strings.TrimPrefix(line, "descr:")
				currentDisk.Model = strings.TrimSpace(currentDisk.Model)
			}
		}
	}

	// Add last disk
	if inDisk && currentDisk.Device != "" {
		disks = append(disks, currentDisk)
	}

	// Filter out CD-ROM and other non-disk devices
	var filtered []DiskInfo
	for _, disk := range disks {
		// Skip CD-ROM, memory disks, etc.
		if !strings.HasPrefix(disk.Device, "cd") &&
			!strings.HasPrefix(disk.Device, "md") &&
			!strings.HasPrefix(disk.Device, "pass") {
			filtered = append(filtered, disk)
		}
	}

	return filtered
}

// loadDisksFromSysctl uses sysctl kern.disks to discover disks
func loadDisksFromSysctl() []DiskInfo {
	// Run: sysctl -n kern.disks
	cmd := exec.Command("sysctl", "-n", "kern.disks")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	// Parse space-separated list of disk names
	diskNames := strings.Fields(string(output))
	if len(diskNames) == 0 {
		return nil
	}

	var disks []DiskInfo
	for _, name := range diskNames {
		// Skip CD-ROM and memory disks
		if strings.HasPrefix(name, "cd") ||
			strings.HasPrefix(name, "md") ||
			strings.HasPrefix(name, "pass") {
			continue
		}

		// Get disk size using diskinfo
		size := "Unknown"
		model := "Disk"

		cmd := exec.Command("diskinfo", name)
		if output, err := cmd.Output(); err == nil {
			fields := strings.Fields(string(output))
			if len(fields) >= 3 {
				// diskinfo output: device sectors size model
				size = formatBytes(fields[2])
			}
			if len(fields) >= 4 {
				model = strings.Join(fields[3:], " ")
			}
		}

		disks = append(disks, DiskInfo{
			Device: name,
			Size:   size,
			Model:  model,
		})
	}

	return disks
}

// formatBytes converts byte count to human-readable size
func formatBytes(bytesStr string) string {
	// Try to parse as integer
	var bytes int64
	if _, err := fmt.Sscanf(bytesStr, "%d", &bytes); err != nil {
		return bytesStr
	}

	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fTB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// performInstallation executes the installation pipeline and returns a command
func (m model) performInstallation() tea.Cmd {
	image := m.images[m.selectedImg]
	disk := m.disks[m.selectedDisk]

	return func() tea.Msg {
		// Perform the installation synchronously
		// The installation log messages will be collected
		var logs []string
		var installErr error

		cfg := install.Config{
			ImagePath:  image.Path,
			TargetDisk: disk.Device,
			ZpoolName:  "pgsd", // Default pool name
			LogFunc: func(msg string) {
				logs = append(logs, msg)
			},
		}

		// Add initial log messages
		logs = append(logs, "Starting installation...")
		logs = append(logs, fmt.Sprintf("Image: %s", image.ID))
		logs = append(logs, fmt.Sprintf("Target disk: %s", disk.Device))

		// Run installation
		installErr = install.Install(cfg)

		// Build a batch of log messages to send
		var cmds []tea.Cmd
		for _, log := range logs {
			msg := log // Capture for closure
			cmds = append(cmds, func() tea.Msg {
				return installLogMsg(msg)
			})
		}

		// Add final status message
		if installErr != nil {
			cmds = append(cmds, func() tea.Msg {
				return installErrorMsg{err: installErr}
			})
		} else {
			logs = append(logs, "Installation complete!")
			cmds = append(cmds, func() tea.Msg {
				return installLogMsg("Installation complete!")
			})
			cmds = append(cmds, func() tea.Msg {
				return installCompleteMsg{}
			})
		}

		// Execute first command and return it
		// Note: This is a simplified version - in production we'd use a better streaming approach
		if len(cmds) > 0 {
			return tea.Batch(cmds...)()
		}

		return installCompleteMsg{}
	}
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error running TUI:", err)
		os.Exit(1)
	}
}
