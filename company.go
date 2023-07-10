package downloadstock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Company
type Company struct {
	StockCode              int
	Securities             string
	Category               string
	SubCategory            string
	BoardLot               int
	ParValue               string
	ISIN                   string
	ExpiryDate             string // change to date later
	SubjectToStampDuty     bool
	ShortSellEligible      bool
	CASEligible            bool
	VCMEligible            bool
	AdmittedToStockOptions bool
	AdmittedToStockFutures bool
	AdmittedToCCASS        bool
	ETFOrFundManager       string
	DebtSecuritiesBoardLot string
	DebtSecuritiesInvestor string
	POSEligble             bool
	SpreadTable            int
}

// GetCompanyList gets a list of stocks
func GetCompanyList() ([]Company, error) {

	var stocks []Company

	// Read bytes from url
	url := "https://www.hkex.com.hk/eng/services/trading/securities/securitieslists/ListOfSecurities.xlsx"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return stocks, err
	}

	// Open bytes with excel reader
	f, err := excelize.OpenReader(bytes.NewReader(b))
	if err != nil {
		return stocks, err
	}
	defer func() {
		if err = f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Get rows from excel sheet ListOfSecurities
	rows, err := f.GetRows("ListOfSecurities")
	if len(rows) > 3 { // Remove the first 3 rows, which are title, date, and headers
		rows = rows[3:]
	} else {
		return stocks, err
	}

	fmt.Println(PrettyPrint(rows))

	for i, r := range rows {

		// parse string to correct type
		stockCode, _ := strconv.Atoi(r[0])
		boardLot, _ := strconv.Atoi(strings.ReplaceAll(r[4], ",", ""))
		spreadTable, _ := strconv.Atoi(strings.TrimSpace(r[19]))

		if len(r) == 20 {
			s := Company{
				StockCode:              stockCode,               // r[0] - "Stock Code"
				Securities:             r[1],                    // r[1] - "Name of Securities"
				Category:               r[2],                    // r[2] - "Category"
				SubCategory:            r[3],                    // r[3] - "Sub-Category"
				BoardLot:               boardLot,                // r[4] - "Board Lot"
				ParValue:               strings.TrimSpace(r[5]), // r[5] - "Par Value"
				ISIN:                   r[6],                    // r[6] - "ISIN"
				ExpiryDate:             r[7],                    // r[7] - "Expiry Date"
				SubjectToStampDuty:     convertBool(r[8]),       // r[8] - "Subject to Stamp Duty"
				ShortSellEligible:      convertBool(r[9]),       // r[9] - "Shortsell Eligible"
				CASEligible:            convertBool(r[10]),      // r[10] - "CAS Eligible"
				VCMEligible:            convertBool(r[11]),      // r[11] - "VCM Eligible"
				AdmittedToStockOptions: convertBool(r[12]),      // r[12] - "Admitted to Stock Options"
				AdmittedToStockFutures: convertBool(r[13]),      // r[13]- "Admitted to Stock Futures"
				AdmittedToCCASS:        convertBool(r[14]),      // r[14] - "Admitted to CCASS"
				ETFOrFundManager:       r[15],                   // r[15] - "ETF / Fund Manager"
				DebtSecuritiesBoardLot: r[16],                   // r[16] - "Debt Securities Board Lot (Nominal)"
				DebtSecuritiesInvestor: r[17],                   // r[17] - "Debt Securities Investor Type"
				POSEligble:             convertBool(r[18]),      // r[18] - "POS Eligble"
				SpreadTable:            spreadTable,             // r[19] - "Spread Table
			}

			// return equity only
			if s.Category == "Equity" {
				stocks = append(stocks, s)
			}

		} else {
			msg := fmt.Sprintf("Something is wrong with row: %d", i)
			fmt.Println(msg)
		}

	}

	return stocks, nil
}

func convertBool(s string) bool {
	if s == "Y" {
		return true
	}
	return false
}
