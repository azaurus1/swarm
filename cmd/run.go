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
		var dataChannels []chan string
		var radioChan chan string

		r := radio.Radio{}
		drones := []drone.Drone{
			{Id: 1, X: 100, Y: 100, VX: 0.5, VY: 0.2, TransmissionRange: 10, DataChan: make(chan string, 10)},
			{Id: 2, X: 200, Y: 200, VX: -0.1, VY: 0.3, TransmissionRange: 5, DataChan: make(chan string, 10)},
			{Id: 3, X: 300, Y: 20, VX: -0.1, VY: 0.3, TransmissionRange: 3, DataChan: make(chan string, 10)},
			{Id: 4, X: 100, Y: 200, VX: -0.1, VY: 0.3, TransmissionRange: 10, DataChan: make(chan string, 10)},
		}
		wg.Add(1)
		radioChan = make(chan string)
		go r.Serve(drones, &wg, radioChan)

		for _, d := range drones {
			wg.Add(1)
			dataChannels = append(dataChannels, d.DataChan)

			go d.Start(r.Addr, &wg, radioChan)
		}

		// for _, c := range dataChannels {
		// 	c <- "testing"
		// }

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
