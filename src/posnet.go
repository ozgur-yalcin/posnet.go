package posnet

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

var EndPoints = map[string]string{
	"PROD":   "https://posnet.yapikredi.com.tr/PosnetWebService/XML",
	"PROD3D": "https://posnet.yapikredi.com.tr/3DSWebService/YKBPaymentService",

	"TEST":   "https://setmpos.ykb.com/PosnetWebService/XML",
	"TEST3D": "https://setmpos.ykb.com/3DSWebService/YKBPaymentService",
}

var CurrencyCode = map[string]string{
	"TRY": "TL",
	"YTL": "TL",
	"TRL": "TL",
	"TL":  "TL",
	"USD": "US",
	"US":  "US",
	"EUR": "EU",
	"EU":  "EU",
}

type API struct {
	Mode string
	Key  string
	Pid  string
}

type Request struct {
	XMLName     xml.Name     `xml:"posnetRequest,omitempty"`
	MerchantID  string       `xml:"mid,omitempty"`
	TerminalID  string       `xml:"tid,omitempty"`
	TranDate    string       `xml:"tranDateRequired,omitempty"`
	OOS         *OOS         `xml:"oosRequestData,omitempty"`
	OOSMerchant *OOSMerchant `xml:"oosResolveMerchantData,omitempty"`
	OOSTran     *OOSTran     `xml:"oosTranData,omitempty"`
	PreAuth     *PreAuth     `xml:"auth,omitempty"`
	Auth        *Auth        `xml:"sale,omitempty"`
	PostAuth    *PostAuth    `xml:"capt,omitempty"`
	Refund      *Refund      `xml:"return,omitempty"`
	Cancel      *Cancel      `xml:"reverse,omitempty"`
}

type Form struct {
	MerchantID string `form:"mid,omitempty"`
	PosnetID   string `form:"posnetID,omitempty"`
	Data1      string `form:"posnetData,omitempty"`
	Data2      string `form:"posnetData2,omitempty"`
	Sign       string `form:"digest,omitempty"`
	VftCode    string `form:"vftCode,omitempty"`
	ReturnUrl  string `form:"merchantReturnURL,omitempty"`
	Url        string `form:"url,omitempty"`
	NewWindow  string `form:"openANewWindow,omitempty"`
	Lang       string `form:"lang,omitempty"`
}

type OOS struct {
	PosnetID    string `xml:"posnetid,omitempty"`
	XID         string `xml:"XID,omitempty"`
	TranType    string `xml:"tranType,omitempty"`
	CardHolder  string `xml:"cardHolderName,omitempty"`
	CardNumber  string `xml:"ccno,omitempty"`
	CardExpiry  string `xml:"expDate,omitempty"`
	CardCode    string `xml:"cvc,omitempty"`
	Amount      string `xml:"amount,omitempty"`
	Currency    string `xml:"currencyCode,omitempty"`
	Installment string `xml:"installment,omitempty"`
}

type OOSMerchant struct {
	BankData     string `xml:"bankData,omitempty"`
	MerchantData string `xml:"merchantData,omitempty"`
	SIGN         string `xml:"sign,omitempty"`
	MAC          string `xml:"mac,omitempty"`
}

type OOSTran struct {
	BankData string `xml:"bankData,omitempty"`
	WpAmount string `xml:"wpAmount,omitempty"`
	MAC      string `xml:"mac,omitempty"`
}

type Card struct {
	InquiryValue string `xml:"inquiryValue,omitempty"`
	CardNoFirst  string `xml:"cardNoFirst,omitempty"`
	CardNoLast   string `xml:"cardNoLast,omitempty"`
}

type PreAuth struct {
	Card        *Card  `xml:"cardInfo,omitempty"`
	CardNumber  string `xml:"ccno,omitempty"`
	CardExpiry  string `xml:"expDate,omitempty"`
	CardCode    string `xml:"cvc,omitempty"`
	Amount      string `xml:"amount,omitempty"`
	Currency    string `xml:"currencyCode,omitempty"`
	Installment string `xml:"installment,omitempty"`
	OrderId     string `xml:"orderID,omitempty"`
}

