package main

import (
	"bytes"
	"fmt"
	"log"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"jchess/pkg"
)

type Game struct {
	Drawer *pkg.Drawer
	Board  *pkg.Board
	Chess  *pkg.Chess
}

func (g *GamePlayer) PlayMove() {
	g.movePlayer.Rewind()
	g.movePlayer.Play()
}

func (g *GamePlayer) PlayCapture() {
	g.capturePlayer.Rewind()
	g.capturePlayer.Play()
}

func (g *GamePlayer) PlayCheck() {
	g.checkPlayer.Rewind()
	g.checkPlayer.Play()
}

func (g *GamePlayer) PlayCastle() {
	g.castlePlayer.Rewind()
	g.castlePlayer.Play()
}

type GamePlayer struct {
	movePlayer    *audio.Player
	checkPlayer   *audio.Player
	capturePlayer *audio.Player
	castlePlayer  *audio.Player
}

//go:embed assets/move.mp3
var moveSound []byte

//go:embed assets/capture.mp3
var caputreSound []byte

//go:embed assets/check.mp3
var checkSound []byte

//go:embed assets/castle.mp3
var castleSound []byte

const offset = 120
const tileSize = 120

func NewGame() *Game {
	const sampleRate = 44100

	audioCtx := audio.NewContext(sampleRate)

	decodedMove, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(moveSound))
	if err != nil {
		log.Fatal(err)
	}

	decodedCheck, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(checkSound))
	if err != nil {
		log.Fatal(err)
	}

	decodedCapture, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(caputreSound))
	if err != nil {
		log.Fatal(err)
	}

	decodedCastle, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(castleSound))
	if err != nil {
		log.Fatal(err)
	}

	movePlayer, err := audioCtx.NewPlayer(decodedMove)
	if err != nil {
		log.Fatal(err)
	}

	checkPlayer, err := audioCtx.NewPlayer(decodedCheck)
	if err != nil {
		log.Fatal(err)
	}

	capturePlayer, err := audioCtx.NewPlayer(decodedCapture)
	if err != nil {
		log.Fatal(err)
	}

	castlePlayer, err := audioCtx.NewPlayer(decodedCastle)
	if err != nil {
		log.Fatal(err)
	}

	audioPlayer := &GamePlayer{
		movePlayer:    movePlayer,
		checkPlayer:   checkPlayer,
		capturePlayer: capturePlayer,
		castlePlayer:  castlePlayer,
	}

	board := pkg.DefaultBoard()
	chess := pkg.NewChess(&board, audioPlayer)
	drawer := pkg.NewDrawer(&board)

	game := &Game{
		Drawer: &drawer,
		Board:  &board,
		Chess:  &chess,
	}

	fmt.Println("Finished board game construction")

	return game
}

func (g *Game) Update() error {
	// Things only happen when the mouse is release by a player
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton0) {
		// Compute the coordinate of the click
		pressX, pressY := ebiten.CursorPosition()

		// Code for clicking a piece to promote
		if g.Board.Promoting && pressX-offset < 0 || pressY-offset < 0 || pressX > offset+8*tileSize || pressY > offset+8*tileSize {
			// Check promotion click
			if pressY >= 1220 && pressY <= 1340 && pressX >= 280 && pressX <= 760 {
				tile := (pressX - 280) / tileSize
				g.Chess.Promote(tile)
			}

			return nil
		} else if g.Board.Promoting {
			// Disable all other clicks when Promoting
			return nil
		}

		tileX := (pressX - offset) / tileSize
		tileY := 7 - ((pressY - offset) / tileSize)

		to := tileY*8 + tileX
		from := g.Board.Active

		if g.Chess.MakeMove {
			for _, move := range g.Board.Moves {
				if to == move.To() && from == move.From() {
					g.Chess.Move(move)
					return nil
				}
			}
		}

		// Activate the tile
		if g.Board.Active == to {
			g.Board.Active = 64
		} else {
			g.Board.Active = to
		}
		g.Chess.GenerateMoves()
		g.Chess.MakeMove = len(g.Board.Moves) > 0
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Drawer.DrawBoard(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1200, 1400
}

func main() {
	ebiten.SetWindowSize(900, 1050)
	ebiten.SetWindowTitle("jchess")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}

}
