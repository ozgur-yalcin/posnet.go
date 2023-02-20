[![license](https://img.shields.io/:license-mit-blue.svg)](https://github.com/ozgur-yalcin/posnet.go/blob/main/LICENSE.md)
[![documentation](https://pkg.go.dev/badge/github.com/ozgur-yalcin/posnet.go)](https://pkg.go.dev/github.com/ozgur-yalcin/posnet.go/src)

# Posnet.go
Yapı Kredi (Posnet) POS API with golang

# Installation
```bash
go get github.com/ozgur-yalcin/posnet.go
```

# Satış
```go
package main

import (
	"context"
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-yalcin/posnet.go/src"
)

// Üye işyeri bilgileri
const (
	envmode  = "TEST"       // Çalışma ortamı (Production : "PROD" - Test : "TEST")
	merchant = "6706598320" // İşyeri numarası
	terminal = "67005551"   // Terminal numarası
)

func main() {
	api, req := posnet.Api(merchant, terminal)
	api.Mode = envmode
	req.TranDate = "1"
	req.Auth = new(posnet.Auth)
	req.Auth.OrderId = posnet.XID(20)
	req.Auth.Amount = posnet.Amount("1.00")         // Satış tutarı (zorunlu)
	req.Auth.Installment = posnet.Installment("0")  // Taksit sayısı (peşin: "0") (zorunlu)
	req.Auth.Currency = posnet.Currency("TRY")      // Para birimi (zorunlu)
	req.Auth.CardNumber = ""                        // Kart numarası (zorunlu)
	req.Auth.CardExpiry = posnet.Expiry("02", "20") // Son kullanma tarihi - AA,YY (zorunlu)
	req.Auth.CardCode = "123"                       // Kart arkasındaki 3 haneli numara (zorunlu)

	ctx := context.Background()
	if res, err := api.Transaction(ctx, req); err == nil {
		pretty, _ := xml.MarshalIndent(res, " ", " ")
		fmt.Println(string(pretty))
	} else {
		fmt.Println(err)
	}
}
```

# İade
```go
package main

import (
	"context"
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-yalcin/posnet.go/src"
)

// Üye işyeri bilgileri
const (
	envmode  = "TEST"       // Çalışma ortamı (Production : "PROD" - Test : "TEST")
	merchant = "6706598320" // İşyeri numarası
	terminal = "67005551"   // Terminal numarası
)

func main() {
	api, req := posnet.Api(merchant, terminal)
	api.Mode = envmode
	req.Refund = new(posnet.Refund)
	req.Refund.Transaction = "sale"              // "sale" : Satış , "auth" : Provizyon
	req.Refund.HostLogKey = ""                   // İşlem numarası
	req.Refund.Amount = posnet.Amount("1.00")    // İade tutarı
	req.Refund.Currency = posnet.Currency("TRY") // Para birimi (zorunlu)

	ctx := context.Background()
	if res, err := api.Transaction(ctx, req); err == nil {
		pretty, _ := xml.MarshalIndent(res, " ", " ")
		fmt.Println(string(pretty))
	} else {
		fmt.Println(err)
	}
}
```

# İptal
```go
package main

import (
	"context"
	"encoding/xml"
	"fmt"

	posnet "github.com/ozgur-yalcin/posnet.go/src"
)

// Üye işyeri bilgileri
const (
	envmode  = "TEST"       // Çalışma ortamı (Production : "PROD" - Test : "TEST")
	merchant = "6706598320" // İşyeri numarası
	terminal = "67005551"   // Terminal numarası
)

func main() {
	api, req := posnet.Api(merchant, terminal)
	api.Mode = envmode
	req.Cancel = new(posnet.Cancel)
	req.Cancel.Transaction = "sale" // "sale" : Satış , "auth" : Provizyon
	req.Cancel.HostLogKey = ""      // İşlem numarası

	ctx := context.Background()
	if res, err := api.Transaction(ctx, req); err == nil {
		pretty, _ := xml.MarshalIndent(res, " ", " ")
		fmt.Println(string(pretty))
	} else {
		fmt.Println(err)
	}
}
```
