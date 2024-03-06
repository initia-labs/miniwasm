package types

const (
	// ModuleName defines the module name
	ModuleName = "tokenfactory"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_tokenfactory"
)

var (
	CreatorDenomsPrefix  = []byte{0x11}
	DenomAuthorityPrefix = []byte{0x12}
	DenomHookAddrPrefix  = []byte{0x13}
	ParamsKeyPrefix      = []byte{0x14}
)
