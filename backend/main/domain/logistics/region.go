package logistics

type Region struct {
	SourceVersion string `json:"source_version,omitempty"`
	SourceID      string `json:"source_id,omitempty"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Bounds        Bounds `json:"bounds"`

	// Optimization overrides the catalog freight policy for this region;
	// nil keeps the configured default.
	Optimization *int `json:"optimization,omitempty"`
}

func (r Region) Validate() error {

	if err := validateText("region ID", r.ID); err != nil {
		return err
	}

	if err := validateText("region name", r.Name); err != nil {
		return err
	}

	return r.Bounds.Validate()
}
