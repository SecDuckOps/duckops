package attack_path_renderer

type Frame struct {
	Step      int      `json:"step"`
	Node      string   `json:"node"`
	Visited   []string `json:"visited"`
	Completed bool     `json:"completed"`
}

func BuildReplay(path []string) []Frame {
	frames := make([]Frame, 0, len(path))
	visited := []string{}
	for i, n := range path {
		visited = append(visited, n)
		frames = append(frames, Frame{
			Step:      i + 1,
			Node:      n,
			Visited:   append([]string{}, visited...),
			Completed: i == len(path)-1,
		})
	}
	return frames
}
