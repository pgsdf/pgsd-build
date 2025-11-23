-- Boot environment variant for PGSD using Arcan + Durden as the desktop
-- and pgsd-inst as the installer.

return {
  id = "pgsd-bootenv-arcan",
  name = "PGSD Boot Environment (Arcan)",

  pkg_lists = {
    "base",
    "arcan",
    "installer/pgsd-inst",
  },

  overlays = {
    "common",
    "arcan",
    "bootenv",
  },

  images_dir = "/usr/local/share/pgsd/images",
}
