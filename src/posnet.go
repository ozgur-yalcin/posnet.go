package posnet

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var EndPoints map[string]string = map[string]string{
	"yapikredi":     "https://posnet.yapikredi.com.tr/PosnetWebService/XML",
	"yapikreditest": "https://setmpos.ykb.com/PosnetWebService/XML",
}

type API struct {
	Bank string
}

type Request struct {
	XMLName    xml.Name    `xml:"posnetRequest,omitempty"`
	TranDate   interface{} `xml:"tranDateRequired,omitempty"`
	MerchantID interface{} `xml:"mid,omitempty"`
	TerminalID interface{} `xml:"tid,omitempty"`
	OOS        *OOS        `xml:"oosRequestData,omitempty"`
	Auth       *Auth       `xml:"auth,omitempty"`
	Sale       *Sale       `xml:"sale,omitempty"`
	Capt       *Capt       `xml:"capt,omitempty"`
	Return     *Return     `xml:"return,omitempty"`
	Reverse    *Reverse    `xml:"reverse,omitempty"`
}

type OOS struct {
	PosnetID     interface{} `xml:"posnetid,omitempty"`
	XID          interface{} `xml:"XID,omitempty"`
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	Installment  interface{} `xml:"installment,omitempty"`
	TranType     interface{} `xml:"tranType,omitempty"`
	CardHolder   interface{} `xml:"cardHolderName,omitempty"`
	CardNumber   interface{} `xml:"ccno,omitempty"`
	ExpireDate   interface{} `xml:"expDate,omitempty"`
	CVV2         interface{} `xml:"cvc,omitempty"`
}

type Auth struct {
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	Card         struct {
		InquiryValue interface{} `xml:"inquiryValue,omitempty"`
		CardNoFirst  interface{} `xml:"cardNoFirst,omitempty"`
		CardNoLast   interface{} `xml:"cardNoLast,omitempty"`
	} `xml:"cardInfo,omitempty"`
	CVV2        interface{} `xml:"cvc,omitempty"`
	OrderID     interface{} `xml:"orderID,omitempty"`
	Installment interface{} `xml:"installment,omitempty"`
}

type Sale struct {
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	CardNumber   interface{} `xml:"ccno,omitempty"`
	ExpireDate   interface{} `xml:"expDate,omitempty"`
	CVV2         interface{} `xml:"cvc,omitempty"`
	OrderID      interface{} `xml:"orderID,omitempty"`
	Installment  interface{} `xml:"installment,omitempty"`
}

type Capt struct {
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	HostLogKey   interface{} `xml:"hostlogkey,omitempty"`
	Installment  interface{} `xml:"installment,omitempty"`
}

type Return struct {
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	Transaction  interface{} `xml:"transaction,omitempty"`
	HostLogKey   interface{} `xml:"hostlogkey,omitempty"`
}

type Reverse struct {
	Transaction interface{} `xml:"transaction,omitempty"`
	HostLogKey  interface{} `xml:"hostlogkey,omitempty"`
}

type Response struct {
	XMLName    xml.Name    `xml:"posnetResponse,omitempty"`
	Approved   interface{} `xml:"approved,omitempty"`
	HostLogKey interface{} `xml:"hostlogkey,omitempty"`
	AuthCode   interface{} `xml:"authCode,omitempty"`
	RespCode   interface{} `xml:"respCode,omitempty"`
	RespText   interface{} `xml:"respText,omitempty"`
	TranDate   interface{} `xml:"tranDate,omitempty"`
	YourIP     interface{} `xml:"yourIP,omitempty"`
	OOS        *struct {
		Data1 interface{} `xml:"data1,omitempty"`
		Data2 interface{} `xml:"data2,omitempty"`
		Sign  interface{} `xml:"sign,omitempty"`
	} `xml:"oosRequestDataResponse,omitempty"`
}

func (api *API) Transaction(request *Request) (response Response) {
	postdata, _ := xml.MarshalIndent(request, " ", " ")
	cli := new(http.Client)
	fmt.Println(EndPoints[api.Bank])
	fmt.Println(xml.Header + string(postdata))
	req, err := http.NewRequest("POST", EndPoints[api.Bank], strings.NewReader(xml.Header+string(postdata)))
	if err != nil {
		log.Println(err)
		return response
	}
	req.Header.Set("X-MERCHANT-ID", fmt.Sprintf("%v", request.MerchantID))
	req.Header.Set("X-TERMINAL-ID", fmt.Sprintf("%v", request.TerminalID))
	if request.OOS != nil {
		req.Header.Set("X-POSNET-ID", fmt.Sprintf("%v", request.OOS.PosnetID))
	}
	req.Header.Set("X-CORRELATION-ID", uuid.New().String())
	res, err := cli.Do(req)
	if err != nil {
		log.Println(err)
		return response
	}
	defer res.Body.Close()
	decoder := xml.NewDecoder(res.Body)
	decoder.Decode(&response)
	return response
}
