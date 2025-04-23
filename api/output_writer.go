package api

import (
	"github.com/spf13/cobra"
)

type CobraOutputWriter struct {
	Cmd *cobra.Command
}

func (c *CobraOutputWriter) Print(args ...interface{}) {
	c.Cmd.Print(args...)
}
func (c *CobraOutputWriter) Printf(format string, args ...interface{}) {
	c.Cmd.Printf(format, args...)
}
func (c *CobraOutputWriter) Println(args ...interface{}) {
	c.Cmd.Println(args...)
}
func (c *CobraOutputWriter) Errorf(format string, args ...interface{}) {
	c.Cmd.Printf(format, args...) // Cobra doesn't have Errorf, so use Printf
}
