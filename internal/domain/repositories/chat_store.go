package repositories

type ChatStorage interface {
	MessageStorage
	SessionStorage
}
