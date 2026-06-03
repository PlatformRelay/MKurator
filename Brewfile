# MKurator dev dependencies — install from repo root:
#   brew bundle
# See docs/LOCAL_SETUP.md for tiers and task tools:install (CI-pinned kind/mkcert/task/terraform in bin/).

# Tier A — inner loop
brew "go"
brew "go-task/tap/go-task"

# Tier B — container runtime (Docker Desktop; skip if you use Linux docker.io)
cask "docker"

# Tier C — local kind platform
brew "kind"
brew "kubectl"
brew "helm"
brew "terraform"
brew "mkcert"

# Optional quality-of-life
brew "pre-commit"
brew "direnv"
brew "gitleaks"
