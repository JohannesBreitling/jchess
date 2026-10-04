package pkg

import (
	"fmt"
	"math"
	"slices"
)

var (
	W_pawns   = 0
	W_rooks   = 1
	W_bishops = 2
	W_knights = 3
	W_queen   = 4
	W_king    = 5

	B_pawns   = 6
	B_rooks   = 7
	B_bishops = 8
	B_knights = 9
	B_queen   = 10
	B_king    = 11
)

type Board struct {
	Active     int
	Moves      []LegalMove
	LegalMoves []LegalMove

	AttackedByWhite Bitboard
	AttackedByBlack Bitboard

	LastFrom int
	LastTo   int

	W_can_castle bool
	B_can_castle bool

	Rook_0_moved  bool
	Rook_7_moved  bool
	Rook_56_moved bool
	Rook_63_moved bool

	PawnDoubleMove Bitboard

	Promoting    bool
	WhitesTurn   bool
	WhiteInCheck bool
	BlackInCheck bool

	Pieces [12]Bitboard

	MoveCount int
}

const mask uint64 = 1

func (b *Board) get_coords(board Bitboard) []int {
	return board.GetAll()
}

func (b *Board) GetCoords(piece int) []int {
	return b.get_coords(b.Pieces[piece])
}

func (b *Board) WhitePieceOnPos(pos int) bool {
	whites := b.Pieces[W_pawns].Get(pos) | b.Pieces[W_bishops].Get(pos) | b.Pieces[W_king].Get(pos) | b.Pieces[W_knights].Get(pos) | b.Pieces[W_queen].Get(pos) | b.Pieces[W_rooks].Get(pos)
	return whites > 0
}

func (b *Board) BlackPieceOnPos(pos int) bool {
	blacks := b.Pieces[B_pawns].Get(pos) | b.Pieces[B_bishops].Get(pos) | b.Pieces[B_king].Get(pos) | b.Pieces[B_knights].Get(pos) | b.Pieces[B_queen].Get(pos) | b.Pieces[B_rooks].Get(pos)
	return blacks > 0
}

func (b *Board) PieceOnPos(pos int) bool {
	return b.WhitePieceOnPos(pos) || b.BlackPieceOnPos(pos)
}

var (
	Move     = 0
	Caputure = 1
	Castle   = 2
	Check    = 3
)

func (b *Board) Move(move LegalMove) int {
	b.MoveCount += 1
	r := b.applyMove(move)
	b.GenerateAllMoves()

	if len(b.LegalMoves) == 0 {
		if b.WhitesTurn && b.WhiteInCheck {
			fmt.Println("Checkmate: White lost!")
			return 0
		}

		if !b.WhitesTurn && b.BlackInCheck {
			fmt.Println("Checkmate: Black lost!")
			return 0
		}

		fmt.Println("Stalemate: No winner!")
	}

	b.Moves = []LegalMove{}

	if b.MoveCount >= 50 {
		fmt.Println("Draw: 50 moves without capture or pawn move!")
		return 0
	}

	// TODO: Check Draws...
	// - 50 Moves without Capture or Pawn Move -> done
	// - Repetition of same position >= 3
	// - Insufficient Material

	return r
}

