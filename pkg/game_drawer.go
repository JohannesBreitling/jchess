package pkg

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

type Drawer struct {
	Board *Board
}

const tileSize = 120
const offset = 120

func NewDrawer(board *Board) Drawer {
	d := Drawer{}
	d.Board = board
	return d
}

func (d *Drawer) drawPromotion(screen *ebiten.Image) {
	var prefix string
	if d.Board.WhitesTurn {
		prefix = "w"
	} else {
		prefix = "b"
	}

	queen_file, err := os.Open("./assets/" + prefix + "-queen.png")
	if err != nil {
		log.Fatal(err)
	}

	rook_file, err := os.Open("./assets/" + prefix + "-rook.png")
	if err != nil {
		log.Fatal(err)
	}

	knight_file, err := os.Open("./assets/" + prefix + "-knight.png")
	if err != nil {
		log.Fatal(err)
	}

	bishop_file, err := os.Open("./assets/" + prefix + "-bishop.png")
	if err != nil {
		log.Fatal(err)
	}

	queen_bytes, err := png.Decode(queen_file)
	if err != nil {
		log.Fatal(err)
	}

	rook_bytes, err := png.Decode(rook_file)
	if err != nil {
		log.Fatal(err)
	}

	bishop_bytes, err := png.Decode(bishop_file)
	if err != nil {
		log.Fatal(err)
	}

	knight_bytes, err := png.Decode(knight_file)
	if err != nil {
		log.Fatal(err)
	}

	queen_img := ebiten.NewImageFromImage(queen_bytes)
	rook_img := ebiten.NewImageFromImage(rook_bytes)
	bishop_img := ebiten.NewImageFromImage(bishop_bytes)
	knight_img := ebiten.NewImageFromImage(knight_bytes)

	x := float64(280)
	y := float64(1220)

	bound := queen_img.Bounds()
	scale := float64(tileSize) / float64(bound.Dx())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	screen.DrawImage(queen_img, op)
	op.GeoM.Translate(tileSize, 0)
	screen.DrawImage(rook_img, op)
	op.GeoM.Translate(tileSize, 0)
	screen.DrawImage(bishop_img, op)
	op.GeoM.Translate(tileSize, 0)
	screen.DrawImage(knight_img, op)
}

func (d *Drawer) drawCheck(screen *ebiten.Image) {
	if !(d.Board.WhiteInCheck || d.Board.BlackInCheck) {
		return
	}

	fmt.Println("Whites Turn", d.Board.WhitesTurn)
	fmt.Println("White in Check:", d.Board.WhiteInCheck)
	fmt.Println("Black in Check:", d.Board.BlackInCheck)

	// Get the kings position
	var board Bitboard
	if d.Board.WhitesTurn {
		board = d.Board.Pieces[W_king]
	} else {
		board = d.Board.Pieces[B_king]
	}

	pos := board.FirstPiece()

	x := offset + 60 + (pos%8)*tileSize
	y := offset + 60 + (7-pos/8)*tileSize
	vector.FillCircle(screen, float32(x), float32(y), 55, color.RGBA{224, 101, 101, 185}, true)
}

func (d *Drawer) drawPieces(screen *ebiten.Image, asset_path string, coords []int) {
	asset_file, err := os.Open(asset_path)
	if err != nil {
		log.Fatal(err)
	}
	defer asset_file.Close()

	raw_img, err := png.Decode(asset_file)
	if err != nil {
		log.Fatal(err)
	}

	img := ebiten.NewImageFromImage(raw_img)
	for _, pos := range coords {
		op := &ebiten.DrawImageOptions{}

		pos_x := (pos % 8)
		pos_y := 7 - (pos / 8)

		x := float64(offset + tileSize*pos_x)
		y := float64(offset + tileSize*pos_y)

		bound := img.Bounds()
		scale := float64(tileSize) / float64(bound.Dx())
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x, y)
		screen.DrawImage(img, op)
	}
}

func (d *Drawer) DrawBoard(screen *ebiten.Image) {
	d.emptyBoard(screen)

	d.drawCheck(screen)

	// Draw white pieces
	w_knights := d.Board.GetCoords(W_knights)
	d.drawPieces(screen, "./assets/w-knight.png", w_knights)

	w_pawns := d.Board.GetCoords(W_pawns)
	d.drawPieces(screen, "./assets/w-pawn.png", w_pawns)

	w_king := d.Board.GetCoords(W_king)
	d.drawPieces(screen, "./assets/w-king.png", w_king)

	w_queen := d.Board.GetCoords(W_queen)
	d.drawPieces(screen, "./assets/w-queen.png", w_queen)

	w_rooks := d.Board.GetCoords(W_rooks)
	d.drawPieces(screen, "./assets/w-rook.png", w_rooks)

	w_bishops := d.Board.GetCoords(W_bishops)
	d.drawPieces(screen, "./assets/w-bishop.png", w_bishops)

	// Draw black pieces
	b_knights := d.Board.GetCoords(B_knights)
	d.drawPieces(screen, "./assets/b-knight.png", b_knights)

	b_pawns := d.Board.GetCoords(B_pawns)
	d.drawPieces(screen, "./assets/b-pawn.png", b_pawns)

	b_king := d.Board.GetCoords(B_king)
	d.drawPieces(screen, "./assets/b-king.png", b_king)

	b_queen := d.Board.GetCoords(B_queen)
	d.drawPieces(screen, "./assets/b-queen.png", b_queen)

	b_rooks := d.Board.GetCoords(B_rooks)
	d.drawPieces(screen, "./assets/b-rook.png", b_rooks)

	b_bishops := d.Board.GetCoords(B_bishops)
	d.drawPieces(screen, "./assets/b-bishop.png", b_bishops)

	if d.Board.Promoting {
		d.drawPromotion(screen)
	}
}

