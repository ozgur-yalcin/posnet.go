package main

import (
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-soft/posnet/src"
)

// Üye işyeri bilgileri
const (
	environment = "TEST"       // Çalışma ortamı "PROD", "TEST"
	merchantID  = "6706598320" // Üye işyeri numarası
	terminalID  = "67005551"   // Terminal numarası
)

// DİREKT SATIŞ (3D'siz)
func main() {
	api := &posnet.API{environment}
	request := new(posnet.Request)
	request.MerchantID = merchantID
	request.TerminalID = terminalID
	request.TranDate = "1"
	request.Sale = new(posnet.Sale)
	request.Sale.OrderID = posnet.XID(20)        // Sipariş numarası
	request.Sale.Amount = "100"                  // Satış tutarı (1,00 -> 100) Son 2 hane kuruş
	request.Sale.CurrencyCode = "TL"             // Para birimi (TL, US, EU)
	request.Sale.CardNumber = "4506349116608409" // Kart numarası
	request.Sale.ExpireDate = "0703"             // Son kullanma tarihi (Yıl ve ayın son 2 hanesi) YYAA
	request.Sale.CVV2 = "000"                    // Cvv2 Kodu (kartın arka yüzündeki 3 haneli numara)
	request.Sale.Installment = "00"              // peşin: "00", 2 taksit: "02"
	response := api.Transaction(request)
	pretty, _ := xml.MarshalIndent(response, " ", " ")
	fmt.Println(string(pretty))
}
