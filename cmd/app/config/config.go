package config

type Policy string

const (
	PodsCount Policy = "podscount"
	NodesLoad Policy = "nodesLoad"
)

type Config struct {
	Policy         string
	UpperThreshold string
	LowerThreshold string
}
