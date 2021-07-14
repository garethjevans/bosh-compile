package manifest

import "github.com/stevenle/topsort"

type Package struct {
	Name         string   `yaml:"name"`
	Dependencies []string `yaml:"dependencies"`
}

type Manifest struct {
	Name     string    `yaml:"name"`
	Packages []Package `yaml:"packages"`
}

func (m *Manifest) Graph() (*topsort.Graph, error) {
	// Initialize the graph.
	graph := topsort.NewGraph()

	for _, p := range m.Packages {
		for _, d := range p.Dependencies {
			err := graph.AddEdge(p.Name, d)
			if err != nil {
				return nil, err
			}
		}
	}

	return graph, nil
}

func (m *Manifest) Dependencies(name string) ([]string, error) {
	graph, err := m.Graph()
	if err != nil {
		return nil, err
	}

	return graph.TopSort(name)
}

func (m *Manifest) TopLevelPackages() ([]string, error) {
	graph, err := m.Graph()
	if err != nil {
		return nil, err
	}

	topLevel := []string{}
	for _, p := range m.Packages {
		r, err := graph.TopSort(p.Name)
		if err != nil {
			return nil, err
		}
		if len(r) > 1 {
			topLevel = append(topLevel, p.Name)
		}
	}
	return topLevel, nil
}
