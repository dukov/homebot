package prometheus

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"log"
	"math/rand"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Client struct {
	v1api v1.API
}

func NewClient(address string) (*Client, error) {
	promClient, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	return &Client{v1api: v1.NewAPI(promClient)}, nil
}

func (c *Client) QueryRange(q string, rng v1.Range) (model.Matrix, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := c.v1api.QueryRange(ctx, q, rng, v1.WithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Printf("Warnings: %v\n", warnings)
	}
	if result.Type() != model.ValMatrix {
		return nil, fmt.Errorf("Return value is not Vector")
	}
	return result.(model.Matrix), nil
}

func (c *Client) Render(data model.Matrix, rng v1.Range) (io.WriterTo, error) {
	start := model.TimeFromUnixNano(rng.Start.UnixNano())
	p := plot.New()
	for _, graph := range data {
		dots := make(plotter.XYs, len(graph.Values))
		for i, val := range graph.Values {
			dots[i].X = float64(val.Timestamp-start) / 1000
			dots[i].Y = float64(val.Value)
		}
		pltr, err := plotter.NewLine(dots)

		lineColor := color.RGBA{
			A: 1,
			R: uint8(rand.Intn(255)),
			G: uint8(rand.Intn(255)),
			B: uint8(rand.Intn(255)),
		}

		pltr.LineStyle = draw.LineStyle{
			Color:    lineColor,
			Width:    vg.Points(3),
			Dashes:   []vg.Length{},
			DashOffs: 0,
		}
		if err != nil {
			return nil, err
		}
		p.Add(pltr)
	}
	p.Add(plotter.NewGrid())

	return p.WriterTo(20*vg.Centimeter, 10*vg.Centimeter, "png")
}
