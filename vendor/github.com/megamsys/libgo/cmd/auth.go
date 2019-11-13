package cmd

type loginScheme struct {
	Name string
	Data map[string]string
}

type login struct {
	scheme *loginScheme
}

func (c *login) Run(context *Context) error {
	return nil
}

func (c *login) Info() *Info {
	usage := "login [email]"
	return &Info{
		Name:    "login",
		Usage:   usage,
		Desc:    `Initiates a new megam session for a user.`,
		MinArgs: 0,
	}
}
