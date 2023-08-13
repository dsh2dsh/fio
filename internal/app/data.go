package app

import (
	"fmt"
	"sort"
	"time"
)

func NewReportData(cfg *Config) *ReportData {
	return &ReportData{
		sections: make(map[string]*Section),

		cfg: cfg,
	}
}

type ReportData struct {
	beginDate     time.Time
	endDate       time.Time
	monthsBetween int

	money    float32
	count    int
	sections map[string]*Section

	cfg *Config
}

func (self *ReportData) Count() int {
	return self.count
}

func (self *ReportData) Money() float32 {
	return self.money
}

func (self *ReportData) addRecord(sectName, sectKey string, rec Record) {
	self.updateTimes(rec)

	money := rec.Money()
	if !self.cfg.SkipFromSum(sectName) {
		self.count++
		self.money += money
	}

	self.addSection(sectName, money).addItem(sectKey, money)
}

func (self *ReportData) addSection(sectName string, money float32) *Section {
	sect := self.sections[sectName]
	if sect == nil {
		sect = newSection(sectName).withMonthsBetween(self.MonthsBetween).
			withSkipFromSum(self.cfg.SkipFromSum(sectName))
		self.sections[sectName] = sect
	}
	sect.Add(money)
	return sect
}

func (self *ReportData) updateTimes(rec Record) {
	if self.beginDate.IsZero() || rec.Date().Before(self.beginDate) {
		self.beginDate = rec.Date()
	}
	if self.endDate.IsZero() || rec.Date().After(self.endDate) {
		self.endDate = rec.Date()
	}
}

func (self *ReportData) SortedSections() []*Section {
	sections := make([]*Section, 0, len(self.sections))
	for _, sect := range self.sections {
		sections = append(sections, sect)
	}
	sort.Slice(sections, func(i, j int) bool {
		order1 := self.cfg.sectionIndex[sections[i].Name()].Order
		order2 := self.cfg.sectionIndex[sections[j].Name()].Order
		if order1 == order2 {
			return sections[i].Money() >= sections[j].Money()
		}
		return order1 < order2
	})
	return sections
}

func (self *ReportData) updateMonthsBetween() {
	beginYear, beginMonth, _ := self.beginDate.Date()
	endYear, endMonth, _ := self.endDate.Date()

	years := endYear - beginYear
	months := int(endMonth - beginMonth)
	if months < 0 {
		months += 12
		years--
	}

	self.monthsBetween = years*12 + months + 1
}

func (self *ReportData) BeginDateString() string {
	return self.beginDate.Format("2006-01-02")
}

func (self *ReportData) EndDateString() string {
	return self.endDate.Format("2006-01-02")
}

func (self *ReportData) PerMonthString(format string) string {
	if self.Count() < 2 || self.MonthsBetween() < 2 {
		return ""
	}
	return fmt.Sprintf(format, self.Money()/float32(self.MonthsBetween()))
}

func (self *ReportData) finish() {
	self.updateMonthsBetween()
}

func (self *ReportData) MonthsBetween() int {
	return self.monthsBetween
}

// --------------------------------------------------

func newSection(name string) *Section {
	return &Section{
		name:  name,
		items: make(map[string]*SectionItem),
	}
}

type Section struct {
	name  string
	money float32
	count int

	monthsBetween func() int
	skipFromSum   bool

	items map[string]*SectionItem
}

func (self *Section) withMonthsBetween(m func() int) *Section {
	self.monthsBetween = m
	return self
}

func (self *Section) withSkipFromSum(v bool) *Section {
	self.skipFromSum = v
	return self
}

func (self *Section) Name() string {
	return self.name
}

func (self *Section) Money() float32 {
	return self.money
}

func (self *Section) Count() int {
	return self.count
}

func (self *Section) MonthsBetween() int {
	return self.monthsBetween()
}

func (self *Section) Add(money float32) {
	self.count++
	self.money += money
}

func (self *Section) SortedItems() []*SectionItem {
	sortedItems := make([]*SectionItem, 0, len(self.items))
	for _, item := range self.items {
		sortedItems = append(sortedItems, item)
	}
	sort.Slice(sortedItems, func(i, j int) bool {
		return sortedItems[i].Money() >= sortedItems[j].Money()
	})
	return sortedItems
}

func (self *Section) addItem(sectKey string, money float32) *SectionItem {
	item := self.items[sectKey]
	if item == nil {
		item = newSectionItem(sectKey).withMonthsBetween(self.monthsBetween).
			withSkipFromSum(self.skipFromSum)
		self.items[sectKey] = item
	}
	item.Add(money)
	return item
}

func (self *Section) PerMonthString(format string) string {
	if self.skipFromSum || self.count < 2 || self.MonthsBetween() < 2 {
		return ""
	}
	return fmt.Sprintf(format, self.Money()/float32(self.MonthsBetween()))
}

// --------------------------------------------------

func newSectionItem(name string) *SectionItem {
	return &SectionItem{
		name: name,
	}
}

type SectionItem struct {
	name  string
	money float32
	count int

	monthsBetween func() int
	skipFromSum   bool
}

func (self *SectionItem) withMonthsBetween(m func() int) *SectionItem {
	self.monthsBetween = m
	return self
}

func (self *SectionItem) withSkipFromSum(v bool) *SectionItem {
	self.skipFromSum = v
	return self
}

func (self *SectionItem) Name() string {
	return self.name
}

func (self *SectionItem) Money() float32 {
	return self.money
}

func (self *SectionItem) Count() int {
	return self.count
}

func (self *SectionItem) Add(money float32) {
	self.count++
	self.money += money
}

func (self *SectionItem) MonthsBetween() int {
	return self.monthsBetween()
}

func (self *SectionItem) PerMonthString(format string) string {
	if self.skipFromSum || self.count < 2 || self.MonthsBetween() < 2 {
		return ""
	}
	return fmt.Sprintf(format, self.Money()/float32(self.MonthsBetween()))
}
