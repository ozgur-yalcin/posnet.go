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

// Üye işyeri bilgileri
const (
	merchantID = "6706598320" // Üye işyeri numarası
	terminalID = "67005551"   // Terminal numarası
)

// DİREKT SATIŞ (3D'siz)
func main() {
	api := &posnet.API{"TEST"} // "PROD","TEST"
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
```

# Sanalpos 3D secure satış işlemi
```go
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	posnet "github.com/ozgur-soft/posnet/src"
)

// Üye işyeri bilgileri
const (
	merchantID = "6706598320"              // Üye işyeri numarası
	terminalID = "67005551"                // Terminal numarası
	posnetID   = "9644"                    // POSNET numarası
	secretKey  = "10,10,10,10,10,10,10,10" // Güvenlik anahtarı
	currency   = "TL"                      // TL, US, EU
	language   = "tr"                      // Dil
)

// Test sunucu bilgileri
const (
	httpHost = "localhost"
	httpPort = ":8080"
)

func main() {
	http.HandleFunc("/", OOSHandler)
	server := http.Server{Addr: httpHost + httpPort, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second}
	e := server.ListenAndServe()
	if e != nil {
		fmt.Println(e)
	}
}

// 3d secure - Ödeme test sayfası
func OOSHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		switch r.Method {
		case "GET":
			data := make(map[string]interface{})
			buffer := new(bytes.Buffer)
			err := template.Must(template.New("3d_payment.html").ParseGlob("*.html")).Execute(buffer, data)
			if err != nil {
				fmt.Println(err)
			}
			buffer.WriteTo(w)
		case "POST":
			cardowner := r.FormValue("cardowner")
			cardnumber := r.FormValue("cardnumber")
			cardmonth := r.FormValue("cardmonth")
			cardyear := r.FormValue("cardyear")
			cardcvc := r.FormValue("cardcvc")
			amount := r.FormValue("amount")
			decimal := r.FormValue("decimal")
			installment := r.FormValue("installment")
			res := OOS(cardowner, cardnumber, cardmonth, cardyear, cardcvc, fmt.Sprintf("%v", amount)+fmt.Sprintf("%02v", decimal), installment)
			if res.Approved == "1" {
				data := make(map[string]interface{})
				data["url"] = posnet.EndPoints["test3d"] // "prod3d","test3d"
				data["host"] = httpHost
				data["port"] = httpPort
				data["lang"] = language
				data["mid"] = merchantID
				data["pid"] = posnetID
				data["data1"] = res.OOS.Data1
				data["data2"] = res.OOS.Data2
				data["sign"] = res.OOS.Sign
				buffer := new(bytes.Buffer)
				err := template.Must(template.New("3d_post.html").ParseGlob("*.html")).Execute(buffer, data)
				if err != nil {
					fmt.Println(err)
				}
				buffer.WriteTo(w)
			} else {
				data := make(map[string]interface{})
				data["code"] = res.ErrorCode
				data["text"] = res.ErrorText
				buffer := new(bytes.Buffer)
				err := template.Must(template.New("3d_error.html").ParseGlob("*.html")).Execute(buffer, data)
				if err != nil {
					fmt.Println(err)
				}
				buffer.WriteTo(w)
			}
		}
	case "/payment":
		switch r.Method {
		case "POST":
			mdata := r.FormValue("MerchantPacket")
			bdata := r.FormValue("BankPacket")
			sign := r.FormValue("Sign")
			xid := r.FormValue("Xid")
			amount := strings.ReplaceAll(r.FormValue("Amount"), ",", "")
			OOSMerchant(xid, amount, currency, mdata, bdata, sign)
			OOSTransaction(xid, amount, currency, bdata)
			http.Redirect(w, r, "//"+httpHost+httpPort, http.StatusMovedPermanently)
		}
	}
}

// 3d secure - Verilerin şifrelenmesi 1. adım
func OOS(cardowner, cardnumber, cardmonth, cardyear, cardcvc, amount, installment string) (response posnet.Response) {
	api := &posnet.API{"TEST"} // "PROD","TEST"
	request := new(posnet.Request)
	request.MerchantID = merchantID
	request.TerminalID = terminalID
	request.OOS = new(posnet.OOS)
	request.OOS.PosnetID = posnetID
	request.OOS.XID = posnet.XID(20) // Sipariş numarası
	request.OOS.TranType = "Sale"    // İşlem tipi ("Sale","Auth")
	request.OOS.Amount = amount
	request.OOS.CurrencyCode = currency
	request.OOS.CardHolder = cardowner
	request.OOS.CardNumber = cardnumber
	request.OOS.ExpireDate = fmt.Sprintf("%02v", cardyear) + fmt.Sprintf("%02v", cardmonth)
	request.OOS.CVV2 = fmt.Sprintf("%03v", cardcvc)
	request.OOS.Installment = fmt.Sprintf("%02v", installment)
	response = api.Transaction(request)
	pretty, _ := xml.MarshalIndent(response, " ", " ")
	fmt.Println(string(pretty))
	return response
}

// 3d secure - Kullanıcı Doğrulama (2. adım)
func OOSMerchant(xid, amount, currency, mdata, bdata, sign string) (response posnet.Response) {
	api := &posnet.API{"TEST"} // "PROD","TEST"
	request := new(posnet.Request)
	request.MerchantID = merchantID
	request.TerminalID = terminalID
	request.OOSMerchant = new(posnet.OOSMerchant)
	request.OOSMerchant.MerchantData = mdata
	request.OOSMerchant.BankData = bdata
	request.OOSMerchant.SIGN = sign
	request.OOSMerchant.MAC = posnet.MAC(xid, amount, currency, merchantID, secretKey, terminalID, "")
	response = api.Transaction(request)
	pretty, _ := xml.MarshalIndent(response, " ", " ")
	fmt.Println(string(pretty))
	check := posnet.MAC(xid, amount, currency, merchantID, secretKey, terminalID, response.OOSMerchant.MdStatus)
	if check == response.OOSMerchant.MAC {
		return response
	}
	return posnet.Response{}
}

// 3d secure - Finansallaştırma (3. adım)
func OOSTransaction(xid, amount, currency, bdata string) (response posnet.Response) {
	api := &posnet.API{"TEST"} // "PROD","TEST"
	request := new(posnet.Request)
	request.MerchantID = merchantID
	request.TerminalID = terminalID
	request.OOSTran = new(posnet.OOSTran)
	request.OOSTran.BankData = bdata
	request.OOSTran.MAC = posnet.MAC(xid, amount, currency, merchantID, secretKey, terminalID, "")
	request.OOSTran.WpAmount = "0"
	response = api.Transaction(request)
	pretty, _ := xml.MarshalIndent(response, " ", " ")
	fmt.Println(string(pretty))
	check := posnet.MAC(xid, amount, currency, merchantID, secretKey, terminalID, response.HostLogKey)
	if check == response.MAC {
		return response
	}
	return posnet.Response{}
}
```
