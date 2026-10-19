package logistics

type Region struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Bounds Bounds `json:"bounds"`
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
