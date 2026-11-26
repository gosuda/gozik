# CI
PRs targeting dev, main or master and merge queues run one Depot Linux amd64
build. Weekly/manual runs retain Linux arm64/riscv64 cross-builds and native
Windows x64/macOS ARM64 builds. All Apple execution uses depot-macos-latest.
Tagged installer releases retain their complete supported payload build.
The duplicate rolling/tag payload pipeline is manual-only; plugin packaging
runs for plugin tags or manual dispatch, not workflow-file branch pushes.

## Windows ARM64
Depot provides Windows x64 runners, not Windows ARM64. The GTK/MSYS2 application
and PyInstaller YouTube Music plugin depended on native ARM64 execution; their
ARM64 rows are removed rather than mislabeled as x64-equivalent.
Windows ARM64 installers/uninstallers are no longer produced because they
would request an application payload this CI cannot supply. Existing published
assets are untouched. The portable Go Spotify plugin still cross-builds its
Windows ARM64 asset; that is a compile guarantee, not a native runtime test.
