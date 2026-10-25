package logistics

type TransportMode string

const (
	TransportRail  TransportMode = "rail"
	TransportSea   TransportMode = "sea"
	TransportRiver TransportMode = "river"
)

func (m TransportMode) Valid() bool {
	switch m {
	case TransportRail, TransportSea, TransportRiver:
		return true
	default:
		return false
	}
}
