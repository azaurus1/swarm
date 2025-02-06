package renderer

import (
	"fmt"
	"sync"

	"github.com/azaurus1/swarm/internal/drone"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func Run(wg *sync.WaitGroup, drones map[string]*drone.Drone, lBound, rBound, bBound, tBound float64) {
	defer wg.Done()

	cfg := opengl.WindowConfig{
		Title:  "Swarm Simulation",
		Bounds: pixel.R(lBound, bBound, rBound, tBound),
	}
	win, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	defer win.Destroy()

	var imDrones []*imdraw.IMDraw

	for _, drone := range drones {
		imDrone := imdraw.New(nil)
		imDrone.Color = colornames.Green
		imDrone.Push(pixel.V(drone.X, drone.Y))
		imDrone.Circle(15, 0)

		imDrones = append(imDrones, imDrone)
	}

	for !win.Closed() {
		win.Clear(colornames.Black)

		for i := range drones {
			drones[i].X += drones[i].VX
			drones[i].Y += drones[i].VY

			if drones[i].X < 0 || drones[i].X > win.Bounds().W() {
				drones[i].VX = -drones[i].VX
			}
			if drones[i].Y < 0 || drones[i].Y > win.Bounds().H() {
				drones[i].VY = -drones[i].VY
			}

			imDrone := imdraw.New(nil)
			imDrone.Color = colornames.Green
			imDrone.Push(pixel.V(drones[i].X, drones[i].Y))
			imDrone.Circle(5, 0)

			atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
			imDroneTxt := text.New(pixel.V(drones[i].X+20, drones[i].Y), atlas)
			fmt.Fprintln(imDroneTxt, "Id:", drones[i].Id)
			fmt.Fprintf(imDroneTxt, "X: %.2f\n", drones[i].X)
			fmt.Fprintf(imDroneTxt, "Y: %.2f\n", drones[i].Y)
			fmt.Fprintf(imDroneTxt, "VX: %.2f\n", drones[i].VX)
			fmt.Fprintf(imDroneTxt, "VY: %.2f\n", drones[i].VY)

			imDrone.Draw(win)
			imDroneTxt.Draw(win, pixel.IM)

		}

		// Update the window
		win.Update()
	}
}