type Auth struct {
	Card        *Card  `xml:"cardInfo,omitempty"`
	CardNumber  string `xml:"ccno,omitempty"`
	CardExpiry  string `xml:"expDate,omitempty"`
	CardCode    string `xml:"cvc,omitempty"`
	Amount      string `xml:"amount,omitempty"`
	Currency    string `xml:"currencyCode,omitempty"`
	Installment string `xml:"installment,omitempty"`
	OrderId     string `xml:"orderID,omitempty"`
	Mailorder   string `xml:"mailorderflag,omitempty"`
}

type PostAuth struct {
	Amount      string `xml:"amount,omitempty"`
	Currency    string `xml:"currencyCode,omitempty"`
	Installment string `xml:"installment,omitempty"`
	HostLogKey  string `xml:"hostlogkey,omitempty"`
}

type Refund struct {
	Amount      string `xml:"amount,omitempty"`
	Currency    string `xml:"currencyCode,omitempty"`
	Transaction string `xml:"transaction,omitempty"`
	HostLogKey  string `xml:"hostlogkey,omitempty"`
}

type Cancel struct {
	Transaction string `xml:"transaction,omitempty"`
	HostLogKey  string `xml:"hostlogkey,omitempty"`
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

func B64(data string) (hash string) {
	hash = base64.StdEncoding.EncodeToString([]byte(data))
	return hash
}

func D64(data string) []byte {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Println(err)
		return nil
	}
	return b
}

func XID(n int) string {
	const alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var bytes = make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func Amount(amount string) string {
	return strings.ReplaceAll(amount, ".", "")
}

func Installment(installment string) string {
	return fmt.Sprintf("%02v", installment)
}

func Currency(currency string) string {
	return CurrencyCode[currency]
}

func Expiry(month, year string) string {
	return fmt.Sprintf("%02v", year) + fmt.Sprintf("%02v", month)
}

func Api(merchant, terminal string) (*API, *Request) {
	api := new(API)
	request := new(Request)
	request.MerchantID = merchant
	request.TerminalID = terminal
	return api, request
}

func (api *API) Transaction(ctx context.Context, req *Request) (res Response, err error) {
	xmldata, err := xml.Marshal(req)
	if err != nil {
		return res, err
	}
	postdata := url.Values{}
	postdata.Set("xmldata", string(xmldata))
	request, err := http.NewRequestWithContext(ctx, "POST", EndPoints[api.Mode], strings.NewReader(postdata.Encode()))
	if err != nil {
		return res, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-MERCHANT-ID", req.MerchantID)
	request.Header.Set("X-TERMINAL-ID", req.TerminalID)
	if req.OOS != nil {
		if req.OOS.XID != "" {
			request.Header.Set("X-CORRELATION-ID", req.OOS.XID)
		}
		if req.OOS.PosnetID != "" {
			request.Header.Set("X-POSNET-ID", req.OOS.PosnetID)
		}
	}
	client := new(http.Client)
	response, err := client.Do(request)
	if err != nil {
		return res, err
	}
	defer response.Body.Close()
	decoder := xml.NewDecoder(response.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&res); err != nil {
		return res, err
	}
	switch res.Approved {
	case "1":
		return res, nil
	default:
		return res, errors.New(res.ErrorText)
	}
}

func (api *API) Transaction3D(ctx context.Context, req *Form) (res string, err error) {
	postdata, err := QueryString(req)
	if err != nil {
		return res, err
	}
	html := []string{}
	html = append(html, `<!DOCTYPE html>`)
	html = append(html, `<html>`)
	html = append(html, `<head>`)
	html = append(html, `<script type="text/javascript">function submitonload() {document.payment.submit();document.getElementById('button').remove();document.getElementById('body').insertAdjacentHTML("beforeend", "Lütfen bekleyiniz...");}</script>`)
	html = append(html, `</head>`)
	html = append(html, `<body onload="javascript:submitonload();" id="body" style="text-align:center;margin:10px;font-family:Arial;font-weight:bold;">`)
	html = append(html, `<form action="`+EndPoints[api.Mode+"3D"]+`" method="post" name="payment">`)
	for k := range postdata {
		html = append(html, `<input type="hidden" name="`+k+`" value="`+postdata.Get(k)+`">`)
	}
	html = append(html, `<input type="submit" name="Submit" value="Gönder" id="button">`)
	html = append(html, `</form>`)
	html = append(html, `</body>`)
	html = append(html, `</html>`)
	res = B64(strings.Join(html, "\n"))
	return res, err
}
