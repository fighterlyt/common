package helpers

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const ClaimsKey = "JWT_PAYLOAD"

var (
	ErrNoToken = errors.New(`没有token`)

	ErrUnknownType = errors.New("unknown type")
)

func CtxError(ctx *gin.Context, err error) {
	_ = ctx.Error(errors.Wrapf(err, "url:%s", ctx.Request.URL.String()))
}

// GetUserID 从session登录的用户信息
func GetUserID(ctx *gin.Context) (int64, error) {
	token := ctx.GetHeader("token")

	if token == `` {
		return 0, ErrUnknownType
	}

	j := NewJWT()

	clamis, err := j.ParseToken(token)
	if err != nil {
		return 0, errors.Wrap(err, "未获取到token")
	}

	return clamis.UserID, nil
}

func Getclaims(ctx context.Context) (*JwtCustomClaims, error) {
	claims := ctx.Value(ClaimsKey)
	if claims != nil {
		if c, ok := claims.(*JwtCustomClaims); ok {
			return c, nil
		}
	}

	return nil, errors.New("未获取到claims")
}

type JWT struct {
	SigningKey []byte
}

// NewJWT 初始化
func NewJWT() *JWT {
	// 这里要注意，SigningKey 这个值，需要自定义
	return &JWT{SigningKey: []byte("acw_login_jwt")}
}

type JwtCustomClaims struct {
	UserID     int64  // 用户id
	Address    string // 地址
	QrCode     string // 二维码链接
	Mobile     string // 手机号
	Role       string // 角色
	CreateTime int64
	jwt.StandardClaims
}

// CreateToken 创建 token
func (j *JWT) CreateToken(claims JwtCustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// 定义错误信息
var (
	TokenExpired        = ErrTokenExpired
	TokenNotValidYet    = ErrTokenNotValidYet
	TokenMalformed      = ErrTokenMalformed
	TokenInvalid        = ErrTokenInvalid
	ErrTokenExpired     = errors.New("Token 已经过期")
	ErrTokenNotValidYet = errors.New("Token 未激活")
	ErrTokenMalformed   = errors.New("Token 错误")
	ErrTokenInvalid     = errors.New("Token 无效")
)

func convertErr(err error) error {
	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return TokenMalformed
		} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
			return TokenExpired
		} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
			return TokenNotValidYet
		} else {
			return TokenInvalid
		}
	}

	return nil
}

func (j *JWT) ParseToken(tokenString string) (*JwtCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if convertedErr := convertErr(err); convertedErr != nil {
			return nil, convertedErr
		}
	}

	if token == nil {
		return nil, TokenInvalid
	}
	// 解析到Claims 构造中
	if c, ok := token.Claims.(*JwtCustomClaims); ok && token.Valid {
		return c, nil
	}

	return nil, TokenInvalid
}

// RefreshToken 更新 token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}

	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}

	if c, ok := token.Claims.(*JwtCustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now

		c.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()

		return j.CreateToken(*c)
	}

	return "", TokenInvalid
}
