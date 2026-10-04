package pkg

import (
	"fmt"
	"math/bits"
)

type Bitboard struct {
	data uint64
}

func NewBitboard(val uint64) Bitboard {
	positions := []int{}

	curr := val
	for i := range 64 {
		if curr&1 == 1 {
			positions = append(positions, i)
		}
		curr >>= 1
	}

	return Bitboard{val}
}

func (b *Bitboard) AllOn() {
	b.data = ^uint64(0)
}

func (b *Bitboard) GetAll() []int {
	positions := []int{}

	x := b.data

	for x != 0 {
		pos := bits.LeadingZeros64(x) // square index, MSB = square 0
		positions = append(positions, pos)
		x &^= uint64(1) << (63 - pos) // clear that bit
	}

	return positions
}

func (b *Bitboard) AllOff() {
	b.data = 0
}

func (b *Bitboard) FirstPiece() int {
	return bits.LeadingZeros64(b.data)
}

func (b *Bitboard) On(idx int) {
	mask := uint64(1) << (63 - idx)
	b.data = b.data | mask
}

func (b *Bitboard) Off(idx int) {
	mask := ^uint64(0) ^ (uint64(1) << (63 - idx))
	b.data = b.data & mask
}

func (b *Bitboard) Get(idx int) int {
	return int((b.data >> (63 - idx)) & 1)
}

func (b *Bitboard) Print() {
	for i := range 64 {
		if b.Get(i) == 0 {
			fmt.Print("0")
		} else {
			fmt.Print("1")
		}
	}
	fmt.Print("\n")
}
