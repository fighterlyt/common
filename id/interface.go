package id

import "github.com/bwmarrin/snowflake"

// Generator ID 生成器
type Generator interface {
	Generate() snowflake.ID
}
