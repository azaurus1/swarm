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
		var dataChannels []chan []byte
		var radioChan chan []byte

		drones := []drone.Drone{
			{Id: "1", X: 1, Y: 1, VX: 0.5, VY: 0.2, TransmissionRange: 1, DataChan: make(chan []byte, 1024)},
			{Id: "2", X: 3, Y: 1, VX: -0.1, VY: 0.3, TransmissionRange: 3, DataChan: make(chan []byte, 1024)},
			{Id: "3", X: 8, Y: 1, VX: -0.1, VY: 0.3, TransmissionRange: 6, DataChan: make(chan []byte, 1024)},
			{Id: "4", X: 4, Y: 5, VX: -0.1, VY: 0.3, TransmissionRange: 4, DataChan: make(chan []byte, 1024)},
			{Id: "5", X: 8, Y: 8, VX: -0.1, VY: 0.3, TransmissionRange: 4, DataChan: make(chan []byte, 1024)},
			{Id: "6", X: 1, Y: 6, VX: -0.1, VY: 0.3, TransmissionRange: 4, DataChan: make(chan []byte, 1024)},
			{Id: "7", X: 1, Y: 8, VX: -0.1, VY: 0.3, TransmissionRange: 4, DataChan: make(chan []byte, 1024)},
		}
		r := radio.Radio{}

		droneMap := make(map[string]drone.Drone)

		for _, drone := range drones {
			droneMap[drone.Id] = drone
		}

		r.Drones = droneMap

		wg.Add(1)
		radioChan = make(chan []byte, 1024)
		go r.Serve(&wg, radioChan)

		for _, d := range drones {
			wg.Add(1)
			dataChannels = append(dataChannels, d.DataChan)

			go d.Start(&wg, radioChan)
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
