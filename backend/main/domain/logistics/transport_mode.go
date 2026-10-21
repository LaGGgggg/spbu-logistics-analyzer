package logistics

type TransportMode string

const (
	TransportRail  TransportMode = "rail"
	TransportSea   TransportMode = "sea"
	TransportRiver TransportMode = "river"
	TransportAir   TransportMode = "air"
)

func (m TransportMode) Valid() bool {
	switch m {
	case TransportRail, TransportSea, TransportRiver, TransportAir:
		return true
	default:
		return false
	}
}
