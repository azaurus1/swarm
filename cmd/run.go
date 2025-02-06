/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"sync"
	"time"

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
		var simDuration time.Duration
		var lBound, rBound, tBound, bBound float64

		lBound = 0
		rBound = 550
		bBound = 0
		tBound = 550

		drones := []drone.Drone{
			{Id: "1", X: 1, Y: 50, VX: 0.1, VY: 0, TransmissionRange: 1, DataChan: make(chan []byte, 1024)},
			{Id: "2", X: 2, Y: 50, VX: 0, VY: 0, TransmissionRange: 1, DataChan: make(chan []byte, 1024)},
			{Id: "3", X: 3, Y: 50, VX: 0, VY: 0, TransmissionRange: 1, DataChan: make(chan []byte, 1024)},
			{Id: "4", X: 4, Y: 50, VX: 0, VY: 0, TransmissionRange: 1, DataChan: make(chan []byte, 1024)},
			{Id: "5", X: 5, Y: 50, VX: 0, VY: 0, TransmissionRange: 1, DataChan: make(chan []byte, 1024)},
		}
		r := radio.Radio{}

		droneMap := make(map[string]*drone.Drone)

		for i := range drones {
			droneMap[drones[i].Id] = &drones[i]
		}

		r.Drones = droneMap

		wg.Add(1)
		radioChan = make(chan []byte, 1024)

		for _, d := range drones {
			wg.Add(1)
			dataChannels = append(dataChannels, d.DataChan)

			go d.Start(&wg, radioChan)
		}

		go r.Serve(&wg, radioChan)

		// simulate time passing
		simDuration = 100 * time.Millisecond
		simTicker := time.NewTicker(simDuration)
		done := make(chan bool)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				case <-simTicker.C:
					// loop drones, update locations
					for _, drone := range r.Drones {
						drone.UpdateLocation(simDuration, lBound, rBound, tBound, bBound)
					}
				}
			}
		}()

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
