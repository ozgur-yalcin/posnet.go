package posnet

import (
	"encoding/xml"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var EndPoints map[string]string = map[string]string{
	"yapikredi":     "https://posnet.yapikredi.com.tr/PosnetWebService/XML",
	"yapikreditest": "https://setmpos.ykb.com/PosnetWebService/XML",
}

var Currencies map[string]string = map[string]string{
	"TRY": "949",
	"YTL": "949",
	"TRL": "949",
	"TL":  "949",
	"USD": "840",
	"EUR": "978",
	"GBP": "826",
	"JPY": "392",
}

type API struct {
	Bank string
}

type Request struct {
	XMLName    xml.Name `xml:"posnetRequest,omitempty"`
	TranDate   string   `xml:"tranDateRequired,omitempty"`
	MerchantID string   `xml:"mid,omitempty"`
	TerminalID string   `xml:"tid,omitempty"`
	OOS        struct {
		PosnetID     string `xml:"posnetid,omitempty"`
		XID          string `xml:"XID,omitempty"`
		Amount       string `xml:"amount,omitempty"`
		CurrencyCode string `xml:"currencyCode,omitempty"`
		Installment  string `xml:"installment,omitempty"`
		TranType     string `xml:"tranType,omitempty"`
		CardHolder   string `xml:"cardHolderName,omitempty"`
		CardNumber   string `xml:"ccno,omitempty"`
		ExpireDate   string `xml:"expDate,omitempty"`
		CVV2         string `xml:"cvc,omitempty"`
	} `xml:"oosRequestData,omitempty"`
	Auth struct {
		Amount       string `xml:"amount,omitempty"`
		CurrencyCode string `xml:"currencyCode,omitempty"`
		Card         struct {
			InquiryValue string `xml:"inquiryValue,omitempty"`
			CardNoFirst  string `xml:"cardNoFirst,omitempty"`
			CardNoLast   string `xml:"cardNoLast,omitempty"`
		} `xml:"cardInfo,omitempty"`
		CVV2        string `xml:"cvc,omitempty"`
		OrderID     string `xml:"orderID,omitempty"`
		Installment string `xml:"installment,omitempty"`
	} `xml:"auth,omitempty"`
	Sale struct {
		Amount       string `xml:"amount,omitempty"`
		CurrencyCode string `xml:"currencyCode,omitempty"`
		CardNumber   string `xml:"ccno,omitempty"`
		ExpireDate   string `xml:"expDate,omitempty"`
		CVV2         string `xml:"cvc,omitempty"`
		OrderID      string `xml:"orderID,omitempty"`
		Installment  string `xml:"installment,omitempty"`
	} `xml:"sale,omitempty"`
	Capt struct {
		Amount       string `xml:"amount,omitempty"`
		CurrencyCode string `xml:"currencyCode,omitempty"`
		HostLogKey   string `xml:"hostlogkey,omitempty"`
		Installment  string `xml:"installment,omitempty"`
	} `xml:"capt,omitempty"`
	Return struct {
		Amount       string `xml:"amount,omitempty"`
		CurrencyCode string `xml:"currencyCode,omitempty"`
		Transaction  string `xml:"transaction,omitempty"`
		HostLogKey   string `xml:"hostlogkey,omitempty"`
	} `xml:"return,omitempty"`
	Reverse struct {
		Transaction string `xml:"transaction,omitempty"`
		HostLogKey  string `xml:"hostlogkey,omitempty"`
	} `xml:"reverse,omitempty"`
}

type Response struct {
	XMLName    xml.Name `xml:"posnetResponse,omitempty"`
	Approved   string   `xml:"approved,omitempty"`
	HostLogKey string   `xml:"hostlogkey,omitempty"`
	AuthCode   string   `xml:"authCode,omitempty"`
	RespCode   string   `xml:"respCode,omitempty"`
	RespText   string   `xml:"respText,omitempty"`
	TranDate   string   `xml:"tranDate,omitempty"`
	YourIP     string   `xml:"yourIP,omitempty"`
	OOS        struct {
		Data1 string `xml:"data1,omitempty"`
		Data2 string `xml:"data2,omitempty"`
		Sign  string `xml:"sign,omitempty"`
	} `xml:"oosRequestDataResponse,omitempty"`
}

func (api *API) Transaction(request Request) (response Response) {
	postdata, _ := xml.Marshal(request)
	cli := new(http.Client)
	req, err := http.NewRequest("POST", EndPoints[api.Bank], strings.NewReader(xml.Header+string(postdata)))
	if err != nil {
		log.Println(err)
		return response
	}
	req.Header.Set("X-MERCHANT-ID", request.MerchantID)
	req.Header.Set("X-TERMINAL-ID", request.TerminalID)
	req.Header.Set("X-POSNET-ID", request.OOS.PosnetID)
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
