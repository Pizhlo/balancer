package balancer

type ConfigDB struct {
	ID       int
	Address  string
	RRWeight int
	IsActive bool
}
