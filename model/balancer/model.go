package balancer

type Target struct {
	Address  string
	RRWeight int
	IsActive bool
}