func (b *Board) applyMove(move LegalMove) int {

	from := move.From()
	to := move.To()

	b.PawnDoubleMove.AllOff()

	// Check which piece is moving
	from_mask := ^uint64(0) ^ uint64((1 << (63 - from)))
	to_mask := uint64(1) << (63 - to)

	r := Move

	for i := range &b.Pieces {
		board := &b.Pieces[i]
		if (board.data >> (63 - to) & 1) > 0 {
			// Capture
			board.data = board.data ^ to_mask
			r = Caputure
			b.MoveCount = 0
		}

		if from == 0 || to == 0 {
			b.Rook_0_moved = true
		}
		if from == 7 || to == 7 {
			b.Rook_7_moved = true
		}
		if from == 56 || to == 56 {
			b.Rook_56_moved = true
		}
		if from == 63 || to == 63 {
			b.Rook_63_moved = true
		}

		// Make the move
		if (board.data >> (63 - from) & 1) > 0 {
			board.data = (board.data & from_mask) | to_mask

			// Check castle -> TODO: Turn switched
			if i == B_king && (from/8 == to/8) && (math.Abs(float64(from-to)) == 2) {
				if to == 62 {
					b.applyMove(NewMove(63, 61, false, 0))
				} else {
					b.applyMove(NewMove(56, 59, false, 0))
				}
				r = Castle
			} else if i == B_king {
				// King move that is not castling
				b.B_can_castle = false
			}

			// Check castle
			if i == W_king && (from/8 == to/8) && (math.Abs(float64(from-to)) == 2) {
				if to == 6 {
					b.applyMove(NewMove(7, 5, false, 0))
				} else {
					b.applyMove(NewMove(0, 3, false, 0))
				}
				r = Castle
			} else if i == W_king {
				// King move that is not castling
				b.W_can_castle = false
			}

			// Check pawn moved
			if i == W_pawns || i == B_pawns {
				b.MoveCount = 0
			}

			if i == W_pawns && to > 55 || i == B_pawns && to < 8 {
				b.Promoting = true
				b.Active = to
			}

			if (i == W_pawns || i == B_pawns) && math.Abs(float64(from-to)) == 16 {
				b.PawnDoubleMove.On(to)
			}
		}
	}

	if move.EnPassant() {
		diff := move.To() - move.From()
		delt := diff - 8*Sign(diff)
		rem_pos := move.From() + delt

		mask := uint64(1) << (63 - rem_pos)
		if b.WhitesTurn {
			b.Pieces[B_pawns].data = b.Pieces[B_pawns].data ^ mask
		} else {
			b.Pieces[W_pawns].data = b.Pieces[W_pawns].data ^ mask

		}
		r = Caputure
	}

	b.GenerateAttacks()

	if !b.Promoting {
		b.WhitesTurn = !b.WhitesTurn
	}

	b.CheckCheck()

	return r
}

// TODO: Board from fen string
// func NewBoard(pos string) (Board, error) {
// 	board := Board{}
// }

func DefaultBoard() Board {
	board := Board{}
	board.Pieces[W_rooks].data = (1 << 63) | (1 << 56)
	board.Pieces[W_pawns].data = 255 << 48
	board.Pieces[W_knights].data = (1 << 62) | (1 << 57)
	board.Pieces[W_queen].data = 1 << 60
	board.Pieces[W_king].data = 1 << 59
	board.Pieces[W_bishops].data = (1 << 61) | (1 << 58)

	board.Pieces[B_rooks].data = 1 | (1 << 7)
	board.Pieces[B_pawns].data = 255 << 8
	board.Pieces[B_knights].data = (1 << 1) | (1 << 6)
	board.Pieces[B_queen].data = 1 << 4
	board.Pieces[B_king].data = 1 << 3
	board.Pieces[B_bishops].data = (1 << 2) | (1 << 5)

	board.Active = 64
	board.Moves = []LegalMove{}
	board.LegalMoves = []LegalMove{}

	board.PawnDoubleMove = NewBitboard(0)

	board.B_can_castle = true
	board.W_can_castle = true
	board.WhitesTurn = true

	board.Rook_0_moved = false
	board.Rook_7_moved = false
	board.Rook_56_moved = false
	board.Rook_63_moved = false

	board.Promoting = false
	board.WhiteInCheck = false
	board.BlackInCheck = false

	board.LastFrom = 64
	board.LastTo = 64

	board.MoveCount = 0

	fmt.Println("Finished board creation")

	return board
}

