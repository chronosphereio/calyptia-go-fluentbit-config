package networking

type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
)

func (p Protocol) OK() bool {
	return p == ProtocolTCP || p == ProtocolUDP || p == ProtocolSCTP
}
