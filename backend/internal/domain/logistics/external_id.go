package logistics

import "fmt"

type ExternalNamespace string

const (
	ExternalOSMNode     ExternalNamespace = "osm_node"
	ExternalOSMWay      ExternalNamespace = "osm_way"
	ExternalOSMRelation ExternalNamespace = "osm_relation"
	ExternalESR         ExternalNamespace = "esr"
	ExternalUNLOCODE    ExternalNamespace = "unlocode"
	ExternalIATA        ExternalNamespace = "iata"
	ExternalICAO        ExternalNamespace = "icao"
)

type ExternalID struct {
	Namespace ExternalNamespace `json:"namespace"`
	Value     string            `json:"value"`
}

func (id ExternalID) Validate() error {
	switch id.Namespace {
	case ExternalOSMNode, ExternalOSMWay, ExternalOSMRelation, ExternalESR, ExternalUNLOCODE, ExternalIATA, ExternalICAO:
	default:
		return fmt.Errorf("unknown external namespace %q", id.Namespace)
	}
	return validateText("external ID value", id.Value)
}
