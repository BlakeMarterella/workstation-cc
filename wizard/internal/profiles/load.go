package profiles

import (
	"fmt"

	assets "github.com/BlakeMarterella/workstation-cc"
)

// Load reads the profile definitions embedded in the binary.
func Load() (*Set, error) {
	data, err := assets.ProfilesFS.ReadFile("profiles.yaml")
	if err != nil {
		return nil, fmt.Errorf("read embedded profiles: %w", err)
	}
	return parseSet(data)
}
