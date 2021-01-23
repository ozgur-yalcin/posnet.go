[![license](https://img.shields.io/:license-mit-blue.svg)](https://github.com/ozgur-soft/posnet/blob/master/LICENSE.md)
[![documentation](https://pkg.go.dev/badge/github.com/ozgur-soft/posnet)](https://pkg.go.dev/github.com/ozgur-soft/posnet/src)

# posnet
Posnet (Yapı Kredi) Sanal POS API with golang

# Installation
```bash
go get github.com/ozgur-soft/posnet
```

# Sanalpos satış işlemi
```go
package main

import (
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-soft/posnet/src"
)

func main() {
	api := &posnet.API{"yapikreditest"} // "yapikredi","yapikreditest"
	request := new(posnet.Request)
	request.MerchantID = "6706598320"
	request.TerminalID = "67005551"
	request.TranDate = "1"
	request.Sale = new(posnet.Sale)
	request.Sale.OrderID = ""                    // Sipariş numarası
	request.Sale.Amount = "2451"                 // Satış tutarı (1,00 -> 100) Son 2 hane kuruş
	request.Sale.CurrencyCode = "TL"             // TL, US, EU
	request.Sale.CardNumber = "4506349116608409" // Kart numarası
	request.Sale.ExpireDate = "0703"             // Son kullanma tarihi (Yıl ve ayın son 2 hanesi) YYAA
	request.Sale.CVV2 = "000"                    // Cvv2 Kodu (kartın arka yüzündeki 3 haneli numara)
	request.Sale.Installment = "00"              // peşin: "00", 2 taksit: "02"
	response := api.Transaction(request)
	pretty, _ := xml.MarshalIndent(response, " ", " ")
	fmt.Println(string(pretty))
}
```

# Sanalpos 3D satış işlemi
```go
package main

import (
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-soft/posnet/src"
)

func main() {
	api := &posnet.API{"yapikreditest"} // "yapikredi","yapikreditest"
	request := new(posnet.Request)
	request.MerchantID = "6706022701" // Üye işyeri numarası
	request.TerminalID = "67002706"   // Terminal numarası
	request.OOS = new(posnet.OOS)
	request.OOS.PosnetID = "142"                // POSNET numarası
	request.OOS.XID = ""                        // Sipariş numarası
	request.OOS.TranType = "Sale"               // İşlem tipi ("Sale","Auth")
	request.OOS.Amount = "5696"                 // Satış tutarı (1,00 -> 100) Son 2 hane kuruş
	request.OOS.CurrencyCode = "TL"             // TL, US, EU
	request.OOS.CardHolder = ""                 // Kart sahibi
	request.OOS.CardNumber = "5400637500005263" // Kart numarası
	request.OOS.ExpireDate = "0607"             // Son kullanma tarihi (Yıl ve ayın son 2 hanesi) YYAA
	request.OOS.CVV2 = "111"                    // Cvv2 Kodu (kartın arka yüzündeki 3 haneli numara)
	request.OOS.Installment = "00"              // peşin: "00", 2 taksit: "02"
	response := api.Transaction(request)
	pretty, _ := xml.MarshalIndent(response, " ", " ")
	fmt.Println(string(pretty))
}
```