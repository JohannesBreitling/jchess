package pkg

import "fmt"

// Layout of a move:
// 6 bits from
// 6 bits to
// 1 bit en passant
// 3 bits promotion
//
// pppefffffftttttt
type LegalMove struct {
	move uint16
}

func NewMove(from int, to int, enpassant bool, promote int) LegalMove {
	var enpassant_mask uint16
	if enpassant {
		enpassant_mask = uint16(1) << 12
	} else {
		enpassant_mask = 0
	}

	move := uint16(from)<<6 | uint16(to) | enpassant_mask | uint16(promote)<<13
	return LegalMove{move}
}

func (m *LegalMove) get(pos int) int {
	return int(m.move) >> (15 - pos)
}

func (m *LegalMove) EnPassant() bool {
	return m.get(3) == 1
}

func (m *LegalMove) From() int {
	mask := uint16(63) << 6
	return int(m.move&mask) >> 6
}

func (m *LegalMove) To() int {
	mask := uint16(63)
	return int(m.move & mask)
}

func (m *LegalMove) Promotion() int {
	return int(m.move >> 13)
}

func (m *LegalMove) SetTo(to int) {
	mask := ^uint16(63)
	m.move = (m.move & mask) | uint16(to)
}

func (m *LegalMove) Print() {
	fmt.Printf("From: %d, To: %d, Enpassant: %t, Promotion: %d\n", m.From(), m.To(), m.EnPassant(), m.Promotion())
}