const num_string = "87654321"
const char_string = "ABCDEFGH"

func (d *Drawer) emptyBoard(screen *ebiten.Image) {

	highlight := false

	colorDark := color.RGBA{R: 181, G: 136, B: 99, A: 255}
	colorLight := color.RGBA{R: 240, G: 217, B: 181, A: 255}
	colorBackground := color.RGBA{R: 242, G: 227, B: 203, A: 255}
	colorActive := color.RGBA{R: 180, G: 230, B: 172, A: 255}
	colorMove := color.RGBA{84, 160, 166, 155}
	colorLastMoved := color.RGBA{155, 173, 127, 205}
	colorHighlight := color.RGBA{237, 159, 159, 205}

	// Fill the screen with the default color
	screen.Fill(colorBackground)

	// Create the dark and light tiles
	tileDark := ebiten.NewImage(tileSize, tileSize)
	tileDark.Fill(colorDark)
	tileLight := ebiten.NewImage(tileSize, tileSize)
	tileLight.Fill(colorLight)

	tileActive := ebiten.NewImage(tileSize, tileSize)
	tileActive.Fill(colorActive)
	tileMove := ebiten.NewImage(tileSize, tileSize)
	tileMove.Fill(colorMove)
	tileLastMoved := ebiten.NewImage(tileSize, tileSize)
	tileLastMoved.Fill(colorLastMoved)

	tileHighlight := ebiten.NewImage(tileSize, tileSize)
	tileHighlight.Fill(colorHighlight)

	face := text.NewGoXFace(basicfont.Face7x13)

	// Draw the board labels
	for i := range 8 {
		c := string(char_string[i])
		w, h := text.Measure(c, face, 0)
		x := offset + float64(tileSize*i) + float64(tileSize-4*w)/2
		y := float64(tileSize-4*h) / 2

		op1 := &text.DrawOptions{}
		op2 := &text.DrawOptions{}
		op1.GeoM.Scale(4, 4)
		op2.GeoM.Scale(4, 4)

		op1.GeoM.Translate(x, y+20)
		op2.GeoM.Translate(x, 1200-offset+y-20)
		text.Draw(screen, c, face, op1)
		text.Draw(screen, c, face, op2)
	}
	for i := range 8 {
		c := string(num_string[i])
		w, h := text.Measure(c, face, 0)
		x := float64(tileSize-4*w) / 2
		y := float64(tileSize*i) + float64(tileSize-4*h)/2 + offset

		op1 := &text.DrawOptions{}
		op2 := &text.DrawOptions{}
		op1.GeoM.Scale(4, 4)
		op2.GeoM.Scale(4, 4)

		op1.GeoM.Translate(x+20, y)
		op2.GeoM.Translate(1200-offset+x-20, y)

		text.Draw(screen, c, face, op1)
		text.Draw(screen, c, face, op2)
	}

	moves_to := []int{}
	for _, move := range d.Board.Moves {
		moves_to = append(moves_to, move.To())
	}

	// Draw the board tiles
	for y_it := range 8 {
		y := 7 - y_it
		for x_it := range 8 {
			x := x_it
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(offset+x*tileSize), float64(offset+y*tileSize))

			pos := x_it + y_it*8

			if (x_it+y_it)%2 == 1 {
				screen.DrawImage(tileLight, op)
			} else {
				screen.DrawImage(tileDark, op)
			}

			if pos == d.Board.LastFrom || pos == d.Board.LastTo {
				screen.DrawImage(tileLastMoved, op)
			}

			if pos == d.Board.Active {
				screen.DrawImage(tileActive, op)
			}

			if slices.Contains(moves_to, pos) {
				screen.DrawImage(tileMove, op)
			}

			if highlight && d.Board.WhitesTurn && d.Board.AttackedByBlack.Get(pos) == 1 {
				screen.DrawImage(tileHighlight, op)
			}

			if highlight && !d.Board.WhitesTurn && d.Board.AttackedByWhite.Get(pos) == 1 {
				screen.DrawImage(tileHighlight, op)
			}
		}
	}

	// Draw turn
	x := float32(160)
	y := float32(1280)
	if d.Board.WhitesTurn {
		vector.FillCircle(screen, x, y, 40, color.RGBA{255, 255, 255, 255}, true)
	} else {
		vector.FillCircle(screen, x, y, 40, color.RGBA{0, 0, 0, 255}, true)
	}
}