func (b *Board) Promote(idx int) {
	var board *uint64
	var indices []int

	prom_mask := uint64(1) << (63 - b.Active)

	if b.WhitesTurn {
		indices = []int{W_queen, W_rooks, W_bishops, W_knights, W_pawns}
	} else {
		indices = []int{B_queen, B_rooks, B_bishops, B_knights, B_pawns}
	}

	board = &b.Pieces[indices[idx]].data

	b.Pieces[indices[4]].data = b.Pieces[indices[4]].data ^ prom_mask
	*board = *board | prom_mask

	b.WhitesTurn = !b.WhitesTurn
	b.Active = 64
	b.Promoting = false
}

func (b *Board) GenerateAllMoves() {
	white_was_in_check := b.WhiteInCheck
	black_was_in_check := b.BlackInCheck

	b.AttackedByWhite.AllOff()
	b.AttackedByBlack.AllOff()

	white_moves := []LegalMove{}
	white_pawn_moves := []LegalMove{}
	white_pawn_attacks := []LegalMove{}

	black_moves := []LegalMove{}
	black_pawn_moves := []LegalMove{}
	black_pawn_attacks := []LegalMove{}

	for i := range 64 {
		if b.WhitePieceOnPos(i) {
			if b.Pieces[W_pawns].Get(i) > 0 {
				white_pawn_moves = append(white_pawn_moves, b.generateWhiteMoves(i)...)
				white_pawn_attacks = append(white_pawn_attacks, b.generatePawnAttacks(i, true)...)
			} else {
				white_moves = append(white_moves, b.generateWhiteMoves(i)...)
			}
		} else if b.BlackPieceOnPos(i) {
			if b.Pieces[B_pawns].Get(i) > 0 {
				black_pawn_moves = append(black_pawn_moves, b.generateBlackMoves(i)...)
				black_pawn_attacks = append(black_pawn_attacks, b.generatePawnAttacks(i, false)...)
			} else {
				black_moves = append(black_moves, b.generateBlackMoves(i)...)
			}
		}
	}

	for _, move := range white_moves {
		b.AttackedByWhite.On(move.To())
	}

	for _, move := range black_moves {
		b.AttackedByBlack.On(move.To())
	}

	for _, move := range white_pawn_attacks {
		b.AttackedByWhite.On(move.To())
	}

	for _, move := range black_pawn_attacks {
		b.AttackedByBlack.On(move.To())
	}

	var all_moves []LegalMove
	if b.WhitesTurn {
		all_moves = white_moves
		all_moves = append(all_moves, white_pawn_moves...)
	} else {
		all_moves = black_moves
		all_moves = append(all_moves, black_pawn_moves...)
	}

	legal_moves := []LegalMove{}
	// Filter out the moves where still check
	for _, move := range all_moves {
		new_board := *b
		new_board.applyMove(move)

		if black_was_in_check && new_board.BlackInCheck {
			continue
		}

		if white_was_in_check && new_board.WhiteInCheck {
			continue
		}

		if b.WhitesTurn && new_board.WhiteInCheck {
			continue
		}

		if !b.WhitesTurn && new_board.BlackInCheck {
			continue
		}

		legal_moves = append(legal_moves, move)
	}

	b.LegalMoves = legal_moves
}

func (b *Board) GenerateAttacks() {
	b.AttackedByWhite.AllOff()
	b.AttackedByBlack.AllOff()

	white_moves := []LegalMove{}
	white_pawn_attacks := []LegalMove{}

	black_moves := []LegalMove{}
	black_pawn_attacks := []LegalMove{}

	for i := range 64 {
		if b.WhitePieceOnPos(i) {
			if b.Pieces[W_pawns].Get(i) > 0 {
				white_pawn_attacks = append(white_pawn_attacks, b.generatePawnAttacks(i, true)...)
			} else {
				white_moves = append(white_moves, b.generateWhiteMoves(i)...)
			}
		} else if b.BlackPieceOnPos(i) {

			if b.Pieces[B_pawns].Get(i) > 0 {
				black_pawn_attacks = append(black_pawn_attacks, b.generatePawnAttacks(i, false)...)
			} else {
				black_moves = append(black_moves, b.generateBlackMoves(i)...)
			}
		}
	}

	for _, move := range white_moves {
		b.AttackedByWhite.On(move.To())
	}

	for _, move := range black_moves {
		b.AttackedByBlack.On(move.To())
	}

	for _, move := range white_pawn_attacks {
		b.AttackedByWhite.On(move.To())
	}

	for _, move := range black_pawn_attacks {
		b.AttackedByBlack.On(move.To())
	}
}

