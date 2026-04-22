package lexer

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"unicode"
)

type Lexer struct {
	input *bufio.Reader
}

func New(input io.Reader) *Lexer {
	return &Lexer{
		input: bufio.NewReader(input),
	}
}

func (l *Lexer) Next() (Token, error) {
	var token Token

	if err := l.skipWhitespace(); err != nil {
		return token, err
	}

	ch, err := l.readRune()
	if err != nil {
		return token, err
	}

	switch ch {
	case ',':
		token.Type = TokenTypeComma
	case '+':
		token.Type = TokenTypePlus
	case '-':
		token.Type = TokenTypeMinus
	case '*':
		token.Type = TokenTypeAsterisk
	case '/':
		token.Type = TokenTypeSlash
	case '=':
		token.Type = TokenTypeEqual
	case '>':
		token.Type = TokenTypeGreaterThan
	case '(':
		token.Type = TokenTypeLeftParen
	case ')':
		token.Type = TokenTypeRightParen
	case 0:
		token.Type = TokenTypeEOF
	default:
		if err := l.unreadRune(ch); err != nil {
			return token, err
		}

		if unicode.IsDigit(ch) {
			value, err := l.readInt()
			if err != nil {
				return token, err
			}

			token.Type = TokenTypeInt
			token.Value = value
		} else {
			value, err := l.readIdent()
			if err != nil {
				return token, err
			}

			token.Type = TokenTypeIdent
			token.Value = value
		}
	}

	return token, nil
}

func (l *Lexer) readByte() (byte, error) {
	ch, err := l.input.ReadByte()
	if errors.Is(err, io.EOF) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return ch, err
}

func (l *Lexer) readRune() (rune, error) {
	ch, _, err := l.input.ReadRune()
	if errors.Is(err, io.EOF) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return ch, err
}

func (l *Lexer) unreadByte(ch byte) error {
	if ch != 0 {
		if err := l.input.UnreadByte(); err != nil {
			return err
		}
	}

	return nil
}

func (l *Lexer) unreadRune(ch rune) error {
	if ch != 0 {
		if err := l.input.UnreadRune(); err != nil {
			return err
		}
	}

	return nil
}

func (l *Lexer) skipWhitespace() error {
	for {
		ch, err := l.readByte()
		if err != nil {
			return err
		}

		switch ch {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			return l.unreadByte(ch)
		}
	}
}

func (l *Lexer) readInt() (int64, error) {
	var buf []rune

	for {
		ch, err := l.readRune()
		if err != nil {
			return 0, err
		}

		if !unicode.IsDigit(ch) {
			if err := l.unreadRune(ch); err != nil {
				return 0, err
			}

			break
		}

		buf = append(buf, ch)
	}

	return strconv.ParseInt(string(buf), 10, 64)
}

func (l *Lexer) readIdent() (string, error) {
	var buf []rune

	for {
		ch, err := l.readRune()
		if err != nil {
			return "", err
		}

		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch)) {
			if err := l.unreadRune(ch); err != nil {
				return "", err
			}

			break
		}

		buf = append(buf, ch)
	}

	return string(buf), nil
}
