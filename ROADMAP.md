# PGSD Build System Roadmap

This document outlines the development roadmap for the PGSD Build System, tracking completed milestones and planned future enhancements.

## Project Status: **Phase 2 Complete** ✅

The PGSD Build System is now production-ready with a complete ZFS-based installer pipeline and comprehensive build tools.

---

## Phase 1: Core Build Infrastructure ✅ **COMPLETED**

### Milestone 1.1: Project Foundation ✅
- [x] Project structure and organization
- [x] Go module setup (go.mod, go.sum)
- [x] BSD Make-compatible Makefile
- [x] .gitignore configuration
- [x] Basic documentation (README.md, CONTRIBUTING.md)
- [x] License (BSD 3-Clause)

### Milestone 1.2: Lua Configuration System ✅
- [x] Gopher-Lua integration
- [x] Image recipe loader (images/*.lua)
- [x] Variant configuration loader (variants/*.lua)
- [x] Schema validation
- [x] Error reporting and debugging

### Milestone 1.3: Image Build Pipeline ✅
- [x] ZFS sparse file creation
- [x] Memory disk (mdconfig) management
- [x] GPT partitioning
- [x] ZFS pool creation with compression
- [x] Dataset hierarchy (ROOT/default)
- [x] Package installation via pkg
- [x] Overlay application
- [x] ZFS snapshot and stream export
- [x] EFI partition creation
- [x] Cleanup and resource management

### Milestone 1.4: ISO Build Pipeline ✅
- [x] Boot environment assembly
- [x] Bootable ISO creation (makefs + mkisofs)
- [x] Hybrid ISO support (BIOS + UEFI)
- [x] Live environment configuration
- [x] Memory filesystem setup (tmpfs)
- [x] Auto-login support
- [x] Installer integration

---

## Phase 2: Installer Implementation ✅ **COMPLETED**

### Milestone 2.1: CLI Architecture ✅
- [x] Command-line interface with comprehensive flags
- [x] Structured logging system (Debug, Info, Warn, Error)
- [x] Centralized configuration management
- [x] Environment variable support
- [x] Version injection via ldflags
- [x] Builder pattern refactoring
- [x] Utility packages (files, logging)

### Milestone 2.2: Installation Backend ✅
- [x] ZFS-based installation pipeline
- [x] Disk partitioning (GPT: EFI + ZFS)
- [x] EFI filesystem creation (FAT32)
- [x] ZFS pool creation with optimizations
- [x] Root filesystem extraction (xzcat | zfs receive)
- [x] EFI partition installation
- [x] Bootloader installation (boot1.efifat)
- [x] Installation finalization (bootfs, export)
- [x] Comprehensive error handling
- [x] Device path normalization
- [x] Installation state cleanup on failure
- [x] Path traversal protection

### Milestone 2.3: Interactive TUI ✅
- [x] Bubble Tea framework integration
- [x] Welcome screen
- [x] Image selection interface
- [x] Disk discovery and selection
- [x] Confirmation dialog with warnings
- [x] Installation progress display
- [x] Completion and error screens
- [x] Keyboard navigation (vim-style)
- [x] Multi-method disk detection (geom, sysctl)
- [x] Robust parsing with fallbacks
- [x] Async installation with proper message handling
- [x] Bounds checking and defensive programming

### Milestone 2.4: Documentation ✅
- [x] Installer documentation (INSTALLER.md)
- [x] Package lists reference (PACKAGE_LISTS.md)
- [x] Image recipes guide (IMAGE_RECIPES.md)
- [x] Variants documentation (VARIANTS.md)
- [x] Minimal variant guide (MINIMAL_VARIANT.md)
- [x] Build pipeline documentation
- [x] Troubleshooting guides
- [x] Code review and security analysis

---

## Phase 3: Production Recipes and Variants ✅ **IN PROGRESS**

### Milestone 3.1: Desktop Image ✅
- [x] pgsd-desktop image recipe
- [x] Arcan/Durden desktop environment
- [x] Graphics drivers (i915kms, amdgpu)
- [x] Audio support (sndio)
- [x] Network configuration
- [x] Development tools
- [x] Common overlays (boot, system)
- [x] Desktop overlays (profile, zshrc)
- [x] Arcan configuration

### Milestone 3.2: Boot Environments ✅
- [x] pgsd-bootenv-arcan variant
- [x] Live user setup (pgsd/pgsd)
- [x] Auto-login configuration
- [x] RC initialization scripts
- [x] Launcher scripts (pgsd-install, pgsd-welcome)
- [x] Installer integration
- [x] Network utilities
- [x] Disk management tools

### Milestone 3.3: Minimal Installer Variant ✅
- [x] pgsd-bootenv-minimal variant
- [x] Reduced package sets (~600MB)
- [x] Minimal service configuration
- [x] Auto-launch installer
- [x] Optimized boot time (13s to installer)
- [x] Reduced memory footprint (1GB minimum)
- [x] Essential drivers only
- [x] Minimal Durden configuration

### Milestone 3.4: Additional Variants (PLANNED)
- [ ] pgsd-server: Headless server variant
- [ ] pgsd-developer: Development workstation
- [ ] pgsd-gaming: Gaming-focused variant
- [ ] pgsd-embedded: Embedded/IoT variant

---

## Phase 4: Testing and Quality Assurance 🔄 **CURRENT PHASE**

### Milestone 4.1: Unit Testing (IN PROGRESS)
- [x] Installation backend tests
  - [x] Config validation tests
  - [x] Device path normalization tests
  - [ ] Partition creation tests (mock)
  - [ ] ZFS operation tests (mock)
- [ ] TUI tests
  - [ ] State machine tests
  - [ ] Message handling tests
  - [ ] View rendering tests
- [ ] Configuration loader tests
  - [ ] Lua parsing tests
  - [ ] Schema validation tests
  - [ ] Error handling tests
- [ ] Build pipeline tests
  - [ ] Image builder tests
  - [ ] ISO builder tests
  - [ ] Overlay application tests

### Milestone 4.2: Integration Testing (PLANNED)
- [ ] Virtual machine testing
  - [ ] QEMU/KVM test automation
  - [ ] VirtualBox compatibility
  - [ ] VMware compatibility
- [ ] Full installation workflow
  - [ ] Desktop image installation
  - [ ] Server image installation
  - [ ] Boot environment testing
- [ ] Hardware compatibility testing
  - [ ] Intel graphics
  - [ ] AMD graphics
  - [ ] NVIDIA graphics (nouveau)
  - [ ] Various disk controllers
  - [ ] Network interfaces

### Milestone 4.3: Continuous Integration (PLANNED)
- [ ] GitHub Actions workflow
- [ ] Automated builds on commit
- [ ] Automated testing
- [ ] Artifact retention
- [ ] Release automation
- [ ] FreeBSD-specific runners

---

## Phase 5: Advanced Features 📋 **PLANNED**

### Milestone 5.1: Post-Installation Configuration
- [ ] Hostname configuration
- [ ] Root password setup
- [ ] User account creation
- [ ] Network configuration wizard
- [ ] Timezone selection
- [ ] Locale selection
- [ ] SSH key generation
- [ ] Package selection customization

### Milestone 5.2: Real-Time Installation Progress
- [ ] Progress tracking for long operations
- [ ] Percentage completion display
- [ ] Estimated time remaining
- [ ] Detailed operation logging
- [ ] Cancel/abort support
- [ ] Resume failed installations

### Milestone 5.3: Advanced Partitioning
- [ ] Custom partition layouts
- [ ] Manual partitioning mode
- [ ] Multiple disk support
- [ ] RAID-Z configurations
- [ ] ZFS mirror setups
- [ ] Separate /home dataset
- [ ] Swap configuration

### Milestone 5.4: Encryption Support
- [ ] GELI full-disk encryption
- [ ] ZFS native encryption
- [ ] Encrypted boot support
- [ ] Key management
- [ ] Password entry UI
- [ ] Recovery options

### Milestone 5.5: Package Management
- [ ] Custom package set selection
- [ ] Meta-package support
- [ ] Package group selection UI
- [ ] Offline package installation
- [ ] Package cache management
- [ ] Update channel selection

---

## Phase 6: Enhanced User Experience 🎨 **PLANNED**

### Milestone 6.1: Installation UI Improvements
- [ ] Graphical installer (Arcan-native)
- [ ] Better progress visualization
- [ ] Disk usage graphs
- [ ] Network status indicators
- [ ] Hardware detection display
- [ ] Accessibility features
- [ ] Multiple language support

### Milestone 6.2: Documentation Enhancements
- [ ] Video tutorials
- [ ] Interactive guides
- [ ] FAQ expansion
- [ ] Troubleshooting flowcharts
- [ ] Architecture diagrams
- [ ] API documentation
- [ ] Developer guides

### Milestone 6.3: Web Interface (FUTURE)
- [ ] Web-based installer
- [ ] Remote installation support
- [ ] Headless server setup
- [ ] REST API
- [ ] WebSocket progress updates
- [ ] Browser compatibility

---

## Phase 7: Distribution and Deployment 🚀 **PLANNED**

### Milestone 7.1: Official Releases
- [ ] Versioning scheme (semver)
- [ ] Release notes automation
- [ ] Changelog generation
- [ ] Binary distribution
- [ ] Checksum verification
- [ ] GPG signing
- [ ] Mirror infrastructure

### Milestone 7.2: Package Repository
- [ ] FreeBSD ports entry
- [ ] pkg repository setup
- [ ] Package building automation
- [ ] Dependency management
- [ ] Version tracking
- [ ] Security updates

### Milestone 7.3: Community Infrastructure
- [ ] Official website
- [ ] Documentation site
- [ ] Issue tracker
- [ ] Discussion forums
- [ ] Wiki
- [ ] Blog
- [ ] Social media presence

---

## Technical Debt and Maintenance 🔧

### High Priority
- [ ] Add comprehensive unit tests (coverage >80%)
- [ ] Implement integration test suite
- [ ] Add CI/CD pipeline
- [ ] Performance profiling and optimization
- [ ] Memory leak detection and fixes

### Medium Priority
- [ ] Refactor image builder for modularity
- [ ] Improve error messages and hints
- [ ] Add more logging throughout codebase
- [ ] Code coverage reporting
- [ ] Static analysis integration

### Low Priority
- [ ] Code style consistency (golangci-lint)
- [ ] Comment and documentation completeness
- [ ] Example configurations expansion
- [ ] Performance benchmarking suite

---

## Security Roadmap 🔒

### Completed ✅
- [x] Path traversal protection
- [x] Command injection prevention
- [x] Root privilege verification
- [x] Input validation and sanitization
- [x] Resource cleanup on failure

### Planned
- [ ] Secure boot support
- [ ] Signature verification for images
- [ ] Checksum validation for downloads
- [ ] Audit logging
- [ ] Security scanning automation
- [ ] Vulnerability reporting process
- [ ] Security documentation

---

## Performance Goals 🎯

### Current Status
- Image build time: ~10-15 minutes (500MB-2GB images)
- ISO build time: ~2-5 minutes
- Installation time: ~5-10 minutes (depends on disk speed)
- Boot environment boot time: ~13 seconds to installer
- Memory usage: 1-2 GB minimum

### Optimization Targets
- [ ] Reduce image build time by 30% (parallel pkg install)
- [ ] Reduce ISO build time by 50% (caching)
- [ ] Improve installation progress feedback
- [ ] Optimize ZFS stream compression
- [ ] Reduce memory footprint for minimal variant

---

## Community and Contribution 👥

### Documentation for Contributors
- [x] CONTRIBUTING.md guide
- [x] Code of conduct
- [ ] Developer documentation
- [ ] Architecture overview
- [ ] Testing guidelines
- [ ] Release process documentation

### Community Building
- [ ] Set up discussion forums
- [ ] Create contribution guidelines
- [ ] Establish review process
- [ ] Create good first issues
- [ ] Mentorship program
- [ ] Regular community calls

---

## Research and Exploration 🔬

### Investigating
- [ ] Alternative bootloaders (rEFInd, systemd-boot)
- [ ] ZFS send/receive optimizations
- [ ] Incremental image updates
- [ ] Delta updates for packages
- [ ] Container integration (FreeBSD jails)
- [ ] Cloud deployment (AWS, GCP, Azure)
- [ ] ARM64 support

### Experimental Features
- [ ] Live system snapshots
- [ ] Boot environment management UI
- [ ] System rollback from installer
- [ ] Network installation
- [ ] PXE boot support
- [ ] Automated deployment scripts

---

## Version History

### v0.1.0 (Current - In Development)
- ✅ Core build infrastructure
- ✅ Image and ISO build pipelines
- ✅ ZFS-based installer (production-ready)
- ✅ Interactive TUI installer
- ✅ Desktop and boot environment variants
- ✅ Comprehensive documentation
- 🔄 Testing and quality assurance

### v0.2.0 (Planned - Q2 2025)
- Unit and integration tests
- Additional variants (server, developer)
- Post-installation configuration
- Enhanced error handling
- Performance optimizations

### v0.3.0 (Planned - Q3 2025)
- Advanced partitioning support
- Encryption support
- Package management UI
- Real-time installation progress
- CI/CD pipeline

### v1.0.0 (Planned - Q4 2025)
- Production release
- Complete test coverage
- Full documentation
- Community infrastructure
- Official FreeBSD ports entry

---

## Success Metrics 📊

### Technical Metrics
- Build success rate: >99%
- Installation success rate: >95%
- Code coverage: >80%
- Documentation coverage: 100%
- Average build time: <10 minutes
- Average installation time: <10 minutes

### Community Metrics
- Active contributors: TBD
- GitHub stars: TBD
- Issues resolved: TBD
- Pull requests merged: TBD
- Documentation views: TBD

---

## Dependencies and Requirements

### Build Dependencies
- FreeBSD 14.0 or later
- Go 1.21 or later
- BSD Make
- ZFS kernel module
- Arcan/Durden (for boot environments)

### Runtime Dependencies
- FreeBSD base system
- pkg package manager
- ZFS utilities
- UEFI boot support
- 1-2 GB RAM minimum
- 8 GB disk minimum

---

## Contributing to the Roadmap

The roadmap is a living document. Community feedback and contributions are welcome:

1. **Propose New Features**: Open an issue with the "enhancement" label
2. **Vote on Priorities**: React to issues with 👍 for features you want
3. **Submit Pull Requests**: Implement roadmap items and submit PRs
4. **Provide Feedback**: Share your experience and suggestions

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## Contact and Resources

- **Repository**: https://github.com/pgsdf/pgsd-build
- **Documentation**: [docs/](docs/)
- **Issues**: https://github.com/pgsdf/pgsd-build/issues
- **Discussions**: TBD

---

**Last Updated**: 2025-11-23
**Status**: Phase 2 Complete, Phase 4 In Progress
**Next Milestone**: Unit Testing (4.1)
