package git

// Push sends commits and tags to the remote repository.
func (g *Git) Push() error {
	args := []string{"push"}
	args = append(args, g.config.PushArgs...)

	if g.config.PushRepo != "" {
		args = append(args, g.config.PushRepo)
	}

	_, err := g.run(args...)
	return err
}
