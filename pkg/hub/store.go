package hub

type Subscription struct {
	Source       string
	ChannelID    string
	UserID       string
	Username     string
	Email        string
	Direct       bool
	OnlyPersonal bool
	CustomFilter string
}

type Store interface {
	FindSubscriptionsBySource(source string) ([]*Subscription, error)
	AddSubscription(sub *Subscription) error
	RemoveSubscription(source string, userID string) (*Subscription, error)
}
