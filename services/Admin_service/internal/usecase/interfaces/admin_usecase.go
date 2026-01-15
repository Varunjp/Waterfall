package interfaces

type AdminUsecase interface {
	Login(email, password string) (string, error)
}