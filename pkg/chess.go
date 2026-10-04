package pkg

type Chess struct {
	board       *Board
	MakeMove    bool
	audioPlayer SoundPlayer
}

func NewChess(board *Board, audioPlayer SoundPlayer) Chess {
	c := Chess{board: board, MakeMove: false, audioPlayer: audioPlayer}
	c.board.GenerateAllMoves()
	return c
}

func (c *Chess) GenerateMoves() {
	c.board.Moves = []LegalMove{}
	if c.board.Active == 64 {
		// No piece is activated
		return
	}
	c.board.GenerateMoves(c.board.Active)
}

func (c *Chess) Promote(piece int) {
	c.Promote(piece)
}

func (c *Chess) Move(move LegalMove) {
	from := move.From()
	to := move.To()

	c.board.LastFrom = from
	c.board.LastTo = to

	// Perform the move
	m := c.board.Move(move)

	if !c.board.Promoting {
		c.board.Active = 64
	}

	if c.board.WhitesTurn && c.board.WhiteInCheck || !c.board.WhitesTurn && c.board.BlackInCheck {
		m = Check
	}

	switch m {
	case Check:
		c.audioPlayer.PlayCheck()
		break
	case Caputure:
		c.audioPlayer.PlayCapture()
		break
	case Castle:
		c.audioPlayer.PlayCastle()
		break
	default:
		c.audioPlayer.PlayMove()

	}
}
