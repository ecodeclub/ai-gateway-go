package domain

type Provider1 struct {
	Id     int64
	Name   string
	ApiKey string
	Models []Model1
}

type Model1 struct {
	Id          int64
	Pid         int64
	Name        string
	InputPrice  int64
	OutPutPrice int64
	PriceMode   string
}
