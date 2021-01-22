package main

import (
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-soft/posnet/src"
)

func main() {
	api := &posnet.API{"yapikreditest"} // "yapikredi"
	request := new(posnet.Request)
	request.MerchantID = "6706598320"
	request.TerminalID = "67005551"
	request.TranDate = "1"
	request.Sale = new(posnet.Sale)
	request.Sale.Amount = "2451"                     // Satış tutarı (1,00 -> 100) Son 2 hane kuruş
	request.Sale.CurrencyCode = "TL"                 // TL, US, EU
	request.Sale.CardNumber = "4506349116608409"     // Kart numarası
	request.Sale.ExpireDate = "0703"                 // Son kullanma tarihi (Yıl ve ayın son 2 hanesi) YYAA
	request.Sale.CVV2 = "000"                        // Cvv2 Kodu (kartın arka yüzündeki 3 haneli numara)
	request.Sale.Installment = "00"                  // peşin: "00", 2 taksit: "02"
	request.Sale.OrderID = "1s3456z8901234567890123" // Sipariş numarası
	response := api.Transaction(request)
	pretty, _ := xml.MarshalIndent(response, " ", " ")
	fmt.Println(string(pretty))
}
