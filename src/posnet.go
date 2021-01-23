package posnet

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html/charset"
)

var EndPoints map[string]string = map[string]string{
	"yapikredi":     "https://posnet.yapikredi.com.tr/PosnetWebService/XML",
	"yapikreditest": "https://setmpos.ykb.com/PosnetWebService/XML",
}

type API struct {
	Bank string
}

type Request struct {
	XMLName     xml.Name     `xml:"posnetRequest,omitempty"`
	MerchantID  interface{}  `xml:"mid,omitempty"`
	TerminalID  interface{}  `xml:"tid,omitempty"`
	TranDate    interface{}  `xml:"tranDateRequired,omitempty"`
	OOS         *OOS         `xml:"oosRequestData,omitempty"`
	OOSMerchant *OOSMerchant `xml:"oosResolveMerchantData,omitempty"`
	OOSTran     *OOSTran     `xml:"oosTranData,omitempty"`
	Auth        *Auth        `xml:"auth,omitempty"`
	Sale        *Sale        `xml:"sale,omitempty"`
	Capt        *Capt        `xml:"capt,omitempty"`
	Return      *Return      `xml:"return,omitempty"`
	Reverse     *Reverse     `xml:"reverse,omitempty"`
}

type OOS struct {
	PosnetID     interface{} `xml:"posnetid,omitempty"`
	XID          interface{} `xml:"XID,omitempty"`
	TranType     interface{} `xml:"tranType,omitempty"`
	CardHolder   interface{} `xml:"cardHolderName,omitempty"`
	CardNumber   interface{} `xml:"ccno,omitempty"`
	ExpireDate   interface{} `xml:"expDate,omitempty"`
	CVV2         interface{} `xml:"cvc,omitempty"`
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	Installment  interface{} `xml:"installment,omitempty"`
}

type OOSMerchant struct {
	BankData     interface{} `xml:"bankData,omitempty"`
	MerchantData interface{} `xml:"merchantData,omitempty"`
	SIGN         interface{} `xml:"sign,omitempty"`
	MAC          interface{} `xml:"mac,omitempty"`
}

type OOSTran struct {
	BankData interface{} `xml:"bankData,omitempty"`
	WpAmount interface{} `xml:"wpAmount,omitempty"`
	MAC      interface{} `xml:"mac,omitempty"`
}

type Auth struct {
	Card         *Card       `xml:"cardInfo,omitempty"`
	CVV2         interface{} `xml:"cvc,omitempty"`
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	Installment  interface{} `xml:"installment,omitempty"`
	OrderID      interface{} `xml:"orderID,omitempty"`
}

type Card struct {
	InquiryValue interface{} `xml:"inquiryValue,omitempty"`
	CardNoFirst  interface{} `xml:"cardNoFirst,omitempty"`
	CardNoLast   interface{} `xml:"cardNoLast,omitempty"`
}

type Sale struct {
	CardNumber   interface{} `xml:"ccno,omitempty"`
	ExpireDate   interface{} `xml:"expDate,omitempty"`
	CVV2         interface{} `xml:"cvc,omitempty"`
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	Installment  interface{} `xml:"installment,omitempty"`
	OrderID      interface{} `xml:"orderID,omitempty"`
}

type Capt struct {
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	Installment  interface{} `xml:"installment,omitempty"`
	HostLogKey   interface{} `xml:"hostlogkey,omitempty"`
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
	XMLName    xml.Name `xml:"posnetResponse,omitempty"`
	Approved   string   `xml:"approved,omitempty"`
	HostLogKey string   `xml:"hostlogkey,omitempty"`
	AuthCode   string   `xml:"authCode,omitempty"`
	ErrorCode  string   `xml:"respCode,omitempty"`
	ErrorText  string   `xml:"respText,omitempty"`
	TranDate   string   `xml:"tranDate,omitempty"`
	YourIP     string   `xml:"yourIP,omitempty"`
	MAC        string   `xml:"mac,omitempty"`
	OOS        struct {
		Data1 string `xml:"data1,omitempty"`
		Data2 string `xml:"data2,omitempty"`
		Sign  string `xml:"sign,omitempty"`
	} `xml:"oosRequestDataResponse,omitempty"`
	OOSMerchant struct {
		XID         string `xml:"XID,omitempty"`
		Amount      string `xml:"amount,omitempty"`
		Currency    string `xml:"currency,omitempty"`
		Installment string `xml:"installment,omitempty"`
		TxStatus    string `xml:"txStatus,omitempty"`
		MdStatus    string `xml:"mdStatus,omitempty"`
		MdError     string `xml:"mdErrorMessage,omitempty"`
		MAC         string `xml:"mac,omitempty"`
	} `xml:"oosResolveMerchantDataResponse,omitempty"`
}

func SHA256(data string) (hash string) {
	h := sha256.New()
	h.Write([]byte(data))
	hash = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return hash
}

func MAC(xid, amount, currency, mid, key, tid string) string {
	return SHA256(xid + ";" + amount + ";" + currency + ";" + mid + ";" + SHA256(key+";"+tid))
}

func (api *API) Transaction(request *Request) (response Response) {
	xmldata, _ := xml.Marshal(request)
	urldata := url.Values{}
	urldata.Set("xmldata", string(xmldata))
	req, err := http.NewRequest("POST", EndPoints[api.Bank], strings.NewReader(urldata.Encode()))
	if err != nil {
		log.Println(err)
		return response
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-MERCHANT-ID", request.MerchantID.(string))
	req.Header.Set("X-TERMINAL-ID", request.TerminalID.(string))
	if request.Sale != nil {
		req.Header.Set("X-CORRELATION-ID", request.Sale.OrderID.(string))
	}
	if request.OOS != nil {
		req.Header.Set("X-POSNET-ID", request.OOS.PosnetID.(string))
	}
	cli := new(http.Client)
	res, err := cli.Do(req)
	if err != nil {
		log.Println(err)
		return response
	}
	defer res.Body.Close()
	decoder := xml.NewDecoder(res.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	decoder.Decode(&response)
	return response
}
