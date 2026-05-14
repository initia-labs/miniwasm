use cosmwasm_schema::cw_serde;
use cosmwasm_std::Uint128;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    IbcTransfer {
        source_channel: String,
        receiver: String,
        denom: String,
        amount: Uint128,
        /// IBC timeout timestamp in nanoseconds.
        timeout_timestamp: u64,
        memo: Option<String>,
    },
}

#[cw_serde]
pub struct QueryMsg {}
