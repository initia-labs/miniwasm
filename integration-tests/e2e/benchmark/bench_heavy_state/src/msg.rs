use cosmwasm_schema::{cw_serde, QueryResponses};

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    WriteMixed {
        shared_count: u64,
        local_count: u64,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(NonceResponse)]
    Nonce { address: String },
}

#[cw_serde]
pub struct NonceResponse {
    pub nonce: u64,
}
