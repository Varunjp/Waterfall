package repository

type AdminQuotaRepository interface {
	CanStart(appID string) (bool, error)
	Increment(appID string) error
	Decrement(appID string) error
}