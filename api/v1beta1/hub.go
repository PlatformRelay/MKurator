package v1beta1

// Hub marks Queue as the conversion hub (v1beta1).
func (*Queue) Hub() {}

// Hub marks Topic as the conversion hub (v1beta1).
func (*Topic) Hub() {}

// Hub marks Channel as the conversion hub (v1beta1).
func (*Channel) Hub() {}

// Hub marks ChannelAuthRule as the conversion hub (v1beta1).
func (*ChannelAuthRule) Hub() {}

// Hub marks AuthorityRecord as the conversion hub (v1beta1).
func (*AuthorityRecord) Hub() {}

// Hub marks QueueManagerConnection as the conversion hub (v1beta1).
func (*QueueManagerConnection) Hub() {}