func (b *Board) GenerateMoves(curr int) {
	b.Moves = []LegalMove{}
	var moves []LegalMove
	if b.WhitesTurn {
		moves = append(moves, b.generateWhiteMoves(curr)...)
	} else {
		moves = append(moves, b.generateBlackMoves(curr)...)
	}

	for _, move := range moves {
		if slices.Contains(b.LegalMoves, move) {
			b.Moves = append(b.Moves, move)
		}
	}
}

func (b *Board) checkTermination(pos int, from int, moves *[]LegalMove, white_to_move bool) bool {
	if b.WhitePieceOnPos(pos) {
		if white_to_move {
			return true
		} else {
			b.addMove(NewMove(from, pos, false, 0), moves, white_to_move)
			return true
		}
	}

	if b.BlackPieceOnPos(pos) {
		if white_to_move {
			b.addMove(NewMove(from, pos, false, 0), moves, white_to_move)
			return true
		} else {
			return true
		}
	}

	return false
}

func (b *Board) generateWhiteMoves(curr int) []LegalMove {
	moves := []LegalMove{}

	w_pawn_on_pos := b.Pieces[W_pawns].Get(curr) > 0
	w_king_on_pos := b.Pieces[W_king].Get(curr) > 0
	w_queen_on_pos := b.Pieces[W_queen].Get(curr) > 0
	w_rook_on_pos := b.Pieces[W_rooks].Get(curr) > 0
	w_knight_on_pos := b.Pieces[W_knights].Get(curr) > 0
	w_bishop_on_pos := b.Pieces[W_bishops].Get(curr) > 0

	if w_pawn_on_pos {
		moves = append(moves, b.generatePawnMoves(curr, 1, true)...)
	} else if w_king_on_pos {
		moves = append(moves, b.generateKing(curr, true)...)
	} else if w_queen_on_pos {
		moves = append(moves, b.generateStraightMoves(curr, true)...)
		moves = append(moves, b.generateDiagonalMoves(curr, true)...)
	} else if w_rook_on_pos {
		moves = append(moves, b.generateStraightMoves(curr, true)...)
	} else if w_knight_on_pos {
		moves = append(moves, b.generateKnightMoves(curr, true)...)
	} else if w_bishop_on_pos {
		moves = append(moves, b.generateDiagonalMoves(curr, true)...)
	}

	return moves
}

func (b *Board) generateBlackMoves(curr int) []LegalMove {
	moves := []LegalMove{}

	b_pawn_on_pos := b.Pieces[B_pawns].Get(curr) > 0
	b_king_on_pos := b.Pieces[B_king].Get(curr) > 0
	b_queen_on_pos := b.Pieces[B_queen].Get(curr) > 0
	b_rook_on_pos := b.Pieces[B_rooks].Get(curr) > 0
	b_knight_on_pos := b.Pieces[B_knights].Get(curr) > 0
	b_bishop_on_pos := b.Pieces[B_bishops].Get(curr) > 0

	if b_pawn_on_pos {
		moves = append(moves, b.generatePawnMoves(curr, -1, false)...)
	} else if b_king_on_pos {
		moves = append(moves, b.generateKing(curr, false)...)
	} else if b_queen_on_pos {
		moves = append(moves, b.generateStraightMoves(curr, false)...)
		moves = append(moves, b.generateDiagonalMoves(curr, false)...)
	} else if b_rook_on_pos {
		moves = append(moves, b.generateStraightMoves(curr, false)...)
	} else if b_knight_on_pos {
		moves = append(moves, b.generateKnightMoves(curr, false)...)
	} else if b_bishop_on_pos {
		moves = append(moves, b.generateDiagonalMoves(curr, false)...)
	}

	return moves
}

