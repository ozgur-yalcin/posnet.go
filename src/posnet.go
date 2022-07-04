package posnet

import (
	"context"
	"crypto/rand"
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
	"PROD":   "https://posnet.yapikredi.com.tr/PosnetWebService/XML",
	"PROD3d": "https://posnet.yapikredi.com.tr/3DSWebService/YKBPaymentService",
	"TEST":   "https://setmpos.ykb.com/PosnetWebService/XML",
	"TEST3d": "https://setmpos.ykb.com/3DSWebService/YKBPaymentService",
}

type API struct {
	Mode string
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
	CardExpiry   interface{} `xml:"expDate,omitempty"`
	CardCode     interface{} `xml:"cvc,omitempty"`
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
	CardCode     interface{} `xml:"cvc,omitempty"`
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
	CardExpiry   interface{} `xml:"expDate,omitempty"`
	CardCode     interface{} `xml:"cvc,omitempty"`
	Amount       interface{} `xml:"amount,omitempty"`
	CurrencyCode interface{} `xml:"currencyCode,omitempty"`
	Installment  interface{} `xml:"installment,omitempty"`
	OrderID      interface{} `xml:"orderID,omitempty"`
	Mailorder    interface{} `xml:"mailorderflag,omitempty"`
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

func MAC(xid, amount, currency, mid, key, tid, extra string) (mac string) {
	if extra != "" {
		mac = SHA256(extra + ";" + xid + ";" + amount + ";" + currency + ";" + mid + ";" + SHA256(key+";"+tid))
	} else {
		mac = SHA256(xid + ";" + amount + ";" + currency + ";" + mid + ";" + SHA256(key+";"+tid))
	}
	return mac
}

func XID(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func Api(merchantid, terminalid string) (*API, *Request) {
	api := new(API)
	request := new(Request)
	request.MerchantID = merchantid
	request.TerminalID = terminalid
	return api, request
}

func (api *API) Transaction(ctx context.Context, req *Request) (res Response) {
	xmldata, _ := xml.Marshal(req)
	urldata := url.Values{}
	urldata.Set("xmldata", string(xmldata))
	request, err := http.NewRequestWithContext(ctx, "POST", EndPoints[api.Mode], strings.NewReader(urldata.Encode()))
	if err != nil {
		log.Println(err)
		return res
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-MERCHANT-ID", req.MerchantID.(string))
	request.Header.Set("X-TERMINAL-ID", req.TerminalID.(string))
	if req.OOS != nil {
		if req.OOS.XID != nil {
			request.Header.Set("X-CORRELATION-ID", req.OOS.XID.(string))
		}
		if req.OOS.PosnetID != nil {
			request.Header.Set("X-POSNET-ID", req.OOS.PosnetID.(string))
		}
	}
	client := new(http.Client)
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return res
	}
	defer response.Body.Close()
	decoder := xml.NewDecoder(response.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	decoder.Decode(&res)
	return res
}
