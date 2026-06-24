package services

type DatabaseService interface {
	GetAll() error
	Get(id int) error
	Create() error
	Update() error
	Delete(id int) error
}

type DatabaseConfiguration struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}
