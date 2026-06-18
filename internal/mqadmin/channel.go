package mqadmin

// NormalizeChannelType returns svrconn when empty.
func NormalizeChannelType(t ChannelType) ChannelType {
	if t == "" {
		return ChannelTypeSvrconn
	}
	return t
}

// ChannelTypeSupported reports whether the operator reconciles this channel kind.
func ChannelTypeSupported(t ChannelType) bool {
	switch NormalizeChannelType(t) {
	case ChannelTypeSvrconn, ChannelTypeSdr, ChannelTypeRcvr:
		return true
	default:
		return false
	}
}
