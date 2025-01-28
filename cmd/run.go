/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"sync"

	"github.com/azaurus1/swarm/internal/drone"
	"github.com/azaurus1/swarm/internal/radio"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the simulation",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup

		r := radio.Radio{
			Addr: "localhost:50000",
		}
		drones := []drone.Drone{
			{Id: 1, X: 100, Y: 100, VX: 0.5, VY: 0.2, TransmissionRange: 10, Addr: "localhost:50001"},
			{Id: 2, X: 200, Y: 200, VX: -0.1, VY: 0.3, TransmissionRange: 5, Addr: "localhost:50002"},
			{Id: 3, X: 300, Y: 20, VX: -0.1, VY: 0.3, TransmissionRange: 3, Addr: "localhost:50003"},
			{Id: 4, X: 100, Y: 200, VX: -0.1, VY: 0.3, TransmissionRange: 10, Addr: "localhost:50004"},
		}

		for _, d := range drones {
			wg.Add(1)
			go d.Start(r.Addr, &wg)
		}
		wg.Add(1)
		go r.Serve(drones, &wg)

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