func (b *Board) addMove(move LegalMove, moves *[]LegalMove, white_to_move bool) {
	// Check if empty
	if white_to_move && b.WhitePieceOnPos(move.To()) {
		return
	}
	if !white_to_move && b.BlackPieceOnPos(move.To()) {
		return
	}

	*moves = append(*moves, move)
}

func (b *Board) generatePawnAttacks(curr int, white_to_move bool) []LegalMove {
	moves := []LegalMove{}
	col := curr % 8

	if white_to_move {
		if col > 0 && !b.WhitePieceOnPos(curr+7) {
			moves = append(moves, NewMove(curr, curr+7, false, 0))
		}

		if col < 7 && !b.WhitePieceOnPos(curr+9) {
			moves = append(moves, NewMove(curr, curr+9, false, 0))
		}
	} else {
		if col < 7 && !b.BlackPieceOnPos(curr-7) {
			moves = append(moves, NewMove(curr, curr-7, false, 0))
		}
		if col > 0 && !b.BlackPieceOnPos(curr-9) {
			moves = append(moves, NewMove(curr, curr-9, false, 0))
		}
	}

	return moves
}

func (b *Board) generatePawnMoves(curr int, dir int, white_to_move bool) []LegalMove {
	moves := &[]LegalMove{}

	row := curr / 8
	col := curr % 8

	if !b.PieceOnPos(curr + dir*8) {
		b.addMove(NewMove(curr, curr+dir*8, false, 0), moves, white_to_move)
	}

	if dir == 1 && curr < 16 && !b.PieceOnPos(curr+dir*16) {
		b.addMove(NewMove(curr, curr+dir*16, false, 0), moves, white_to_move)
	}

	if dir == -1 && curr > 47 && !b.PieceOnPos(curr+dir*16) {
		b.addMove(NewMove(curr, curr+dir*16, false, 0), moves, white_to_move)
	}

	potential_hits := []int{}
	if dir == 1 && row < 7 {
		var pos int
		if col > 0 {
			pos = 8*(row+1) + col - 1
			potential_hits = append(potential_hits, pos)
		}
		if col < 7 {
			pos = 8*(row+1) + col + 1
			potential_hits = append(potential_hits, pos)
		}
	}

	if dir == -1 && row > 0 {
		var pos int
		if col > 0 {
			pos = 8*(row-1) + col - 1
			potential_hits = append(potential_hits, pos)
		}
		if col < 7 {
			pos = 8*(row-1) + col + 1
			potential_hits = append(potential_hits, pos)
		}
	}

	for _, hit := range potential_hits {
		if b.BlackPieceOnPos(hit) && white_to_move {
			b.addMove(NewMove(curr, hit, false, 0), moves, white_to_move)
		}

		if b.WhitePieceOnPos(hit) && !white_to_move {
			b.addMove(NewMove(curr, hit, false, 0), moves, white_to_move)
		}
	}

	if white_to_move && row == 4 {
		if col > 0 && b.PawnDoubleMove.Get(curr-1) == 1 {
			move := NewMove(curr, curr+7, true, 0)
			b.addMove(move, moves, white_to_move)
		}
		if col < 7 && b.PawnDoubleMove.Get(curr+1) == 1 {
			move := NewMove(curr, curr+9, true, 0)
			b.addMove(move, moves, white_to_move)
		}
	} else if !white_to_move && row == 3 {
		if col > 0 && b.PawnDoubleMove.Get(curr-1) == 1 {
			move := NewMove(curr, curr-9, true, 0)
			b.addMove(move, moves, white_to_move)
		}
		if col < 7 && b.PawnDoubleMove.Get(curr+1) == 1 {
			move := NewMove(curr, curr-7, true, 0)
			b.addMove(move, moves, white_to_move)
		}
	}

	return *moves
}

