package geoip

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

// Error defines the error in returning Geographical information
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Response defines the structure of Geographical information
type Response struct {
	IPAddress     string  `json:"IPAddress"`
	CityName      string  `json:"CityName"`
	StateCode     string  `json:"StateCode"`
	CountryCode   string  `json:"CountryCode"`
	ContinentCode string  `json:"ContinentCode"`
	TimeZone      string  `json:"TimeZone"`
	ZipCode       string  `json:"ZipCode"`
	Latitude      float64 `json:"Latitude"`
	Longitude     float64 `json:"Longitude"`
	Language      string  `json:"Language"`
	Error         Error   `json:"error"`
}

// getErrorResponse returns the error if something goes wrong
func getErrorResponse(errResponse string) *Response {
	err := Error{
		Code:    http.StatusBadRequest,
		Message: errResponse,
	}

	return &Response{
		Error: err,
	}
}

// getResponse Maps the record from Maxmind in appropriate format
func getResponse(ipAddress, language string, db *geoip2.Reader) *Response {
	ip := net.ParseIP(ipAddress)
	record, err := db.City(ip)

	if err != nil {
		return getErrorResponse("Could not get Geo information")
	}

	var stateCode string
	if len(record.Subdivisions) > 0 {
		stateCode = record.Subdivisions[0].Names[language]
	}

	return &Response{
		IPAddress:     ipAddress,
		CityName:      record.City.Names[language],
		StateCode:     stateCode,
		CountryCode:   record.Country.IsoCode,
		ContinentCode: record.Continent.Code,
		TimeZone:      record.Location.TimeZone,
		ZipCode:       record.Postal.Code,
		Latitude:      record.Location.Latitude,
		Longitude:     record.Location.Longitude,
		Language:      language,
	}
}

// getLanguage returns the language of the user from the header
func getLanguage(c *gin.Context) string {
	acceptLang := strings.TrimSpace(c.Request.Header.Get("ACCEPT-LANGUAGE"))
	geoIPSupported := []string{"fr", "de", "ja", "ru", "es", "pt-BR", "zh-CN", "en"}

	for _, lang := range geoIPSupported {
		if strings.Contains(acceptLang, lang) {
			return lang
		}
	}

	return "en"
}

// getClientIP returns the IP Address of the user from the headers
func getClientIP(c *gin.Context) (string, error) {
	xForwardedFor := strings.TrimSpace(c.Request.Header.Get("X-FORWARDED-FOR"))
	remoteAddr := strings.TrimSpace(c.Request.Header.Get("REMOTE-ADDR"))
	clientIP := strings.TrimSpace(c.Request.Header.Get("CLIENT-IP"))

	ipAddr := ""

	if xForwardedFor != `` {
		ipAddr = xForwardedFor
	} else if remoteAddr != `` {
		ipAddr = remoteAddr
	} else if clientIP != `` {
		ipAddr = clientIP
	}

	if ipAddr != `` {
		ip := net.ParseIP(ipAddr)
		if ip == nil || ip.IsLoopback() {
			return "", errors.New(" Invalid IP or Loopback IP address")
		}

		if ip = ip.To4(); ip == nil {
			return "", errors.New(" Could not get IPv4 address")
		}

		return ipAddr, nil
	}

	return "", errors.New(" Could not get client IP address")
}

// setContext sets the geographical information in Gin context
func setContext(c *gin.Context, db *geoip2.Reader) {
	start := time.Now()
	ipAddress, err := getClientIP(c)

	if err == nil {
		language := getLanguage(c)
		response := getResponse(ipAddress, language, db)
		c.Set("GeoResponse", response)
	} else {
		response := getErrorResponse(err.Error())
		c.Set("GeoResponse", response)
	}

	duration := time.Since(start)
	log.Println("Geo: Middleware duration", duration)
}

// getDB returns the database handle
func getDB(dbPath string) (*geoip2.Reader, *Response) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, getErrorResponse("Maxmind DB not found")
	}

	return db, nil
}

func getDBFromReader(reader io.Reader) (*geoip2.Reader, *Response) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, getErrorResponse(`read error`)
	}

	db, err := geoip2.FromBytes(data)
	if err != nil {
		return nil, getErrorResponse("Maxmind DB not found")
	}

	return db, nil
}

// Middleware sets the Geographical information
// about the user in the Gin context
func Middleware(dbPath string) gin.HandlerFunc {
	db, dbErr := getDB(dbPath)

	return func(c *gin.Context) {
		if dbErr == nil {
			setContext(c, db)
		}

		c.Next()
	}
}

func MiddleWareWithGeoReader(reader io.Reader) gin.HandlerFunc {
	db, dbErr := getDBFromReader(reader)

	return func(c *gin.Context) {
		if dbErr == nil {
			setContext(c, db)
		}
		c.Next()
	}
}

// Default returns the handler that sets the
// geographical information about the user
func Default(dbPath string) gin.HandlerFunc {
	return Middleware(dbPath)
}
