[![license](https://img.shields.io/:license-mit-blue.svg)](https://github.com/ozgur-soft/posnet.go/blob/main/LICENSE.md)
[![documentation](https://pkg.go.dev/badge/github.com/ozgur-soft/posnet.go)](https://pkg.go.dev/github.com/ozgur-soft/posnet.go/src)

# posnet.go
Posnet (Yapı Kredi) Sanal POS API with golang

# Installation
```bash
go get github.com/ozgur-soft/posnet.go
```

# Sanalpos direkt satış işlemi (3D'siz)
```go
package main

import (
	"context"
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-soft/posnet.go/src"
)

// Üye işyeri bilgileri
const (
	environment = "TEST"       // Çalışma ortamı "PROD", "TEST"
	merchantID  = "6706598320" // Üye işyeri numarası
	terminalID  = "67005551"   // Terminal numarası
)

func main() {
	api, req := posnet.Api(merchantID, terminalID)
	api.Mode = environment
	req.Sale = new(posnet.Sale)
	req.Sale.OrderID = posnet.XID(20)        // Sipariş numarası
	req.Sale.Amount = "100"                  // Satış tutarı (1,00 -> 100) Son 2 hane kuruş
	req.Sale.CurrencyCode = "TL"             // Para birimi (TL, US, EU)
	req.Sale.CardNumber = "4506349116608409" // Kart numarası
	req.Sale.CardExpiry = "0703"             // Son kullanma tarihi (Yıl ve ayın son 2 hanesi) YYAA
	req.Sale.CardCode = "000"                // Cvv2 Kodu (kartın arka yüzündeki 3 haneli numara)
	req.Sale.Installment = "00"              // peşin: "00", 2 taksit: "02"
	req.TranDate = "1"

	ctx := context.Background()
	res := api.Transaction(ctx, req)
	pretty, _ := xml.MarshalIndent(res, " ", " ")
	fmt.Println(string(pretty))
}
```

# Sanalpos 3D secure satış işlemi
```go
package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	posnet "github.com/ozgur-soft/posnet.go/src"
)

// Sunucu bilgileri
const (
	httpHost = "localhost" // "localhost" , "alanadiniz.com"
	httpPort = ":8080"     // ssl için :443 veya :https kullanılmalıdır

	returnurl = "http://localhost:8080/"
)

// Üye işyeri bilgileri
const (
	environment = "TEST"                    // Çalışma ortamı "PROD", "TEST"
	merchantID  = "6706598320"              // Üye işyeri numarası
	terminalID  = "67005551"                // Terminal numarası
	posnetID    = "9644"                    // POSNET numarası
	secretKey   = "10,10,10,10,10,10,10,10" // Güvenlik anahtarı
	currency    = "TL"                      // Para birimi (TL, US, EU)
	language    = "tr"                      // Dil
)

func main() {
	http.HandleFunc("/", OOSHandler)
	server := http.Server{Addr: httpHost + httpPort, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second}
	// ssl için server.ListenAndServeTLS(".cert dosyası", ".key dosyası") kullanılmalıdır.
	if e := server.ListenAndServe(); e != nil {
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
			err := template.Must(template.New("payment.html").ParseGlob("src/3D/*.html")).Execute(buffer, data)
			if err != nil {
				fmt.Println(err)
			}
			buffer.WriteTo(w)
		case "POST":
			cardholder := r.FormValue("cardholder")
			cardnumber := r.FormValue("cardnumber")
			cardmonth := r.FormValue("cardmonth")
			cardyear := r.FormValue("cardyear")
			cardcode := r.FormValue("cardcode")
			amount := r.FormValue("amount")
			decimal := r.FormValue("decimal")
			installment := r.FormValue("installment")
			res := OOS(cardholder, cardnumber, cardmonth, cardyear, cardcode, fmt.Sprintf("%v", amount)+fmt.Sprintf("%02v", decimal), installment)
			if res.Approved == "1" {
				data := make(map[string]interface{})
				data["endpoint"] = posnet.EndPoints[environment+"3d"]
				data["page"] = r.RequestURI
				data["return"] = returnurl
				data["lang"] = language
				data["mid"] = merchantID
				data["pid"] = posnetID
				data["data1"] = res.OOS.Data1
				data["data2"] = res.OOS.Data2
				data["sign"] = res.OOS.Sign
				buffer := new(bytes.Buffer)
				err := template.Must(template.New("post.html").ParseGlob("src/3D/*.html")).Execute(buffer, data)
				if err != nil {
					fmt.Println(err)
				}
				buffer.WriteTo(w)
			} else {
				data := make(map[string]interface{})
				data["code"] = res.ErrorCode
				data["text"] = res.ErrorText
				buffer := new(bytes.Buffer)
				err := template.Must(template.New("error.html").ParseGlob("src/3D/*.html")).Execute(buffer, data)
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
			http.Redirect(w, r, returnurl, http.StatusMovedPermanently)
		}
	}
}

// 3d secure - Verilerin şifrelenmesi 1. adım
func OOS(cardholder, cardnumber, cardmonth, cardyear, cardcode, amount, installment string) (response posnet.Response) {
	api, req := posnet.Api(merchantID, terminalID)
	api.Mode = environment
	req.OOS = new(posnet.OOS)
	req.OOS.PosnetID = posnetID
	req.OOS.XID = posnet.XID(20) // Sipariş numarası
	req.OOS.TranType = "Sale"    // İşlem tipi ("Sale","Auth")
	req.OOS.Amount = amount
	req.OOS.CurrencyCode = currency
	req.OOS.CardHolder = cardholder
	req.OOS.CardNumber = cardnumber
	req.OOS.CardExpiry = fmt.Sprintf("%02v", cardyear) + fmt.Sprintf("%02v", cardmonth)
	req.OOS.CardCode = fmt.Sprintf("%03v", cardcode)
	req.OOS.Installment = fmt.Sprintf("%02v", installment)

	ctx := context.Background()
	res := api.Transaction(ctx, req)
	pretty, _ := xml.MarshalIndent(res, " ", " ")
	fmt.Println(string(pretty))
	return res
}

// 3d secure - Kullanıcı Doğrulama (2. adım)
func OOSMerchant(xid, amount, currency, mdata, bdata, sign string) (response posnet.Response) {
	api, req := posnet.Api(merchantID, terminalID)
	api.Mode = environment
	req.OOSMerchant = new(posnet.OOSMerchant)
	req.OOSMerchant.MerchantData = mdata
	req.OOSMerchant.BankData = bdata
	req.OOSMerchant.SIGN = sign
	req.OOSMerchant.MAC = posnet.MAC(xid, amount, currency, merchantID, secretKey, terminalID, "")

	ctx := context.Background()
	res := api.Transaction(ctx, req)
	pretty, _ := xml.MarshalIndent(res, " ", " ")
	fmt.Println(string(pretty))

	check := posnet.MAC(xid, amount, currency, merchantID, secretKey, terminalID, response.OOSMerchant.MdStatus)
	if check == response.OOSMerchant.MAC {
		return response
	}
	return posnet.Response{}
}

// 3d secure - Finansallaştırma (3. adım)
func OOSTransaction(xid, amount, currency, bdata string) (response posnet.Response) {
	api, req := posnet.Api(merchantID, terminalID)
	api.Mode = environment
	req.OOSTran = new(posnet.OOSTran)
	req.OOSTran.BankData = bdata
	req.OOSTran.MAC = posnet.MAC(xid, amount, currency, merchantID, secretKey, terminalID, "")
	req.OOSTran.WpAmount = "0"

	ctx := context.Background()
	res := api.Transaction(ctx, req)
	pretty, _ := xml.MarshalIndent(res, " ", " ")
	fmt.Println(string(pretty))

	check := posnet.MAC(xid, amount, currency, merchantID, secretKey, terminalID, response.HostLogKey)
	if check == response.MAC {
		return response
	}
	return posnet.Response{}
}
```

# Sanalpos iade işlemi
```go
package main

import (
	"context"
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-soft/posnet.go/src"
)

// Üye işyeri bilgileri
const (
	environment = "TEST"       // Çalışma ortamı "PROD", "TEST"
	merchantID  = "6706598320" // Üye işyeri numarası
	terminalID  = "67005551"   // Terminal numarası
)

func main() {
	api, req := posnet.Api(merchantID, terminalID)
	api.Mode = environment
	req.Return = new(posnet.Return)
	req.Return.Transaction = "sale"
	req.Return.HostLogKey = ""     // İşlem numarası
	req.Return.Amount = "100"      // İade tutarı (1,00 -> 100) Son 2 hane kuruş
	req.Return.CurrencyCode = "TL" // Para birimi (TL, US, EU)

	ctx := context.Background()
	res := api.Transaction(ctx, req)
	pretty, _ := xml.MarshalIndent(res, " ", " ")
	fmt.Println(string(pretty))
}
```

# Sanalpos iptal işlemi
```go
package main

import (
	"context"
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-soft/posnet.go/src"
)

// Üye işyeri bilgileri
const (
	environment = "TEST"       // Çalışma ortamı "PROD", "TEST"
	merchantID  = "6706598320" // Üye işyeri numarası
	terminalID  = "67005551"   // Terminal numarası
)

func main() {
	api, req := posnet.Api(merchantID, terminalID)
	api.Mode = environment
	req.Reverse = new(posnet.Reverse)
	req.Reverse.Transaction = "sale"
	req.Reverse.HostLogKey = "" // İşlem numarası

	ctx := context.Background()
	res := api.Transaction(ctx, req)
	pretty, _ := xml.MarshalIndent(res, " ", " ")
	fmt.Println(string(pretty))
}
```