func (b *Board) generateDiagonalMoves(curr int, white_to_move bool) []LegalMove {
	moves := &[]LegalMove{}
	row := curr / 8
	col := curr % 8

	curr_row := row + 1
	curr_col := col + 1
	for curr_row < 8 && curr_col < 8 {
		pos := 8*curr_row + curr_col
		if b.checkTermination(pos, curr, moves, white_to_move) {
			break
		}

		b.addMove(NewMove(curr, pos, false, 0), moves, white_to_move)
		curr_row += 1
		curr_col += 1
	}

	curr_row = row + 1
	curr_col = col - 1
	for curr_row < 8 && curr_col >= 0 {
		pos := 8*curr_row + curr_col
		if b.checkTermination(pos, curr, moves, white_to_move) {
			break
		}

		b.addMove(NewMove(curr, pos, false, 0), moves, white_to_move)
		curr_row += 1
		curr_col -= 1
	}

	curr_row = row - 1
	curr_col = col + 1
	for curr_row >= 0 && curr_col < 8 {
		pos := 8*curr_row + curr_col
		if b.checkTermination(pos, curr, moves, white_to_move) {
			break
		}

		b.addMove(NewMove(curr, pos, false, 0), moves, white_to_move)
		curr_row -= 1
		curr_col += 1
	}

	curr_row = row - 1
	curr_col = col - 1
	for curr_row >= 0 && curr_col >= 0 {
		pos := 8*curr_row + curr_col
		if b.checkTermination(pos, curr, moves, white_to_move) {
			break
		}

		b.addMove(NewMove(curr, pos, false, 0), moves, white_to_move)
		curr_row -= 1
		curr_col -= 1
	}

	return *moves
}

func (b *Board) generateStraightMoves(curr int, white_to_move bool) []LegalMove {
	moves := &[]LegalMove{}
	row := curr / 8
	col := curr % 8

	curr_col := col + 1
	curr_row := row
	for curr_col < 8 {
		pos := 8*curr_row + curr_col
		if b.checkTermination(pos, curr, moves, white_to_move) {
			break
		}
		b.addMove(NewMove(curr, pos, false, 0), moves, white_to_move)
		curr_col += 1
	}

	curr_col = col - 1
	curr_row = row
	for curr_col >= 0 {
		pos := 8*curr_row + curr_col
		if b.checkTermination(pos, curr, moves, white_to_move) {
			break
		}
		b.addMove(NewMove(curr, pos, false, 0), moves, white_to_move)
		curr_col -= 1
	}

	curr_col = col
	curr_row = row + 1
	for curr_row < 8 {
		pos := 8*curr_row + curr_col
		if b.checkTermination(pos, curr, moves, white_to_move) {
			break
		}
		b.addMove(NewMove(curr, pos, false, 0), moves, white_to_move)
		curr_row += 1
	}

	curr_col = col
	curr_row = row - 1
	for curr_row >= 0 {
		pos := 8*curr_row + curr_col
		if b.checkTermination(pos, curr, moves, white_to_move) {
			break
		}
		b.addMove(NewMove(curr, pos, false, 0), moves, white_to_move)
		curr_row -= 1
	}

	return *moves
}

