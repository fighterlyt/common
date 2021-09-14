package ktexclient

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

// Client http://47.241.192.246:4999/web/#/page/edit/43/243
type Client struct {
	*http.Client
	crypto Crypto
	iv     []byte
	key    []byte
	hash   crypto.Hash
}

func NewClient(client *http.Client, key, iv []byte) (*Client, error) {
	if client == nil {
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		client = &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		}
	}

	aesCryptoCBC, err := NewAESCryptoCBC(key)
	if err != nil {
		return nil, err
	}

	return &Client{Client: client, iv: iv, key: key, crypto: aesCryptoCBC, hash: crypto.SHA256}, nil
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}

	var buf bytes.Buffer

	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}

	if err := b.Close(); err != nil {
		return nil, b, err
	}

	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func dumpReqBody(req *http.Request) (body []byte, err error) {
	save := req.Body

	if req.Body == nil {
		req.Body = nil
	} else {
		// req.Body在这个函数关闭了
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil, err
		}
	}

	var b bytes.Buffer

	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"

	if req.Body != nil {
		var dest io.Writer = &b
		if chunked {
			dest = httputil.NewChunkedWriter(dest)
		}

		_, err = io.Copy(dest, req.Body)
		if chunked {
			dest.(io.Closer).Close()
		}
	}

	if err != nil {
		return nil, err
	}
	req.Body = save
	return b.Bytes(), nil
}

func (c *Client) Do(req *http.Request) (resp *http.Response, err error) {
	str := NonceStr()
	unix := time.Now().Unix()
	unixStr := strconv.Itoa(int(unix))

	req.Header.Set("Nonce", str)
	req.Header.Set("Timestamp", unixStr)

	body, err := dumpReqBody(req)
	if err != nil {
		return nil, err
	}

	h := c.hash.New()

	if _, err = h.Write(body); err != nil {
		return nil, err
	}

	hashed := h.Sum(nil)

	verificationString, err := BuildVerificationString(unixStr, str, hashed)
	if err != nil {
		return nil, err
	}

	encryptWithIV, err := c.crypto.EncryptWithIV(verificationString, c.iv)
	if err != nil {
		return nil, err
	}

	sign := base64.StdEncoding.EncodeToString(encryptWithIV)
	req.Header.Set("Sign", sign)

	return c.Client.Do(req)
}
