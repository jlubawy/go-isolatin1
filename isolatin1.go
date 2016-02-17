package isolatin1

import (
	"fmt"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type InvalidPolicy bool

const (
	InvalidError InvalidPolicy = false
	InvalidSkip                = true
)

type ErrInvalidISOLatin1 struct {
	r rune
}

func (err *ErrInvalidISOLatin1) Error() string {
	return fmt.Sprintf("isolatin1: invalid rune %#U (%X)", err.r, err.r)
}

func Valid(c byte) bool {
	return (c >= 32 && c <= 126) || (c >= 160 && c <= 255)
}

type isolatin1Encoding struct {
	skipInvalid InvalidPolicy
}

func ISOLatin1(skipInvalid InvalidPolicy) encoding.Encoding {
	return isolatin1Encoding{
		skipInvalid: skipInvalid,
	}
}

func (isolatin1Encoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{
		Transformer: &isolatin1Decoder{},
	}
}

func (e isolatin1Encoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{
		Transformer: transform.Chain(norm.NFC, &encoding.Encoder{
			Transformer: &isolatin1Encoder{
				skipInvalid: e.skipInvalid,
			},
		}),
	}
}

type isolatin1Decoder struct{}

func (*isolatin1Decoder) Transform(dst, src []byte, atEOF bool) (int, int, error) {
	return 0, 0, fmt.Errorf("isolatin1: decoder not implemented")
}

func (*isolatin1Decoder) Reset() {
	return
}

type isolatin1Encoder struct {
	skipInvalid InvalidPolicy
}

func (enc *isolatin1Encoder) Transform(dst, src []byte, atEOF bool) (iDst int, iSrc int, err error) {
	nDst := len(dst)
	nSrc := len(src)

	iDst = 0
	iSrc = 0

	if nDst < 1 {
		return 0, 0, transform.ErrShortDst
	}

	for iDst < nDst && iSrc < nSrc {
		if c := src[iSrc]; c < utf8.RuneSelf {
			dst[iDst] = c
			iDst++
			iSrc++
			continue
		}

		r, size := utf8.DecodeRune(src[iSrc:])
		if size == 1 {
			// All valid runes of size 1 (those below utf8.RuneSelf) were
			// handled above. We have invalid UTF-8 or we haven't seen the
			// full character yet.
			err = encoding.ErrInvalidUTF8
			if !atEOF && !utf8.FullRune(src[iSrc:]) {
				err = transform.ErrShortSrc
			}
			return iDst, iSrc, err
		}

		iSrc = iSrc + size

		if iDst >= nDst {
			return iDst, iSrc, transform.ErrShortDst
		}

		c := byte(r)
		if !Valid(c) {
			if enc.skipInvalid {
				continue
			}

			return iDst, iSrc, &ErrInvalidISOLatin1{
				r: r,
			}
		}

		dst[iDst] = c
		iDst++
	}

	return iDst, iSrc, err
}

func (*isolatin1Encoder) Reset() {
	return
}
