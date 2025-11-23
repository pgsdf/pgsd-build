-- Image recipe for the main PGSD desktop system

return {
  id = "pgsd-desktop",
  version = "0.1.0",
  zpool_name = "pgsd",
  root_dataset = "pgsd/ROOT/default",
  pkg_lists = {
    "base",
    "desktop/arcan",
    "desktop/durden",
  },
  overlays = {
    "common",
  },
}
