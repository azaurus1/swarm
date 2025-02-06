/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// renderCmd represents the render command
var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render a simulation",
	Run: func(cmd *cobra.Command, args []string) {
		// drones := []*drone.Drone{
		// 	{Id: "1", X: 100, Y: 100, VX: 0.5, VY: 0.2},
		// 	{Id: "2", X: 200, Y: 200, VX: -0.1, VY: 0.3},
		// 	{Id: "3", X: 300, Y: 20, VX: -0.1, VY: 0.3},
		// 	{Id: "4", X: 100, Y: 200, VX: -0.1, VY: 0.3},
		// }
		// opengl.Run(func() { renderer.Run(drones, 0, 0, 5, 5) })
	},
}

func init() {
	rootCmd.AddCommand(renderCmd)
}
