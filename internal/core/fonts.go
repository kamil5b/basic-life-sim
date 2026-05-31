package core

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	fontS  *text.GoTextFace // 12px
	fontM  *text.GoTextFace // 15px
	fontL  *text.GoTextFace // 20px
	fontXL *text.GoTextFace // 28px
)

func initFonts() error {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		return err
	}
	fontS = &text.GoTextFace{Source: src, Size: 12}
	fontM = &text.GoTextFace{Source: src, Size: 15}
	fontL = &text.GoTextFace{Source: src, Size: 20}
	fontXL = &text.GoTextFace{Source: src, Size: 28}
	return nil
}
