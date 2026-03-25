use cosmwasm_std::Addr;
use cw_storage_plus::Map;

pub const NONCES: Map<&Addr, u64> = Map::new("nonces");
pub const SHARED_STATE: Map<&str, u64> = Map::new("shared");
pub const LOCAL_STATE: Map<(&Addr, &str), u64> = Map::new("local");
