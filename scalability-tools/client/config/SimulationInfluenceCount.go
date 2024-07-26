package config

// SimulationInfluenceCount represents the customized bound on total number of devices that should be affected
type SimulationInfluenceCount struct {
	DummyWork int `json:"dummyWork"`
	Crash     int `json:"crash"`
}
