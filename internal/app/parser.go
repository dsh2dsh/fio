package app

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

func skipBOM(r io.Reader) (*bufio.Reader, error) {
	br := bufio.NewReader(r)
	if r, _, err := br.ReadRune(); err != nil {
		return nil, fmt.Errorf("skip BOM: %w", err)
	} else if r != rune(0xFEFF) {
		if err := br.UnreadRune(); err != nil {
			return nil, fmt.Errorf("skip BOM: %w", err)
		}
	}
	return br, nil
}

func NewParser(file io.Reader) (*Parser, error) {
	// skip BOM
	br, err := skipBOM(file)
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(br)
	csvReader.Comma = ';'
	csvReader.ReuseRecord = true

	p := &Parser{csvReader: csvReader}
	if err := p.parseHeader(); err != nil {
		return nil, err
	}

	return p, nil
}

type Parser struct {
	csvReader *csv.Reader
	headers   []string
}

func (self *Parser) parseHeader() error {
	r, err := self.csvReader.Read()
	if err != nil {
		return fmt.Errorf("csv header: %w", err)
	}
	self.headers = make([]string, 0, len(r))
	self.headers = append(self.headers, r...)
	return nil
}

func (self *Parser) Next() (rec Record, err error) {
	r, err := self.csvReader.Read()
	if errors.Is(err, io.EOF) {
		return rec, nil
	} else if err != nil {
		return
	}
	line, _ := self.csvReader.FieldPos(0)

	fields := map[string]string{}
	for i, s := range r {
		name := self.headers[i]
		fields[name] = s
	}
	err = rec.Parse(fields, line)

	return
}

// --------------------------------------------------

type Record struct {
	date      time.Time
	accountId string
	note      string
	vs        string

	line  int
	money float64
	valid bool
}

func (self *Record) Valid() bool {
	return self.valid
}

func (self *Record) Parse(fields map[string]string, line int) error {
	self.line = line

	if err := self.parseDate(fields); err != nil {
		return err
	}

	if err := self.parseMoney(fields); err != nil {
		return err
	}

	self.valid = true
	self.parseAccount(fields)
	self.parseNote(fields)
	self.vs = fields["VS"]

	return nil
}

func (self *Record) parseDate(fields map[string]string) error {
	if date, err := time.Parse("02.01.2006", fields["Datum"]); err != nil {
		return fmt.Errorf("parse date, line %d: %w", self.line, err)
	} else {
		self.date = date
	}
	return nil
}

func (self *Record) parseMoney(fields map[string]string) error {
	moneyStr := fields["Objem"]
	moneyStr = strings.ReplaceAll(moneyStr, ",", ".")
	money, err := strconv.ParseFloat(moneyStr, 32)
	if err != nil {
		return fmt.Errorf("parse float %q: %w", moneyStr, err)
	}
	self.money = money
	return nil
}

func (self *Record) parseAccount(fields map[string]string) {
	if fields["Protiúčet"] != "" && fields["Kód banky"] != "" {
		self.accountId = fields["Protiúčet"] + "/" + fields["Kód banky"]
	} else {
		self.accountId = fields["Protiúčet"]
	}
}

func (self *Record) parseNote(fields map[string]string) {
	switch {
	case self.accountId != "" && fields["Poznámka"] != "":
		self.note = fields["Poznámka"]
	case fields["Zpráva pro příjemce"] == "" && fields["Poznámka"] != "":
		self.note = fields["Poznámka"]
	default:
		fallback := []string{
			self.accountId, fields["Zpráva pro příjemce"], fields["Poznámka"],
			fields["Typ"],
		}
		for _, v := range fallback {
			if v != "" {
				self.note = v
				break
			}
		}
	}
}

func (self *Record) Line() int {
	return self.line
}

func (self *Record) Out() bool {
	return self.money < 0
}

func (self *Record) Money() float32 {
	return float32(math.Abs(self.money))
}

func (self *Record) Date() time.Time {
	return self.date
}

func (self *Record) AccountId() string {
	return self.accountId
}

func (self *Record) Note() string {
	return self.note
}

func (self *Record) Vs() string {
	return self.vs
}

func (self *Record) Between(d1 time.Time, d2 time.Time) bool {
	if self.Date().Before(d1) {
		return false
	} else if !d2.IsZero() && self.Date().After(d2) {
		return false
	}
	return true
}
