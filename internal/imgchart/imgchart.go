package imgchart

import (
	"github.com/fpawel/elco.v2/internal/data"
	"github.com/wcharczuk/go-chart"

	"image"
)

func New(x data.FirmwareBytes, width, height int) image.Image {

	seriesFon := chart.ContinuousSeries{

		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.ColorBlue,
		},
	}

	tempPoints := x.TempPoints()

	seriesFon.XValues, seriesFon.YValues = tempPoints.Temp[62:190], tempPoints.Fon[62:190]

	seriesSens := chart.ContinuousSeries{
		YAxis: chart.YAxisSecondary,
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.ColorRed,
			//StrokeWidth:      chart.Disabled,

		},
	}
	seriesSens.XValues, seriesSens.YValues = tempPoints.Temp[62:190], tempPoints.Sens[62:190]

	gridStyle := chart.Style{
		Show:        true,
		StrokeColor: chart.ColorAlternateGray,
		StrokeWidth: 0.3,
	}

	vf1 := func(v interface{}) string {
		return chart.FloatValueFormatterWithFormat(v, "%2.1f")
	}
	vf0 := func(v interface{}) string {
		return chart.FloatValueFormatterWithFormat(v, "%2.0f")
	}

	graphChart := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{
				Left: 10,
			},
		},
		Width:  width,
		Height: height,
		XAxis: chart.XAxis{
			Name: "Т⁰C",
			NameStyle: chart.Style{
				FontColor: chart.ColorBlack,
				Show:      true,
				FontSize:  8,
			},
			Style: chart.Style{
				Show:     true, //enables / displays the x-axis
				FontSize: 8,
			},
			TickPosition:   chart.TickPositionBetweenTicks,
			Range:          &chart.ContinuousRange{Min: -60, Max: 60},
			GridMajorStyle: gridStyle,
			GridMinorStyle: gridStyle,
			ValueFormatter: vf0,
		},
		YAxis: chart.YAxis{
			ValueFormatter: vf1,
			Name:           "нА",
			NameStyle: chart.Style{
				FontColor:           chart.ColorBlue,
				Show:                true,
				FontSize:            8,
				TextRotationDegrees: 0.1,
			},
			Style: chart.Style{
				Show:        true, //enables / displays the x-axis
				FontSize:    8,
				StrokeColor: chart.ColorBlue,
			},
			Range:          seriesYRange1(seriesFon),
			GridMajorStyle: gridStyle,
			GridMinorStyle: gridStyle,
		},
		YAxisSecondary: chart.YAxis{
			ValueFormatter: vf1,
			Name:           "Кч,%",
			NameStyle: chart.Style{
				FontColor:           chart.ColorRed,
				Show:                true,
				FontSize:            8,
				TextRotationDegrees: 0.1,
			},
			Style: chart.Style{
				Show:        true, //enables / displays the secondary y-axis
				FontSize:    8,
				StrokeColor: chart.ColorRed,
			},
			Range:          seriesYRange1(seriesSens),
			GridMajorStyle: gridStyle,
			GridMinorStyle: gridStyle,
		},
		Series: []chart.Series{
			seriesFon,
			seriesSens,
		},
	}
	collector := &chart.ImageWriter{}

	if err := graphChart.Render(chart.PNG, collector); err != nil {
		panic(err)
	}
	img, err := collector.Image()
	if err != nil {
		panic(err)
	}
	return img
}

func minMaxFloat64(xs []float64) (min, max float64) {
	if len(xs) == 0 {
		return
	}

	max, min = xs[0], xs[0]
	for _, value := range xs {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

func seriesYRange1(x chart.ContinuousSeries) *chart.ContinuousRange {
	min, max := minMaxFloat64(x.YValues)
	if min == max {
		return &chart.ContinuousRange{Min: 0, Max: 1}
	}
	d := 0.1 * (max - min)
	return &chart.ContinuousRange{Min: min - d, Max: max + d}
}
