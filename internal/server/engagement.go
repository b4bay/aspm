package server

import (
	"github.com/b4bay/aspm/internal/server/sarif"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Engagement struct {
	gorm.Model
	ProductID string `gorm:"index;not null"`
	Tool      string
	RawReport string
	report    *sarif.Report
	// Associations
	Product Product `gorm:"constraint:OnDelete:CASCADE;foreignKey:ProductID;references:ProductID"`
}

func (e *Engagement) AfterFind(db *gorm.DB) (err error) {
	return e.parseRawReport()
}

func (e *Engagement) parseRawReport() (err error) {
	if e.RawReport != "" {
		if e.report, err = sarif.FromBase64(e.RawReport); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engagement) UpdateTool() (err error) {
	if e.report == nil {
		if e.report, err = sarif.FromBase64(e.RawReport); err != nil {
			return err
		}
	}

	e.Tool = e.report.Runs[0].Tool.Driver.Name
	return nil
}

func (e *Engagement) Process(tx *gorm.DB) (err error) {
	if err = e.parseRawReport(); err != nil {
		return err
	}

	for _, run := range e.report.Runs {
		for _, result := range run.Results {
			var v Vulnerability
			v.VulnerabilityID = result.RuleId
			v.LocationHash = result.LocationHash()
			v.ProductID = e.ProductID
			v.Level = result.Level
			v.Text = result.Message.Text
			v.CWE = result.CWE()
			v.CVE = result.CVE()
			v.EngagementID = e.ID

			if r := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&v); r.Error != nil {
				return err
			}

		}
	}

	return nil
}

func (e *Engagement) Report() *sarif.Report {
	return e.report
}
