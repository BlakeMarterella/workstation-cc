# Profiles.md

The configuration profiles are a way to quickly set up a machine with a predefined set of packages, dotfiles, and scripts. Each profile corresponds to a specific use case or profile (e.g. `slim`, `full`, `dev`, etc.) and can be selected during the installation process.

## Profile structure

Each profile is defined in a YAML file under the `profiles/` directory. The structure of a profile file looks like this:

```yaml
name: slim
description: A minimal setup with essential tools and dotfiles.
packages:
  - name: git
    manager: brew
  - name: zsh
    manager: brew
dotfiles:
  - .zshrc
  - .vimrc
scripts:    
  - name: cleanup.sh
    description: A script to clean up temporary files.
```
