package renderer

import (
	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

var red = color.New(color.FgHiRed)

type ColorField string

const (
	ColorHeader  ColorField = "header"
	ColorColumn  ColorField = "column"
	ColorCaption ColorField = "caption"
)

func GetTableStyle(s *config.Config) renderer.ColorizedConfig {
	colorCfg := renderer.ColorizedConfig{}
	colorCfg.Header = renderer.Tint{FG: renderer.Colors{getHeaderColor(s.Table.Color.Header)}}
	colorCfg.Column = renderer.Tint{FG: renderer.Colors{getColumnColor(s.Table.Color.Column)}}
	colorCfg.Border = renderer.Tint{FG: renderer.Colors{color.FgWhite}}
	colorCfg.Separator = renderer.Tint{FG: renderer.Colors{color.FgWhite}}
	colorCfg.Symbols = tw.NewSymbols(resolveStyle(s.Table.Style))
	return colorCfg
}

func getHeaderColor(c config.TableColor) color.Attribute {
	return resolveColor(c, ColorHeader)
}

func getColumnColor(c config.TableColor) color.Attribute {
	return resolveColor(c, ColorColumn)
}

func getCaptionColor(c config.TableColor) color.Attribute {
	return resolveColor(c, ColorCaption)
}

func resolveColor(c config.TableColor, f ColorField) color.Attribute {
	if c == config.FgDefault {
		return defaultColorForField(f)
	}
	if col, ok := tableColorMap[c]; ok {
		return col
	}
	return defaultColorForField(f)
}

func defaultColorForField(f ColorField) color.Attribute {
	if f == ColorHeader {
		return color.FgCyan
	}
	return color.FgWhite
}

func resolveStyle(s config.TableStyle) tw.BorderStyle {
	if style, ok := tableStyleMap[s]; ok {
		return style
	}
	return tw.StyleDefault
}

var tableStyleMap = map[config.TableStyle]tw.BorderStyle{
	config.StyleNone:        tw.StyleNone,
	config.StyleASCII:       tw.StyleASCII,
	config.StyleLight:       tw.StyleLight,
	config.StyleHeavy:       tw.StyleHeavy,
	config.StyleDouble:      tw.StyleDouble,
	config.StyleDoubleLong:  tw.StyleDoubleLong,
	config.StyleLightHeavy:  tw.StyleLightHeavy,
	config.StyleHeavyLight:  tw.StyleHeavyLight,
	config.StyleLightDouble: tw.StyleLightDouble,
	config.StyleDoubleLight: tw.StyleDoubleLight,
	config.StyleRounded:     tw.StyleRounded,
	config.StyleMarkdown:    tw.StyleMarkdown,
	config.StyleGraphical:   tw.StyleGraphical,
	config.StyleMerger:      tw.StyleMerger,
	config.StyleDefault:     tw.StyleDefault,
	config.StyleDotted:      tw.StyleDotted,
	config.StyleArrow:       tw.StyleArrow,
	config.StyleStarry:      tw.StyleStarry,
	config.StyleHearts:      tw.StyleHearts,
	config.StyleCircuit:     tw.StyleCircuit,
	config.StyleNature:      tw.StyleNature,
	config.StyleArtistic:    tw.StyleArtistic,
	config.Style8Bit:        tw.Style8Bit,
	config.StyleChaos:       tw.StyleChaos,
	config.StyleDots:        tw.StyleDots,
	config.StyleBlocks:      tw.StyleBlocks,
	config.StyleZen:         tw.StyleZen,
	config.StyleVintage:     tw.StyleVintage,
	config.StyleSketch:      tw.StyleSketch,
	config.StyleArrowDouble: tw.StyleArrowDouble,
	config.StyleCelestial:   tw.StyleCelestial,
	config.StyleCyber:       tw.StyleCyber,
	config.StyleRunic:       tw.StyleRunic,
	config.StyleIndustrial:  tw.StyleIndustrial,
	config.StyleInk:         tw.StyleInk,
	config.StyleArcade:      tw.StyleArcade,
	config.StyleBlossom:     tw.StyleBlossom,
	config.StyleFrosted:     tw.StyleFrosted,
	config.StyleMosaic:      tw.StyleMosaic,
	config.StyleUFO:         tw.StyleUFO,
	config.StyleSteampunk:   tw.StyleSteampunk,
	config.StyleGalaxy:      tw.StyleGalaxy,
	config.StyleJazz:        tw.StyleJazz,
	config.StylePuzzle:      tw.StylePuzzle,
	config.StyleHypno:       tw.StyleHypno,
}

var tableColorMap = map[config.TableColor]color.Attribute{
	config.FgBlack:     color.FgBlack,
	config.FgRed:       color.FgRed,
	config.FgGreen:     color.FgGreen,
	config.FgYellow:    color.FgYellow,
	config.FgBlue:      color.FgBlue,
	config.FgMagenta:   color.FgMagenta,
	config.FgCyan:      color.FgCyan,
	config.FgWhite:     color.FgWhite,
	config.FgHiBlack:   color.FgHiBlack,
	config.FgHiRed:     color.FgHiRed,
	config.FgHiGreen:   color.FgHiGreen,
	config.FgHiYellow:  color.FgHiYellow,
	config.FgHiBlue:    color.FgHiBlue,
	config.FgHiMagenta: color.FgHiMagenta,
	config.FgHiCyan:    color.FgHiCyan,
	config.FgHiWhite:   color.FgHiWhite,
}
