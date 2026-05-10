package styles

import (
	"image/color"
)

var (
	white      = color.NRGBA{255, 255, 255, 255}
	black      = color.NRGBA{0, 0, 0, 255}
	grayLight  = color.NRGBA{200, 200, 200, 255}
	gray       = color.NRGBA{150, 150, 150, 255}
	grayDark   = color.NRGBA{100, 100, 100, 255}
	grayDarker = color.NRGBA{50, 50, 50, 255}
	grayBg1    = color.NRGBA{20, 20, 20, 255}
	grayBg2    = color.NRGBA{35, 35, 35, 255}
)

// ThemeForProvider returns the Styles associated with the given provider
// ID. Unknown or empty provider IDs yield the default Charmtone Pantera
// theme.
func ThemeForProvider(providerID string) Styles {
	switch providerID {
	case "Duck":
		return HyperduckopsObsidiana()
	default:
		return CharmtonePantera()
	}
}

// CharmtonePantera returns the Charmtone dark theme. It's the default style
// for the UI.
func CharmtonePantera() Styles {
	return quickStyle(quickStyleOpts{
		primary:   white,
		secondary: grayLight,
		accent:    gray,
		keyword:   white,

		fgBase:       white,
		fgMoreSubtle: grayLight,
		fgSubtle:     gray,
		fgMostSubtle: grayDark,

		onPrimary: black,

		bgBase:         black,
		bgLeastVisible: grayBg1,
		bgLessVisible:  grayBg2,
		bgMostVisible:  grayDarker,

		separator: grayDark,

		destructive:       white,
		error:             white,
		warningSubtle:     grayLight,
		warning:           white,
		busy:              gray,
		info:              white,
		infoMoreSubtle:    grayLight,
		infoMostSubtle:    grayDark,
		success:           white,
		successMoreSubtle: grayLight,
		successMostSubtle: gray,
	})
}

// HyperduckopsObsidiana returns the Hyperduckops dark theme.
func HyperduckopsObsidiana() Styles {
	return quickStyle(quickStyleOpts{
		primary:   white,
		secondary: grayLight,
		accent:    gray,

		fgBase:       white,
		fgMoreSubtle: grayLight,
		fgSubtle:     gray,
		fgMostSubtle: grayDark,

		onPrimary: black,

		bgBase:         black,
		bgLeastVisible: grayBg1,
		bgLessVisible:  grayBg2,
		bgMostVisible:  grayDarker,

		separator: grayDark,

		destructive:       white,
		error:             white,
		warningSubtle:     grayLight,
		warning:           white,
		busy:              gray,
		info:              white,
		infoMoreSubtle:    grayLight,
		infoMostSubtle:    grayDark,
		success:           white,
		successMoreSubtle: grayLight,
		successMostSubtle: gray,
	})
}
