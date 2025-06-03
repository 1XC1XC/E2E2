module E2E2/Client

go 1.24.3

require (
	E2E2/Cipher v0.0.0
	E2E2/Storage v0.0.0
)

require golang.org/x/crypto v0.27.0 // indirect

replace E2E2/Cipher => ../Cipher

replace E2E2/Storage => ../Storage
