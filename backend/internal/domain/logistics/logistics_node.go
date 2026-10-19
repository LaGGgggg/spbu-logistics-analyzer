package logistics

import "fmt"

type LogisticsNode struct {
	Kind        string          `json:"kind,omitempty"`
	Cargo       string          `json:"cargo,omitempty"`
	Sources     []string        `json:"sources,omitempty"`
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	RegionID    string          `json:"region_id"`
	Coordinates Coordinates     `json:"coordinates"`
	Modes       []TransportMode `json:"modes"`
	ExternalIDs []ExternalID    `json:"external_ids"`
}

func (n LogisticsNode) Validate() error {

	for _, field := range []struct{ name, value string }{
		{"node ID", n.ID}, {"node name", n.Name}, {"region ID", n.RegionID},
	} {
		if err := validateText(field.name, field.value); err != nil {
			return err
		}
	}

	if err := n.Coordinates.Validate(); err != nil {
		return err
	}

	if len(n.Modes) == 0 {
		return fmt.Errorf("node must support at least one transport mode")
	}

	modes := make(map[TransportMode]bool, len(n.Modes))

	for _, mode := range n.Modes {
		if !mode.Valid() {
			return fmt.Errorf("unknown transport mode %q", mode)
		}

		if modes[mode] {
			return fmt.Errorf("duplicate transport mode %q", mode)
		}

		modes[mode] = true
	}

	ids := make(map[ExternalID]bool, len(n.ExternalIDs))

	for _, id := range n.ExternalIDs {
		if err := id.Validate(); err != nil {
			return err
		}

		if ids[id] {
			return fmt.Errorf("duplicate external ID %s:%s", id.Namespace, id.Value)
		}

		ids[id] = true
	}

	return nil
}