func (b *Board) generateKing(curr int, white_to_move bool) []LegalMove {
	moves := &[]LegalMove{}

	row := curr / 8
	col := curr % 8

	move := NewMove(curr, curr, false, 0)

	if row > 0 {
		move.SetTo(curr - 8)
		b.addMove(move, moves, white_to_move)
	}

	if row < 7 {
		move.SetTo(curr + 8)
		b.addMove(move, moves, white_to_move)
	}

	if col > 0 {
		move.SetTo(curr - 1)
		b.addMove(move, moves, white_to_move)

		if row < 7 {
			move.SetTo(curr + 7)
			b.addMove(move, moves, white_to_move)
		}

		if row > 0 {
			move.SetTo(curr - 9)
			b.addMove(move, moves, white_to_move)
		}
	}

	if col < 7 {
		move.SetTo(curr + 1)
		b.addMove(move, moves, white_to_move)

		if row < 7 {
			move.SetTo(curr + 9)
			b.addMove(move, moves, white_to_move)
		}

		if row > 0 {
			move.SetTo(curr - 7)
			b.addMove(move, moves, white_to_move)
		}
	}

	if white_to_move && b.W_can_castle {
		if !b.WhiteInCheck && !b.PieceOnPos(curr+1) && !b.PieceOnPos(curr+2) && b.AttackedByBlack.Get(curr+1) == 0 && b.AttackedByBlack.Get(curr+2) == 0 {
			move.SetTo(curr + 2)
			b.addMove(move, moves, white_to_move)
		}

		if !b.WhiteInCheck && !b.PieceOnPos(curr-1) && !b.PieceOnPos(curr-2) && b.AttackedByBlack.Get(curr-1) == 0 && b.AttackedByBlack.Get(curr-2) == 0 {
			move.SetTo(curr - 2)
			b.addMove(move, moves, white_to_move)
		}
	}

	if !white_to_move && b.B_can_castle {
		if !b.BlackInCheck && !b.PieceOnPos(curr+1) && !b.PieceOnPos(curr+2) && b.AttackedByWhite.Get(curr+1) == 0 && b.AttackedByWhite.Get(curr+2) == 0 {
			move.SetTo(curr + 2)
			b.addMove(move, moves, white_to_move)
		}

		if !b.BlackInCheck && !b.PieceOnPos(curr-1) && !b.PieceOnPos(curr-2) && b.AttackedByWhite.Get(curr-1) == 0 && b.AttackedByWhite.Get(curr-2) == 0 {
			move.SetTo(curr - 2)
			b.addMove(move, moves, white_to_move)
		}

	}

	return *moves
}

func (b *Board) generateKnightMoves(curr int, white_to_move bool) []LegalMove {
	moves := &[]LegalMove{}

	row := curr / 8
	col := curr % 8

	move := NewMove(curr, curr, false, 0)

	if row < 7 {
		if col < 6 {
			// +1 +2
			move.SetTo(curr + 10)
			b.addMove(move, moves, white_to_move)
		}

		if col > 1 {
			// +1 -2
			move.SetTo(curr + 6)
			b.addMove(move, moves, white_to_move)
		}

		if row < 6 {
			if col < 7 {
				// +2 +1
				move.SetTo(curr + 17)
				b.addMove(move, moves, white_to_move)
			}
			if col > 0 {
				// +2 -1
				move.SetTo(curr + 15)
				b.addMove(move, moves, white_to_move)
			}
		}
	}

	if row > 0 {
		if col < 6 {
			// -1 +2
			move.SetTo(curr - 6)
			b.addMove(move, moves, white_to_move)
		}

		if col > 1 {
			// -1 -2
			move.SetTo(curr - 10)
			b.addMove(move, moves, white_to_move)
		}

		if row > 1 {
			if col < 7 {
				// -2 +1
				move.SetTo(curr - 15)
				b.addMove(move, moves, white_to_move)
			}

			if col > 0 {
				// -2 -1
				move.SetTo(curr - 17)
				b.addMove(move, moves, white_to_move)
			}
		}
	}

	return *moves
}

func (b *Board) CheckCheck() {
	b.WhiteInCheck = false
	b.BlackInCheck = false

	white_king_pos := b.Pieces[W_king].GetAll()[0]
	black_king_pos := b.Pieces[B_king].GetAll()[0]

	if b.AttackedByBlack.Get(white_king_pos) == 1 {
		b.WhiteInCheck = true
	}

	if b.AttackedByWhite.Get(black_king_pos) == 1 {
		b.BlackInCheck = true
	}
}
