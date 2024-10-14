package base

import _ "embed"

//go:embed base-client.ts
var BaseClientSource string

//go:embed base-client.go.txt
var BaseGoClient string
